package utils

import (
	"fmt"
	"io"
	"time"
)

type StepLogger struct {
	w io.Writer
}

func NewStepLogger(w io.Writer) *StepLogger {
	return &StepLogger{w: w}
}

func (l *StepLogger) log(level string, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("15:04:05")
	fmt.Fprintf(l.w, "[%s] [%s] %s\n", timestamp, level, msg)
}

func (l *StepLogger) Info(format string, args ...any) {
	l.log("INFO", format, args...)
}

func (l *StepLogger) Error(format string, args ...any) {
	l.log("ERROR", format, args...)
}

func (l *StepLogger) Warn(format string, args ...any) {
	l.log("WARN", format, args...)
}
