package events

import (
	"fmt"
	"io"
)

type LoggerWare struct {
	outputStream io.Writer
	prefix       string
}

func NewLoggerWare(os io.Writer, prefix string) *LoggerWare {
	return &LoggerWare{
		outputStream: os,
		prefix:       prefix,
	}
}

func (l *LoggerWare) OnUse(e *E) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnUse: %s\n", l.prefix, e.N)
	return err
}

func (l *LoggerWare) OnDisuse(e *E) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnDisuse: %s\n", l.prefix, e.N)
	return err
}

func (l *LoggerWare) OnSubscribe(e *E, h Handler) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnSubscribe: %s %+v\n", l.prefix, e.N, h)
	return err
}

func (l *LoggerWare) OnUnsubscribe(e *E, h Handler) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnUnsubscribe: %s %+v\n", l.prefix, e.N, h)
	return err
}

func (l *LoggerWare) OnAllPreFire(e *E, args []any) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnAllPreFire: %s: %+v\n", l.prefix, e.N, args)
	return err
}

func (l *LoggerWare) OnPreFire(e *E, h Handler, args []any) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnPreFire: %s: %+v: %+v\n", l.prefix, e.N, h, args)
	return err
}

func (l *LoggerWare) OnPostFire(e *E, h Handler, args []any) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnPostFire: %s: %+v: %+v\n", l.prefix, e.N, h, args)
	return err
}

func (l *LoggerWare) OnAllPostFire(e *E, args []any) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnAllPostFire: %s: %+v\n", l.prefix, e.N, args)
	return err
}
