package infrastructuremock

import (
	"context"
	"database/sql/driver"
	"strings"
)

type MysqlDriverMock struct {
	OpenFunc  func(name string) (driver.Conn, error)
	CloseFunc func() error
}

func (m *MysqlDriverMock) Open(name string) (driver.Conn, error) {
	return m.OpenFunc(name)
}

func (m *MysqlDriverMock) OpenConnector(name string) (driver.Connector, error) {
	return &dsnConnectorMock{
		dsn:    name,
		driver: m,
	}, nil
}

type dsnConnectorMock struct {
	dsn    string
	driver driver.Driver
}

func (d *dsnConnectorMock) Connect(_ context.Context) (driver.Conn, error) {
	return d.driver.Open(d.dsn)
}

func (d *dsnConnectorMock) Driver() driver.Driver {
	return d.driver
}

func (d *dsnConnectorMock) Close() error {
	if strings.Contains(d.dsn, "close_with_error") {
		return driver.ErrBadConn
	}
	return nil
}

type DriverConnMock struct {
	PrepareFunc func(query string) (driver.Stmt, error)
	CloseFunc   func() error
	BeginFunc   func() (driver.Tx, error)
}

func (d *DriverConnMock) Prepare(query string) (driver.Stmt, error) {
	return d.PrepareFunc(query)
}

func (d *DriverConnMock) Close() error {
	return d.CloseFunc()
}

func (d *DriverConnMock) Begin() (driver.Tx, error) {
	return d.BeginFunc()
}
