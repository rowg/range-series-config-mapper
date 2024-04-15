package logger

import (
	"fmt"
	"log"
)

type Logger interface {
	Fatal(msg string)
	Fatalf(msg string, args ...any)
}

type StdLogger struct {
}

func (l *StdLogger) Fatal(msg string) {
	log.Fatal(msg)
}

func (l *StdLogger) Fatalf(msg string, args ...any) {
	log.Fatalf(msg, args...)
}

type TestLogger struct {
	Logs        []string
	FatalCalled bool
}

func (l *TestLogger) Fatal(msg string) {
	l.Logs = append(l.Logs, msg)
	l.FatalCalled = true
}

func (l *TestLogger) Fatalf(msg string, args ...any) {
	l.Logs = append(l.Logs, fmt.Sprintf(msg, args...))
	l.FatalCalled = true
}
