package logger

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

const (
	gError  = "Error"
	fatal   = "Fatal"
	panico  = "Panic"
	info    = "Info"
	warning = "warning"
	debug   = "Debug"
)

const value = 1

type logger struct {
	osExitFunc func(int) // out function of SS.OO
}

type Logger interface {
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Panic(...interface{})
	Panicf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Warning(...interface{})
	Warningf(string, ...interface{})
	Debug(...interface{})
	Debugf(string, ...interface{})
}

func NewLogger(fn func(int)) Logger {
	return &logger{
		osExitFunc: fn,
	}
}

func (l *logger) Fatal(v ...interface{}) {
	now := time.Now()
	pc, fi, li, ok := runtime.Caller(value)
	f := runtime.FuncForPC(pc).Name()
	fmt.Printf("%s | %s %s %s | %20s | %20s | %s \n",
		FormatNow(now), magenta, fatal, reset, FileInfo(fi, li, ok), FuncInfo(f), v)
	l.osExitFunc(1)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	now := time.Now()
	pc, fi, li, ok := runtime.Caller(value)
	f := runtime.FuncForPC(pc).Name()
	fmt.Printf("%s | %s %s %s | %20s | %20s | %s \n",
		FormatNow(now), magenta, fatal, reset, FileInfo(fi, li, ok), FuncInfo(f),
		fmt.Errorf(format, args...).Error())
	l.osExitFunc(1)
}

func (l *logger) Panic(v ...interface{}) {
	now := time.Now()
	s := fmt.Sprint(v...)
	pc, fi, li, ok := runtime.Caller(value)
	f := runtime.FuncForPC(pc).Name()
	fmt.Printf("%s | %s %s %s | %20s | %20s | %s \n",
		FormatNow(now), cyan, panico, reset, FileInfo(fi, li, ok), FuncInfo(f), v)
	panic(s)
}

func (l *logger) Panicf(format string, args ...interface{}) {
	now := time.Now()
	s := fmt.Sprintf(format, args...)
	pc, fi, li, ok := runtime.Caller(value)
	f := runtime.FuncForPC(pc).Name()
	fmt.Printf("%s | %s %s %s | %20s | %20s | %s \n",
		FormatNow(now), cyan, panico, reset, FileInfo(fi, li, ok), FuncInfo(f),
		fmt.Errorf(format, args...).Error())
	panic(s)
}

func (l *logger) Error(v ...interface{}) {
	now := time.Now()
	pc, fi, li, ok := runtime.Caller(value)
	f := runtime.FuncForPC(pc).Name()
	fmt.Printf("%s | %s %s %s | %20s | %20s | %s \n",
		FormatNow(now), red, gError, reset, FileInfo(fi, li, ok), FuncInfo(f), v)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	now := time.Now()
	pc, fi, li, ok := runtime.Caller(value)
	f := runtime.FuncForPC(pc).Name()
	fmt.Printf("%s | %s %s %s | %20s | %20s | %s \n",
		FormatNow(now), red, gError, reset, FileInfo(fi, li, ok), FuncInfo(f),
		fmt.Errorf(format, args...).Error())
}

func (l *logger) Info(v ...interface{}) {
	now := time.Now()
	pc, fi, li, ok := runtime.Caller(value)
	f := runtime.FuncForPC(pc).Name()
	fmt.Printf("%s | %s %s %s | %20s | %20s | %s \n",
		FormatNow(now), blue, info, reset, FileInfo(fi, li, ok), FuncInfo(f), v)
}

func (l *logger) Infof(format string, args ...interface{}) {
	now := time.Now()
	pc, fi, li, ok := runtime.Caller(value)
	f := runtime.FuncForPC(pc).Name()
	fmt.Printf("%s | %s %s %s | %20s | %20s | %s \n",
		FormatNow(now), blue, info, reset, FileInfo(fi, li, ok), FuncInfo(f),
		fmt.Sprintf(format, args...))
}

func (l *logger) Warning(v ...interface{}) {
	now := time.Now()
	pc, fi, li, ok := runtime.Caller(value)
	f := runtime.FuncForPC(pc).Name()
	fmt.Printf("%s | %s %s %s | %20s | %20s | %s \n",
		FormatNow(now), yellow, warning, reset, FileInfo(fi, li, ok), FuncInfo(f), v)
}

func (l *logger) Warningf(format string, args ...interface{}) {
	now := time.Now()
	pc, fi, li, ok := runtime.Caller(value)
	f := runtime.FuncForPC(pc).Name()
	fmt.Printf("%s | %s %s %s | %20s | %20s | %s \n",
		FormatNow(now), yellow, warning, reset, FileInfo(fi, li, ok), FuncInfo(f),
		fmt.Sprintf(format, args...))
}

func (l *logger) Debug(v ...interface{}) {
	if os.Getenv("MODE_DEBUG") == "true" {
		now := time.Now()
		pc, fi, li, ok := runtime.Caller(value)
		f := runtime.FuncForPC(pc).Name()
		fmt.Printf("%s | %s %s %s | %20s | %20s | %s \n",
			FormatNow(now), green, debug, reset, FileInfo(fi, li, ok), FuncInfo(f), v)
	}
}

func (l *logger) Debugf(format string, args ...interface{}) {
	if os.Getenv("MODE_DEBUG") == "true" {
		now := time.Now()
		pc, fi, li, ok := runtime.Caller(value)
		f := runtime.FuncForPC(pc).Name()
		fmt.Printf("%s | %s %s %s | %20s | %20s | %s \n",
			FormatNow(now), green, debug, reset, FileInfo(fi, li, ok), FuncInfo(f),
			fmt.Sprintf(format, args...))
	}
}
