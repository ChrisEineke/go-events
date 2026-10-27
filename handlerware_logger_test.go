package events

import (
	"bytes"
	"testing"
)

func TestLoggerwareImplementsInterface(t *testing.T) {
	var _ Handlerware = &LoggerWare{}
}

func TestLoggerWare(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	loggerware := NewLoggerWare(buf, "testing: ")
	handler := func(_ int) error { return nil }
	e := &E1[int]{N: "testEvent"}
	e.Use(loggerware)
	e.On(handler)
	e.Fire(1)
	e.Off(handler)
	e.Disuse(loggerware)
}
