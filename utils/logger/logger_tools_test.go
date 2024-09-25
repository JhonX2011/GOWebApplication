package logger

import (
	"os"
	"testing"
	"time"

	mocks "github.com/JhonX2011/GOWebApplication/test/mocks"
	"github.com/stretchr/testify/assert"
)

type loggerUtilsScenery struct {
	aResult    any
	osExitMock *mocks.OSExitMock
	exitCode   int
}

func givenLoggerUtilsScenery() *loggerUtilsScenery {
	osExitMock := &mocks.OSExitMock{}
	osExitMock.On("Exit", 1).Once()
	return &loggerUtilsScenery{
		exitCode:   1,
		osExitMock: osExitMock,
	}
}

func (l *loggerUtilsScenery) whenFormatNow(t time.Time) {
	l.aResult = FormatNow(t)
}

func (l *loggerUtilsScenery) whenFileInfo(file string, line int, ok bool) {
	l.aResult = FileInfo(file, line, ok)
}

func (l *loggerUtilsScenery) whenFuncInfo(name string) {
	l.aResult = FuncInfo(name)
}

func (l *loggerUtilsScenery) whenDefaultOSExit() {
	// execute function that is executed
	DefaultOSExit(l.exitCode)
}

func (l *loggerUtilsScenery) thenEqual(t *testing.T, expectedMessage any, output any) {
	assert.Equal(t, expectedMessage, output)
}

func (l *loggerUtilsScenery) thenAssertExpectations(t *testing.T) {
	l.osExitMock.AssertExpectations(t)
}

func TestFormatNow(t *testing.T) {
	e := givenLoggerUtilsScenery()
	dateString := "2023-05-12"
	date, _ := time.Parse("2006-01-02", dateString)
	e.whenFormatNow(date)
	e.thenEqual(t, e.aResult, "2023/05/12 - 00:00:00")
}

func TestFileInfoTrue(t *testing.T) {
	e := givenLoggerUtilsScenery()
	e.whenFileInfo("tes/file", 20, true)
	e.thenEqual(t, e.aResult, "file:20")
}

func TestFileInfoFalse(t *testing.T) {
	e := givenLoggerUtilsScenery()
	e.whenFileInfo("tes/file", 20, false)
	e.thenEqual(t, e.aResult, "<???>:1")
}

func TestFuncInfo(t *testing.T) {
	e := givenLoggerUtilsScenery()
	e.whenFuncInfo("testFunction")
	e.thenEqual(t, e.aResult, "testFunction()")
}

func TestDefaultOSExit(t *testing.T) {
	e := givenLoggerUtilsScenery()
	exitFunc = e.osExitMock.Exit
	e.whenDefaultOSExit()
	e.thenAssertExpectations(t)
	exitFunc = os.Exit
}
