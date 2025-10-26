package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	logger *log.Logger
}

var defaultLogger *Logger

func init() {
	defaultLogger = New()
}

func New() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", 0),
	}
}

func (l *Logger) logWithLevel(level, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] [%s] %s", timestamp, level, message)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.logWithLevel("INFO", format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.logWithLevel("ERROR", format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.logWithLevel("WARN", format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.logWithLevel("DEBUG", format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}
