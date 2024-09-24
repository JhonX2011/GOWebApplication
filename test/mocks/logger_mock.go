package infrastructuremock

import (
	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
}

type OSExitMock struct {
	mock.Mock
}

func (m *OSExitMock) Exit(code int) {
	m.Called(code)
}

func (m *MockLogger) Fatal(params ...interface{}) {
	m.Called(params)
}

func (m *MockLogger) Fatalf(msg string, params ...interface{}) {
	m.Called(msg, params)
}

func (m *MockLogger) Panic(params ...interface{}) {
	m.Called(params)
}

func (m *MockLogger) Panicf(msg string, params ...interface{}) {
	m.Called(msg, params)
}

func (m *MockLogger) Error(params ...interface{}) {
	m.Called(params)
}

func (m *MockLogger) Errorf(msg string, params ...interface{}) {
	m.Called(msg, params)
}

func (m *MockLogger) Info(params ...interface{}) {
	m.Called(params)
}

func (m *MockLogger) Infof(msg string, params ...interface{}) {
	m.Called(msg, params)
}

func (m *MockLogger) Warning(params ...interface{}) {
	m.Called(params)
}

func (m *MockLogger) Warningf(msg string, params ...interface{}) {
	m.Called(msg, params)
}

func (m *MockLogger) Debug(params ...interface{}) {
	m.Called(params)
}

func (m *MockLogger) Debugf(msg string, params ...interface{}) {
	m.Called(msg, params)
}
