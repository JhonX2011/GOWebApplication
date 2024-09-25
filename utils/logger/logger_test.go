package logger

import (
	"bytes"
	"log"
	"os"
	"testing"

	mocks "github.com/JhonX2011/GOWebApplication/test/mocks"
	"github.com/stretchr/testify/assert"
)

const expectedMsg = "logger message"

type loggerScenery struct {
	logger     *logger
	IL         Logger
	osExitMock *mocks.OSExitMock
}

func givenLoggerScenery() *loggerScenery {
	varLog := logger{}
	return &loggerScenery{
		logger: &varLog,
		IL:     NewLogger(DefaultOSExit),
	}
}

func givenLoggerSceneryFatalLogger() *loggerScenery {
	osExitMock := &mocks.OSExitMock{}
	osExitMock.On("Exit", 1).Once()
	return &loggerScenery{
		osExitMock: osExitMock,
		IL:         NewLogger(osExitMock.Exit),
	}
}

func (s *loggerScenery) givenBuffer() *bytes.Buffer {
	buf := new(bytes.Buffer)
	buf.WriteString(expectedMsg)
	log.SetOutput(buf)
	return buf
}

func (s *loggerScenery) givenSetEnvironment(key, value string) { //nolint:unparam
	os.Setenv(key, value)
}

func (s *loggerScenery) whenFatalLoggerExecuted() {
	s.IL.Fatal("fatal message")
}

func (s *loggerScenery) whenFatalFLoggerExecuted() {
	s.IL.Fatalf("fatal message")
}

func (s *loggerScenery) whenPanicLoggerExecuted(t *testing.T, panic bool) {
	strTest := "PanicFLogger Test"
	strMessage := "panicf message"
	if panic {
		strTest = "PanicLogger Test"
		strMessage = "panic message"
	}
	t.Run(strTest, func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r != nil {
				expectedPanicValue := strMessage
				actualPanicValue, ok := r.(string)
				if !ok {
					t.Errorf("The value returned by the panic is not of the expected type.")
				} else if actualPanicValue != expectedPanicValue {
					t.Errorf("The value returned by the panic is incorrect\n. Expected: [%s], Got: [%s]", expectedPanicValue, actualPanicValue)
				}
			} else {
				t.Errorf("Se esperaba un panic, pero no se produjo")
			}
		}()
		if panic {
			s.logger.Panic(strMessage)
		} else {
			s.logger.Panicf(strMessage) //nolint:govet
		}
	})
}

func (s *loggerScenery) whenErrorLoggerExecuted() {
	s.logger.Error("error message")
}

func (s *loggerScenery) whenErrorFLoggerExecuted() {
	s.logger.Errorf("error message")
}

func (s *loggerScenery) whenInfoLoggerExecuted() {
	s.logger.Info("info message")
}

func (s *loggerScenery) whenInfoFLoggerExecuted() {
	s.logger.Infof("info message")
}

func (s *loggerScenery) whenWarningLoggerExecuted() {
	s.logger.Warning("warning message")
}

func (s *loggerScenery) whenWarningFLoggerExecuted() {
	s.logger.Warningf("warning message")
}

func (s *loggerScenery) whenDebugLoggerExecuted() {
	s.logger.Debug("debug message")
}

func (s *loggerScenery) whenDebugFLoggerExecuted() {
	s.logger.Debugf("debug message")
}

func (s *loggerScenery) thenLoggerError(t *testing.T, expectedMessage string, output *bytes.Buffer) { //nolint:unparam
	assert.Equal(t, expectedMessage, output.String())
}

func (s *loggerScenery) thenExecuteFatalLogger(t *testing.T) {
	s.osExitMock.AssertExpectations(t)
}

func TestFatalLogger(t *testing.T) {
	t.Parallel()
	s := givenLoggerSceneryFatalLogger()
	s.whenFatalLoggerExecuted()
	s.thenExecuteFatalLogger(t)
}

func TestFatalFLogger(t *testing.T) {
	t.Parallel()
	s := givenLoggerSceneryFatalLogger()
	s.whenFatalFLoggerExecuted()
	s.thenExecuteFatalLogger(t)
}

func TestPanicLogger(t *testing.T) {
	t.Parallel()
	s := givenLoggerScenery()
	s.whenPanicLoggerExecuted(t, true)
}

func TestPanicFLogger(t *testing.T) {
	t.Parallel()
	s := givenLoggerScenery()
	s.whenPanicLoggerExecuted(t, false)
}

func TestErrorLogger(t *testing.T) {
	t.Parallel()
	s := givenLoggerScenery()
	output := s.givenBuffer()
	s.whenErrorLoggerExecuted()
	s.thenLoggerError(t, expectedMsg, output)
}

func TestErrorFLogger(t *testing.T) {
	t.Parallel()
	s := givenLoggerScenery()
	output := s.givenBuffer()
	s.whenErrorFLoggerExecuted()
	s.thenLoggerError(t, expectedMsg, output)
}

func TestInfoLogger(t *testing.T) {
	t.Parallel()
	s := givenLoggerScenery()
	output := s.givenBuffer()
	s.whenInfoLoggerExecuted()
	s.thenLoggerError(t, expectedMsg, output)
}

func TestInfoFLogger(t *testing.T) {
	t.Parallel()
	s := givenLoggerScenery()
	output := s.givenBuffer()
	s.whenInfoFLoggerExecuted()
	s.thenLoggerError(t, expectedMsg, output)
}

func TestWarningLogger(t *testing.T) {
	t.Parallel()
	s := givenLoggerScenery()
	output := s.givenBuffer()
	s.whenWarningLoggerExecuted()
	s.thenLoggerError(t, expectedMsg, output)
}

func TestWarningFLogger(t *testing.T) {
	t.Parallel()
	s := givenLoggerScenery()
	output := s.givenBuffer()
	s.whenWarningFLoggerExecuted()
	s.thenLoggerError(t, expectedMsg, output)
}

func TestDebugLoggerModeDebugTrue(t *testing.T) {
	t.Parallel()
	s := givenLoggerScenery()
	s.givenSetEnvironment("MODE_DEBUG", "true")
	output := s.givenBuffer()
	s.whenDebugLoggerExecuted()
	s.thenLoggerError(t, expectedMsg, output)
}

func TestDebugLoggerModeDebugFalse(t *testing.T) {
	t.Parallel()
	s := givenLoggerScenery()
	s.givenSetEnvironment("MODE_DEBUG", "false")
	output := s.givenBuffer()
	s.whenDebugLoggerExecuted()
	s.thenLoggerError(t, expectedMsg, output)
}

func TestDebugLoggerModeDebugFTrue(t *testing.T) {
	t.Parallel()
	s := givenLoggerScenery()
	s.givenSetEnvironment("MODE_DEBUG", "true")
	output := s.givenBuffer()
	s.whenDebugFLoggerExecuted()
	s.thenLoggerError(t, expectedMsg, output)
}

func TestDebugLoggerModeDebugFFalse(t *testing.T) {
	t.Parallel()
	s := givenLoggerScenery()
	s.givenSetEnvironment("MODE_DEBUG", "false")
	output := s.givenBuffer()
	s.whenDebugFLoggerExecuted()
	s.thenLoggerError(t, expectedMsg, output)
}
