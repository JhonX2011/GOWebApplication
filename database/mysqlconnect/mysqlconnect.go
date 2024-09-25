package mysqlconnect

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/exp/maps"
)

// Config is the configuration needed for opening one or more connections to a MySQL database.
// The MySQL database can be a single instance, a MySQL cluster or a HA MySQL cluster.
// The configuration can be provided in two different ways:
// 1. DSN: The Data Source Name for connecting to the MySQL database.
// 2. Cluster or HACluster: The name of the cluster for a MySQL database running or HA MySQL database running in .
// The configuration is mutually exclusive, meaning that only one of them can be provided.
type Config struct {
	// DSN is the Data Source Name for connecting to the MySQL database.
	// It is usually used when the application is running locally.
	// It has the following format: [username[:password]@][protocol[(address)]]/schema[?param1=value1&...&paramN=valueN]
	// It is mutually exclusive with both Cluster and HACluster.
	DSN string `json:"dsn"`
	// Cluster is the name of the cluster for a MySQL database running in .
	// It is mutually exclusive with both DSN and HACluster.
	Cluster string `json:"cluster"`
	// HACluster is the name of the cluster for a HA MySQL database running in .
	// It is mutually exclusive with both DSN and Cluster.
	HACluster string `json:"ha_cluster"`
	// Schema is the name of the schema to connect to.
	// It is required when using either a Cluster or a HACluster.
	// It must be empty when using DSN since the schema is part of it.
	Schema string `json:"schema"`
	// Connections defines all the connections that will be created upon calling Open.
	// When using DSN, is_master and is_read_only are ignored.
	// For example if you want to create a connection to the master with read-write permissions and a connection to
	// the replica with read-only permissions you would define two connections each one with a different name.
	// For example:
	// 	{
	// 		"cluster": "my_cluster",
	// 		"schema": "my_schema",
	// 		"connections": [
	// 			{
	// 				"name": "master",
	// 				"is_master": true,
	// 				"is_read_only": false,
	// 				"parameters":"parseTime=true&readTimeout=100ms&timeout=100ms&writeTimeout=100ms",
	// 				"connection_pool": {
	// 					"conn_max_lifetime": "10m",
	// 					"max_idle_connections": 100,
	// 					"max_open_connections": 100,
	// 					"conn_max_idle_time": "1m"
	// 				}
	// 			},
	// 			{
	// 				"name": "replica",
	// 				"is_master": false,
	// 				"is_read_only": true,
	// 				"parameters":"parseTime=true&readTimeout=100ms&timeout=100ms&writeTimeout=100ms",
	// 				"connection_pool": {
	// 					"conn_max_lifetime": "10m",
	// 					"max_idle_connections": 100,
	// 					"max_open_connections": 100,
	// 					"conn_max_idle_time": "1m"
	// 				}
	// 			}
	// 		]
	// 	}
	// If you don't define any connection the Open function will return an error.
	// If you defined more than one connection with the same name the Open function will return an error.
	// If IsMaster is false and IsReadOnly is false the Open function will return an error since
	// it would make no sense to create a connection to a replica with read-write permissions.
	Connections []Connection `json:"connections"`
}

// Connection defines a connection to a MySQL database.
type Connection struct {
	// Name is the name of the connection. It must be unique among all the connections.
	Name string `json:"name"`
	// IsMaster indicates whether the connection is to the master.
	// It is ignored when using DSN.
	IsMaster bool `json:"is_master"`
	// IsReadOnly indicates whether the connection is read-only.
	// It is ignored when using DSN.
	IsReadOnly bool `json:"is_read_only"`
	// Parameters are the connection parameters in the form of param1=value1&...&paramN=valueN.
	// For example: parseTime=true&readTimeout=100ms&timeout=100ms&writeTimeout=100ms
	// It is optional and ignored when using DSN.
	Parameters string `json:"parameters"`
	// ConnectionPool is the configuration for a MySQL connection pool usually used by the database/sql package.
	ConnectionPool ConnectionPool `json:"connection_pool"`
}

// ConnectionPool is the configuration for a MySQL connection pool usually used by the database/sql package.
type ConnectionPool struct {
	// ConnMaxLifetime is the maximum amount of time a connection may be reused.
	ConnMaxLifetime *Duration `json:"conn_max_lifetime"`
	// MaxIdleConnections is the maximum number of idle connections in the connection pool.
	MaxIdleConnections *int `json:"max_idle_connections"`
	// MaxOpenConnections is the maximum number of  open connections in the connection pool.
	MaxOpenConnections *int `json:"max_open_connections"`
	// ConnMaxIdleTime is the maximum amount of time a connection may be idle.
	ConnMaxIdleTime *Duration `json:"conn_max_idle_time"`
}

// Duration is a wrapper for time.Duration that allows it to be marshalled and unmarshalled from JSON as a string.
// This type should not propagate beyond the scope of parsing the configuration.
type Duration time.Duration

// MarshalJSON marshals the duration as a string.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON unmarshalls the duration from a string. The string must be a valid duration string.
// See https://golang.org/pkg/time/#ParseDuration for more information.
func (d *Duration) UnmarshalJSON(bytes []byte) error {
	var duration string
	if err := json.Unmarshal(bytes, &duration); err != nil {
		return err
	}

	parsedDuration, err := time.ParseDuration(duration)
	if err != nil {
		return fmt.Errorf("invalid duration: %q", duration)
	}

	*d = Duration(parsedDuration)
	return nil
}

// Connections represents a set of connections to a MySQL database.
type Connections interface {
	// Get returns a connection to the MySQL database with the given name.
	// The name must match the name of a connection defined in the configuration.
	Get(name string) (*sql.DB, error)

	// List returns a list of all connections defined in the configuration.
	// The connections are returned in a non-deterministic order.
	// A common use case for this method is to ping all the connections at startup to verify that they are working.
	List() []*sql.DB

	// Close closes all the connections to the MySQL databases.
	// It should be called when the application is shutting down.
	// It tries to close all the connections even if some of them fail to close.
	// It returns the first error encountered while closing the connections.
	Close() error
}

type connections struct {
	dbs map[string]*sql.DB
}

// Open opens one or more connections to a MySQL database.
// It returns an error if the configuration is invalid or if it fails to open any of the connections.
func Open(config Config) (Connections, error) {
	if config.DSN == "" && config.Cluster == "" && config.HACluster == "" {
		return nil, errors.New("invalid MySQL config: DSN, Cluster and HACluster are empty")
	}

	if config.DSN != "" && (config.Cluster != "" || config.HACluster != "") {
		return nil, errors.New("invalid MySQL config: DSN is mutually exclusive with Cluster and HACluster")
	}

	if config.Cluster != "" && config.HACluster != "" {
		return nil, errors.New("invalid MySQL config: Cluster is mutually exclusive with HACluster")
	}

	if config.DSN != "" && config.Schema != "" {
		return nil, errors.New("invalid MySQL config: DSN is mutually exclusive with Schema since the schema is already defined in the DSN")
	}

	if config.DSN == "" && config.Schema == "" {
		return nil, errors.New("invalid MySQL config: when DSN is empty the Schema must be defined")
	}

	if len(config.Connections) == 0 {
		return nil, errors.New("invalid MySQL config: no connections defined")
	}

	if err := validateDuplicateNames(config.Connections); err != nil {
		return nil, err
	}

	// For each connection defined in the configuration create a connection pool.
	dbs := make(map[string]*sql.DB)
	for _, connectionConfig := range config.Connections {
		var db *sql.DB
		var err error

		if config.DSN == "" && (!connectionConfig.IsMaster && !connectionConfig.IsReadOnly) {
			return nil, fmt.Errorf("invalid MySQL config: cannot write to a replica: connection %q", connectionConfig.Name)
		}

		if config.DSN != "" {
			db, err = openDSN(config.DSN)
		} else if config.Cluster != "" {
			db, err = openMySQL(config.Cluster, config.Schema, connectionConfig)
		} else if config.HACluster != "" {
			db, err = openMySQLHA(config.HACluster, config.Schema, connectionConfig)
		}

		if err != nil {
			return nil, err
		}

		// Set the connection pool parameters if they are defined. Otherwise, use the default values
		// defined by the database/sql package which are not necessarily the default zero values.
		// For example, MaxIdleConnections is 2 by default.
		if connectionConfig.ConnectionPool.ConnMaxLifetime != nil {
			db.SetConnMaxLifetime(time.Duration(*connectionConfig.ConnectionPool.ConnMaxLifetime))
		}

		if connectionConfig.ConnectionPool.MaxIdleConnections != nil {
			db.SetMaxIdleConns(*connectionConfig.ConnectionPool.MaxIdleConnections)
		}

		if connectionConfig.ConnectionPool.MaxOpenConnections != nil {
			db.SetMaxOpenConns(*connectionConfig.ConnectionPool.MaxOpenConnections)
		}

		if connectionConfig.ConnectionPool.ConnMaxIdleTime != nil {
			db.SetConnMaxIdleTime(time.Duration(*connectionConfig.ConnectionPool.ConnMaxIdleTime))
		}

		dbs[connectionConfig.Name] = db
	}

	return &connections{
		dbs: dbs,
	}, nil
}

// validateDuplicateNames validates that there are no duplicated connection names.
func validateDuplicateNames(connections []Connection) error {
	connectionNames := make(map[string]struct{})
	for _, connection := range connections {
		if _, ok := connectionNames[connection.Name]; ok {
			return fmt.Errorf("invalid MySQL config: duplicated connection name %q", connection.Name)
		}
		connectionNames[connection.Name] = struct{}{}
	}
	return nil
}

func openMySQL(cluster, schema string, config Connection) (*sql.DB, error) {
	var host string
	var username string
	var password string

	clusterInUpperCase := strings.ToUpper(cluster)
	schemaInUpperCase := strings.ToUpper(schema)

	if config.IsMaster {
		host = os.Getenv(fmt.Sprintf("DB_MYSQL_%s_%s_%s_ENDPOINT",
			clusterInUpperCase, schemaInUpperCase, schemaInUpperCase))
	} else {
		host = os.Getenv(fmt.Sprintf("DB_MYSQL_%s_%s_%s_LOCAL_REPLICA_ENDPOINT",
			clusterInUpperCase, schemaInUpperCase, schemaInUpperCase))
	}

	if config.IsReadOnly {
		username = fmt.Sprintf("%s_RPROD", schema)
		password = os.Getenv(fmt.Sprintf("DB_MYSQL_%s_%s_%s_RPROD",
			clusterInUpperCase, schemaInUpperCase, schemaInUpperCase))
	} else {
		username = fmt.Sprintf("%s_WPROD", schema)
		password = os.Getenv(fmt.Sprintf("DB_MYSQL_%s_%s_%s_WPROD",
			clusterInUpperCase, schemaInUpperCase, schemaInUpperCase))
	}

	// dsn has the following format: "username:password@tcp(host:port)/schema?parameters"
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, host, schema)
	if config.Parameters != "" {
		dsn = fmt.Sprintf("%s?%s", dsn, config.Parameters)
	}

	return openDSN(dsn)
}

func openMySQLHA(cluster, schema string, config Connection) (*sql.DB, error) {
	var host string
	var username string
	var password string

	clusterInUpperCase := strings.ToUpper(cluster)
	schemaInUpperCase := strings.ToUpper(schema)

	if config.IsMaster {
		host = os.Getenv(fmt.Sprintf("DB_HA_MYSQL_%s_%s_%s_WR_ENDPOINT", clusterInUpperCase, schemaInUpperCase, schemaInUpperCase))
	} else {
		host = os.Getenv(fmt.Sprintf("DB_HA_MYSQL_%s_%s_%s_RO_ENDPOINT", clusterInUpperCase, schemaInUpperCase, schemaInUpperCase))
	}

	if config.IsReadOnly {
		username = fmt.Sprintf("%s_RPROD", schema)
		password = os.Getenv(fmt.Sprintf("DB_HA_MYSQL_%s_%s_%s_RPROD", clusterInUpperCase, schemaInUpperCase, schemaInUpperCase))
	} else {
		username = fmt.Sprintf("%s_WPROD", schema)
		password = os.Getenv(fmt.Sprintf("DB_HA_MYSQL_%s_%s_%s_WPROD", clusterInUpperCase, schemaInUpperCase, schemaInUpperCase))
	}

	// dsn has the following format: "username:password@tcp(host:port)/schema?parameters"
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, host, schema)
	if config.Parameters != "" {
		dsn = fmt.Sprintf("%s?%s", dsn, config.Parameters)
	}

	return openDSN(dsn)
}

// openDSN opens a connection to a MySQL database using the given DSN.
func openDSN(dsn string) (*sql.DB, error) {
	db, err := sql.Open(getDriverName(), dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// getDriverName returns the driver name to use for the MySQL connection.
// It returns "nrmysql" if the driver is available, otherwise it returns "mysql".
// To include the "nrmysql" driver you need to import the nrmysql package.
// For example:
//
//	import (
//		_ "github.com/newrelic/go-agent/v3/integrations/nrmysql"
//	)
func getDriverName() string {
	for _, name := range sql.Drivers() {
		if name == "nrmysql" {
			return "nrmysql"
		}
	}
	return "mysql"
}

// Get implements the Connection interface.
func (c *connections) Get(name string) (*sql.DB, error) {
	connection, ok := c.dbs[name]
	if !ok {
		return nil, fmt.Errorf("unknown connection name %s", name)
	}

	return connection, nil
}

// List implements the Connection interface.
func (c *connections) List() []*sql.DB {
	return maps.Values(c.dbs)
}

// Close implements the Connection interface.
func (c *connections) Close() error {
	// Put the keys of the map in a sorted slice so that we close the connections in a deterministic order.
	// Specially useful for tests.
	names := maps.Keys(c.dbs)
	sort.Strings(names)

	var errs []string
	for _, name := range names {
		if err := c.dbs[name].Close(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close connections: %s", strings.Join(errs, ", "))
	}

	return nil
}
