package mylogger

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type loglevel uint64

const (
	UNKNOWE loglevel = iota
	DEBUG
	TRACE
	INFO
	WARNING
	ERROR
	FATAL
)

type ConsoleLogger struct {
	level loglevel
}

func parseLogevel(s string) (loglevel, error) {
	s = strings.ToLower(s)
	switch s {
	case "debug":
		return DEBUG, nil
	case "tarce":
		return TRACE, nil
	case "info":
		return INFO, nil
	case "warning":
		return WARNING, nil
	case "error":
		return ERROR, nil
	case "fatal":
		return FATAL, nil
	default:
		err := errors.New("日志级别错误")
		return UNKNOWE, err
	}
}
func getlogstring(lv loglevel) string {

	switch lv {
	case DEBUG:
		return "DEBUG"
	case TRACE:
		return "TRACE"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}

}
func Newlog(logstr string) ConsoleLogger {
	loevel, err := parseLogevel(logstr)
	if err != nil {
		panic(err)
	}

	return ConsoleLogger{
		level: loevel}
}

func (c ConsoleLogger) enable(level loglevel) bool {
	return c.level <= level
}
func (c ConsoleLogger) log(lv loglevel, format string, a ...interface{}) {
	if c.enable(lv) {
		msg := fmt.Sprintf(format, a...) //...
		funcName, fileName, lineno := getInfo(3)
		now := time.Now()
		fmt.Printf("[%s][%s][%s:%s:%d]%s\n", now.Format("2006-01-02 03:04:05"), getlogstring(lv), funcName, fileName, lineno, msg)
	}

}

func (c ConsoleLogger) Debug(msg string, arg ...interface{}) {
	c.log(DEBUG, msg, arg...)
}
func (c ConsoleLogger) Info(msg string, arg ...interface{}) {
	c.log(INFO, msg, arg...)
}
func (c ConsoleLogger) Waring(msg string, arg ...interface{}) {
	c.log(WARNING, msg, arg...)
}
func (c ConsoleLogger) Error(msg string, arg ...interface{}) {
	c.log(ERROR, msg, arg...)
}
func (c ConsoleLogger) Fatal(msg string, arg ...interface{}) {
	c.log(FATAL, msg, arg...)
}
