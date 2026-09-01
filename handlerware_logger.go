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

func (l *LoggerWare) OnUse(e EventSource) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnUse: %s\n", l.prefix, e.Name())
	return err
}

func (l *LoggerWare) OnDisuse(e EventSource) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnDisuse: %s\n", l.prefix, e.Name())
	return err
}

func (l *LoggerWare) OnSubscribe(e EventSource, h Handler) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnSubscribe: %s %+v\n", l.prefix, e.Name(), h)
	return err
}

func (l *LoggerWare) OnUnsubscribe(e EventSource, h Handler) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnUnsubscribe: %s %+v\n", l.prefix, e.Name(), h)
	return err
}

func (l *LoggerWare) OnAllPreFire(e EventSource, args ...any) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnAllPreFire: %s: %+v\n", l.prefix, e.Name(), args)
	return err
}

func (l *LoggerWare) OnPreFire(e EventSource, h Handler, args ...any) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnPreFire: %s: %+v: %+v\n", l.prefix, e.Name(), h, args)
	return err
}

func (l *LoggerWare) OnPostFire(e EventSource, h Handler, args ...any) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnPostFire: %s: %+v: %+v\n", l.prefix, e.Name(), h, args)
	return err
}

func (l *LoggerWare) OnAllPostFire(e EventSource, args ...any) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnAllPostFire: %s: %+v\n", l.prefix, e.Name(), args)
	return err
}

var _ Handlerware = (*LoggerWare)(nil)
