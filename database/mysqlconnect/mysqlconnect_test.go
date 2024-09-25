package mysqlconnect

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"
	"time"

	mocks "github.com/JhonX2011/GOWebApplication/test/mocks"
	"github.com/stretchr/testify/require"
)

var mockDriver = mocks.MysqlDriverMock{}

func init() {
	sql.Register("mysql", &mockDriver)
}

func TestOpen_ConfigPreconditions(t *testing.T) {
	testCases := []struct {
		name       string
		config     Config
		errMessage string
	}{
		{
			name: "missing DSN, Cluster and HACluster",
			config: Config{
				DSN:       "",
				Cluster:   "",
				HACluster: "",
			},
			errMessage: "invalid MySQL config: DSN, Cluster and HACluster are empty",
		},
		{
			name: "DSN present and Cluster present",
			config: Config{
				DSN:     "foo",
				Cluster: "bar",
			},
			errMessage: "invalid MySQL config: DSN is mutually exclusive with Cluster and HACluster",
		},
		{
			name: "DSN present and HACluster present",
			config: Config{
				DSN:       "foo",
				HACluster: "bar",
			},
			errMessage: "invalid MySQL config: DSN is mutually exclusive with Cluster and HACluster",
		},
		{
			name: "Cluster present and HACluster present",
			config: Config{
				Cluster:   "foo",
				HACluster: "bar",
			},
			errMessage: "invalid MySQL config: Cluster is mutually exclusive with HACluster",
		},
		{
			name: "DSN and schema are set",
			config: Config{
				DSN:    "foo",
				Schema: "schema",
			},
			errMessage: "invalid MySQL config: DSN is mutually exclusive with Schema since the schema is already defined in the DSN",
		},
		{
			name: "DSN and schema are not set",
			config: Config{
				Cluster: "foo",
			},
			errMessage: "invalid MySQL config: when DSN is empty the Schema must be defined",
		},
		{
			name: "there is no connection defined",
			config: Config{
				DSN: "foo",
			},
			errMessage: "invalid MySQL config: no connections defined",
		},
		{
			name: "duplicate connection names",
			config: Config{
				DSN: "foo",
				Connections: []Connection{
					{
						Name: "foo",
					},
					{
						Name: "foo",
					},
				},
			},
			errMessage: "invalid MySQL config: duplicated connection name \"foo\"",
		},
		{
			name: "connection is not master but it has write permissions",
			config: Config{
				Cluster: "DB_MYSQL_DESAENV08_FOO",
				Schema:  "bar",
				Connections: []Connection{
					{
						Name:       "foo",
						IsMaster:   false,
						IsReadOnly: false,
					},
				},
			},
			errMessage: "invalid MySQL config: cannot write to a replica: connection \"foo\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Open(tc.config)
			require.EqualError(t, err, tc.errMessage)
		})
	}
}

func TestDuration_MarshalJSON(t *testing.T) {
	duration := Duration(100 * time.Millisecond)
	data, err := duration.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, "\"100ms\"", string(data))
}

func TestDuration_UnmarshalJSON_InvalidDuration(t *testing.T) {
	var duration Duration
	err := duration.UnmarshalJSON([]byte("\"foo\""))
	require.EqualError(t, err, "invalid duration: \"foo\"")
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	var duration Duration
	err := duration.UnmarshalJSON([]byte("\"100ms\""))
	require.NoError(t, err)
	require.Equal(t, Duration(100*time.Millisecond), duration)
}

func TestOpen(t *testing.T) {
	testCases := []struct {
		name          string
		config        Config
		expectedDSN   string
		setEnvVarFunc func(t *testing.T)
	}{
		{
			name: "use DSN",
			config: Config{
				DSN: "root:password@tcp(localhost:3306)/foo?timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
				Connections: []Connection{
					{
						Name: "foo",
					},
				},
			},
			expectedDSN:   "root:password@tcp(localhost:3306)/foo?timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
			setEnvVarFunc: func(t *testing.T) {},
		},
		{
			name: "use DSN, parameters are ignored",
			config: Config{
				DSN: "root:password@tcp(localhost:3306)/foo",
				Connections: []Connection{
					{
						Name:       "foo",
						Parameters: "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
					},
				},
			},
			expectedDSN:   "root:password@tcp(localhost:3306)/foo",
			setEnvVarFunc: func(t *testing.T) {},
		},
		{
			name: "fury mysql master with read/write permissions",
			config: Config{
				Cluster: "desaenv08",
				Schema:  "bar",
				Connections: []Connection{
					{
						Name:       "foo",
						IsMaster:   true,
						Parameters: "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
					},
				},
			},
			expectedDSN: "bar_WPROD:password@tcp(localhost:3306)/bar?timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
			setEnvVarFunc: func(t *testing.T) {
				t.Setenv("DB_MYSQL_DESAENV08_BAR_BAR_ENDPOINT", "localhost:3306")
				t.Setenv("DB_MYSQL_DESAENV08_BAR_BAR_WPROD", "password")
			},
		},
		{
			name: "fury mysql master with read only permissions",
			config: Config{
				Cluster: "desaenv08",
				Schema:  "bar",
				Connections: []Connection{
					{
						Name:       "foo",
						IsMaster:   false,
						IsReadOnly: true,
						Parameters: "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
					},
				},
			},
			expectedDSN: "bar_RPROD:password@tcp(localhost:3306)/bar?timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
			setEnvVarFunc: func(t *testing.T) {
				t.Setenv("DB_MYSQL_DESAENV08_BAR_BAR_LOCAL_REPLICA_ENDPOINT", "localhost:3306")
				t.Setenv("DB_MYSQL_DESAENV08_BAR_BAR_RPROD", "password")
			},
		},
		{
			name: "fury mysql ha master with read/write permissions",
			config: Config{
				HACluster: "desaenv08",
				Schema:    "bar",
				Connections: []Connection{
					{
						Name:       "foo",
						IsMaster:   true,
						IsReadOnly: false,
						Parameters: "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
					},
				},
			},
			setEnvVarFunc: func(t *testing.T) {
				t.Setenv("DB_HA_MYSQL_DESAENV08_BAR_BAR_WR_ENDPOINT", "localhost:3306")
				t.Setenv("DB_HA_MYSQL_DESAENV08_BAR_BAR_WPROD", "password")
			},
			expectedDSN: "bar_WPROD:password@tcp(localhost:3306)/bar?timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
		},
		{
			name: "fury mysql ha master with read only permissions",
			config: Config{
				HACluster: "desaenv08",
				Schema:    "bar",
				Connections: []Connection{
					{
						Name:       "foo",
						IsMaster:   false,
						IsReadOnly: true,
						Parameters: "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
					},
				},
			},
			setEnvVarFunc: func(t *testing.T) {
				t.Setenv("DB_HA_MYSQL_DESAENV08_BAR_BAR_RO_ENDPOINT", "localhost:3306")
				t.Setenv("DB_HA_MYSQL_DESAENV08_BAR_BAR_RPROD", "password")
			},
			expectedDSN: "bar_RPROD:password@tcp(localhost:3306)/bar?timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setEnvVarFunc(t)

			mockDriver.OpenFunc = func(name string) (driver.Conn, error) {
				if name != tc.expectedDSN {
					return nil, errors.New("invalid call to Open")
				}

				return &mocks.DriverConnMock{}, nil
			}

			connections, err := Open(tc.config)
			require.NoError(t, err)

			db, err := connections.Get("foo")
			require.NoError(t, err)

			err = db.Ping()
			require.NoError(t, err)
		})
	}
}

func TestConnections_List(t *testing.T) {
	config := Config{
		DSN: "root:password@tcp(localhost:3306)/foo?timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
		Connections: []Connection{
			{
				Name: "foo",
			},
			{
				Name: "bar",
			},
			{
				Name: "baz",
			},
		},
	}

	mockDriver.OpenFunc = func(name string) (driver.Conn, error) {
		return &mocks.DriverConnMock{}, nil
	}

	connections, err := Open(config)
	require.NoError(t, err)

	dbs := connections.List()
	require.Len(t, dbs, 3)

	var count int
	for _, db := range dbs {
		count++
		err = db.Ping()
		require.NoError(t, err)
	}

	require.NoError(t, err)
	require.Equal(t, 3, count)
}

func TestConnections_Close(t *testing.T) {
	config := Config{
		DSN: "root:password@tcp(localhost:3306)/foo",
		Connections: []Connection{
			{
				Name: "foo",
			},
			{
				Name: "bar",
			},
			{
				Name: "baz",
			},
		},
	}

	mockDriver.OpenFunc = func(name string) (driver.Conn, error) {
		return &mocks.DriverConnMock{
			CloseFunc: func() error {
				return nil
			},
		}, nil
	}

	connections, err := Open(config)
	require.NoError(t, err)

	db, err := connections.Get("foo")
	require.NoError(t, err)
	require.Nil(t, db.Ping())

	db, err = connections.Get("bar")
	require.NoError(t, err)
	require.Nil(t, db.Ping())

	db, err = connections.Get("baz")
	require.NoError(t, err)
	require.Nil(t, db.Ping())

	err = connections.Close()
	require.NoError(t, err)

	db, err = connections.Get("foo")
	require.NoError(t, err)
	require.EqualError(t, db.Ping(), "sql: database is closed")

	db, err = connections.Get("bar")
	require.NoError(t, err)
	require.EqualError(t, db.Ping(), "sql: database is closed")

	db, err = connections.Get("baz")
	require.NoError(t, err)
	require.EqualError(t, db.Ping(), "sql: database is closed")
}

func TestConnections_CloseReturnError(t *testing.T) {
	// If the DSN contains the string "close_with_error",
	// it instructs the mock to return a connector that closes with error.
	config := Config{
		DSN: "root:password@tcp(localhost:3306)/close_with_error",
		Connections: []Connection{
			{
				Name: "foo",
			},
			{
				Name: "bar",
			},
			{
				Name: "baz",
			},
		},
	}

	mockDriver.OpenFunc = func(name string) (driver.Conn, error) {
		return &mocks.DriverConnMock{
			CloseFunc: func() error {
				return nil
			},
		}, nil
	}

	connections, err := Open(config)
	require.NoError(t, err)

	db, err := connections.Get("foo")
	require.NoError(t, err)
	require.Nil(t, db.Ping())

	db, err = connections.Get("bar")
	require.NoError(t, err)
	require.Nil(t, db.Ping())

	db, err = connections.Get("baz")
	require.NoError(t, err)
	require.Nil(t, db.Ping())

	closeErr := connections.Close()
	require.EqualError(t, closeErr, "failed to close connections: bar: driver: bad connection, baz: driver: bad connection, foo: driver: bad connection")

	db, err = connections.Get("foo")
	require.NoError(t, err)
	require.EqualError(t, db.Ping(), "sql: database is closed")

	db, err = connections.Get("bar")
	require.NoError(t, err)
	require.EqualError(t, db.Ping(), "sql: database is closed")

	db, err = connections.Get("baz")
	require.NoError(t, err)
	require.EqualError(t, db.Ping(), "sql: database is closed")

}

// TestConfigAsJSON tests that the config can be unmarshalled from JSON.
// The JSON per se is not valid. It is just used to test the unmarshalling with every possible field.
func TestConfigAsJSON(t *testing.T) {
	configJSON := `{
  "dsn": "root:password@tcp(localhost:3306)/foo?timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
  "cluster": "cluster_foo",
  "ha_cluster": "ha_cluster_foo",
  "schema": "bar",
  "connections": [
    {
      "name": "default",
      "is_master": true,
      "is_read_only": true,
      "parameters": "charset=utf8mb4&collation=utf8mb4_unicode_ci",
      "connection_pool": {
        "conn_max_lifetime": "10m",
        "max_idle_connections": 100,
        "max_open_connections": 101,
        "conn_max_idle_time": "11m"
      }
    },
    {
      "name": "default2",
      "connection_pool": {
        "conn_max_lifetime": "12m",
        "max_idle_connections": 102,
        "max_open_connections": 103,
        "conn_max_idle_time": "13m"
      }
    }
  ]
}`
	config := Config{}
	err := json.Unmarshal([]byte(configJSON), &config)
	require.NoError(t, err)

	require.Equal(t, "root:password@tcp(localhost:3306)/foo?timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true", config.DSN)
	require.Equal(t, "cluster_foo", config.Cluster)
	require.Equal(t, "ha_cluster_foo", config.HACluster)
	require.Equal(t, "bar", config.Schema)
	require.Len(t, config.Connections, 2)
	require.Equal(t, "default", config.Connections[0].Name)
	require.Equal(t, true, config.Connections[0].IsMaster)
	require.Equal(t, true, config.Connections[0].IsReadOnly)
	require.Equal(t, "charset=utf8mb4&collation=utf8mb4_unicode_ci", config.Connections[0].Parameters)
	require.Equal(t, Duration(10*time.Minute), *config.Connections[0].ConnectionPool.ConnMaxLifetime)
	require.Equal(t, 100, *config.Connections[0].ConnectionPool.MaxIdleConnections)
	require.Equal(t, 101, *config.Connections[0].ConnectionPool.MaxOpenConnections)
	require.Equal(t, Duration(11*time.Minute), *config.Connections[0].ConnectionPool.ConnMaxIdleTime)

	require.Equal(t, "default2", config.Connections[1].Name)
	require.Equal(t, false, config.Connections[1].IsMaster)
	require.Equal(t, false, config.Connections[1].IsReadOnly)
	require.Equal(t, "", config.Connections[1].Parameters)
	require.Equal(t, Duration(12*time.Minute), *config.Connections[1].ConnectionPool.ConnMaxLifetime)
	require.Equal(t, 102, *config.Connections[1].ConnectionPool.MaxIdleConnections)
	require.Equal(t, 103, *config.Connections[1].ConnectionPool.MaxOpenConnections)
	require.Equal(t, Duration(13*time.Minute), *config.Connections[1].ConnectionPool.ConnMaxIdleTime)
}

func TestConfigAsJSON_DSN_Readme(t *testing.T) {
	configJSON := `{
  "dsn": "root:password@tcp(localhost:3306)/my_schema?timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
  "connections": [
    {
      "name": "default"
    }
  ]
}`
	config := Config{}
	err := json.Unmarshal([]byte(configJSON), &config)
	require.NoError(t, err)

	require.Equal(t, "root:password@tcp(localhost:3306)/my_schema?timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true", config.DSN)
	require.Equal(t, "default", config.Connections[0].Name)

	connections, err := Open(config)
	require.NoError(t, err)
	_, err = connections.Get("default")
	require.NoError(t, err)
}

func TestConfigAsJSON_FuryMySQL_Readme(t *testing.T) {
	t.Setenv("DB_MYSQL_DESAENV08_BAR_BAR_ENDPOINT", "localhost:3306")
	t.Setenv("DB_MYSQL_DESAENV08_BAR_BAR_RPROD", "password")

	configJSON := `{
  "cluster": "desaenv08",
  "schema": "my_schema",
  "connections": [
    {
      "name": "master_rw",
      "is_master": true,
      "is_read_only": false,
      "parameters": "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
      "connection_pool": {
        "conn_max_lifetime": "10m",
        "max_idle_connections": 100,
        "max_open_connections": 100,
        "conn_max_idle_time": "1m"
      }
    },
    {
      "name": "master_ro",
      "is_master": true,
      "is_read_only": true,
      "parameters": "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
      "connection_pool": {
        "conn_max_lifetime": "10m",
        "max_idle_connections": 100,
        "max_open_connections": 100,
        "conn_max_idle_time": "1m"
      }
    },
    {
      "name": "replica_ro",
      "is_master": false,
      "is_read_only": true,
      "parameters": "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
      "connection_pool": {
        "conn_max_lifetime": "10m",
        "max_idle_connections": 100,
        "max_open_connections": 100,
        "conn_max_idle_time": "1m"
      }
    }
  ]
}`
	var config Config
	err := json.Unmarshal([]byte(configJSON), &config)
	require.NoError(t, err)

	require.Equal(t, "desaenv08", config.Cluster)
	require.Equal(t, "my_schema", config.Schema)
	require.Len(t, config.Connections, 3)
	require.Equal(t, "master_rw", config.Connections[0].Name)
	require.Equal(t, true, config.Connections[0].IsMaster)
	require.Equal(t, false, config.Connections[0].IsReadOnly)
	require.Equal(t, "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true", config.Connections[0].Parameters)
	require.Equal(t, Duration(10*time.Minute), *config.Connections[0].ConnectionPool.ConnMaxLifetime)
	require.Equal(t, 100, *config.Connections[0].ConnectionPool.MaxIdleConnections)
	require.Equal(t, 100, *config.Connections[0].ConnectionPool.MaxOpenConnections)
	require.Equal(t, Duration(1*time.Minute), *config.Connections[0].ConnectionPool.ConnMaxIdleTime)

	require.Equal(t, "master_ro", config.Connections[1].Name)
	require.Equal(t, true, config.Connections[1].IsMaster)
	require.Equal(t, true, config.Connections[1].IsReadOnly)
	require.Equal(t, "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true", config.Connections[1].Parameters)
	require.Equal(t, Duration(10*time.Minute), *config.Connections[1].ConnectionPool.ConnMaxLifetime)
	require.Equal(t, 100, *config.Connections[1].ConnectionPool.MaxIdleConnections)
	require.Equal(t, 100, *config.Connections[1].ConnectionPool.MaxOpenConnections)
	require.Equal(t, Duration(1*time.Minute), *config.Connections[1].ConnectionPool.ConnMaxIdleTime)

	require.Equal(t, "replica_ro", config.Connections[2].Name)
	require.Equal(t, false, config.Connections[2].IsMaster)
	require.Equal(t, true, config.Connections[2].IsReadOnly)
	require.Equal(t, "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true", config.Connections[2].Parameters)
	require.Equal(t, Duration(10*time.Minute), *config.Connections[2].ConnectionPool.ConnMaxLifetime)
	require.Equal(t, 100, *config.Connections[2].ConnectionPool.MaxIdleConnections)
	require.Equal(t, 100, *config.Connections[2].ConnectionPool.MaxOpenConnections)
	require.Equal(t, Duration(1*time.Minute), *config.Connections[2].ConnectionPool.ConnMaxIdleTime)

	connections, err := Open(config)
	require.NoError(t, err)

	_, err = connections.Get("master_rw")
	require.NoError(t, err)
	_, err = connections.Get("master_ro")
	require.NoError(t, err)
	_, err = connections.Get("replica_ro")
	require.NoError(t, err)
}

func TestConfigAsJSON_DefaultConnectionPool(t *testing.T) {
	t.Setenv("DB_MYSQL_DESAENV08_FOO_BAR_BAR_ENDPOINT", "localhost:3306")
	t.Setenv("DB_MYSQL_DESAENV08_FOO_BAR_BAR_RPROD", "password")

	configJSON := `{
  "cluster": "desaenv08_foo",
  "schema": "my_schema",
  "connections": [
    {
      "name": "0",
      "is_master": true,
      "is_read_only": false,
      "parameters": "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
      "connection_pool": {
        "conn_max_lifetime": "10m"
      }
    },
    {
      "name": "1",
      "is_master": true,
      "is_read_only": true,
      "parameters": "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
      "connection_pool": {
        "max_idle_connections": 100
      }
    },
    {
      "name": "2",
      "is_master": false,
      "is_read_only": true,
      "parameters": "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
      "connection_pool": {
        "max_open_connections": 100
      }
    },
	{
      "name": "3",
      "is_master": false,
      "is_read_only": true,
      "parameters": "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
      "connection_pool": {
        "conn_max_idle_time": "1m"
      }
    },
	{
      "name": "4",
      "is_master": false,
      "is_read_only": true,
      "parameters": "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true",
      "connection_pool": {}
    },
	{
      "name": "5",
      "is_master": false,
      "is_read_only": true,
      "parameters": "timeout=100ms&readTimeout=100ms&writeTimeout=100ms&parseTime=true"
    }
  ]
}`
	var config Config
	err := json.Unmarshal([]byte(configJSON), &config)
	require.NoError(t, err)

	require.Equal(t, Duration(10*time.Minute), *config.Connections[0].ConnectionPool.ConnMaxLifetime)
	require.Nil(t, config.Connections[0].ConnectionPool.MaxIdleConnections)
	require.Nil(t, config.Connections[0].ConnectionPool.MaxOpenConnections)
	require.Nil(t, config.Connections[0].ConnectionPool.ConnMaxIdleTime)

	require.Nil(t, config.Connections[1].ConnectionPool.ConnMaxLifetime)
	require.Equal(t, 100, *config.Connections[1].ConnectionPool.MaxIdleConnections)
	require.Nil(t, config.Connections[1].ConnectionPool.MaxOpenConnections)
	require.Nil(t, config.Connections[1].ConnectionPool.ConnMaxIdleTime)

	require.Nil(t, config.Connections[2].ConnectionPool.ConnMaxLifetime)
	require.Nil(t, config.Connections[2].ConnectionPool.MaxIdleConnections)
	require.Equal(t, 100, *config.Connections[2].ConnectionPool.MaxOpenConnections)
	require.Nil(t, config.Connections[2].ConnectionPool.ConnMaxIdleTime)

	require.Nil(t, config.Connections[3].ConnectionPool.ConnMaxLifetime)
	require.Nil(t, config.Connections[3].ConnectionPool.MaxIdleConnections)
	require.Nil(t, config.Connections[3].ConnectionPool.MaxOpenConnections)
	require.Equal(t, Duration(1*time.Minute), *config.Connections[3].ConnectionPool.ConnMaxIdleTime)

	require.Nil(t, config.Connections[4].ConnectionPool.ConnMaxLifetime)
	require.Nil(t, config.Connections[4].ConnectionPool.MaxIdleConnections)
	require.Nil(t, config.Connections[4].ConnectionPool.MaxOpenConnections)
	require.Nil(t, config.Connections[4].ConnectionPool.ConnMaxIdleTime)

	require.Nil(t, config.Connections[5].ConnectionPool.ConnMaxLifetime)
	require.Nil(t, config.Connections[5].ConnectionPool.MaxIdleConnections)
	require.Nil(t, config.Connections[5].ConnectionPool.MaxOpenConnections)
	require.Nil(t, config.Connections[5].ConnectionPool.ConnMaxIdleTime)

	connections, err := Open(config)
	require.NoError(t, err)

	_, err = connections.Get("1")
	require.NoError(t, err)
	_, err = connections.Get("2")
	require.NoError(t, err)
	_, err = connections.Get("3")
	require.NoError(t, err)
	_, err = connections.Get("4")
	require.NoError(t, err)
}
