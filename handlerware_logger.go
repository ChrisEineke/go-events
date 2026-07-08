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

func (l *LoggerWare) OnUse(e Event) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnUse: %s\n", l.prefix, e.Name())
	return err
}

func (l *LoggerWare) OnDisuse(e Event) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnDisuse: %s\n", l.prefix, e.Name())
	return err
}

func (l *LoggerWare) OnSubscribe(e Event, h Handler) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnSubscribe: %s %+v\n", l.prefix, e.Name(), h)
	return err
}

func (l *LoggerWare) OnUnsubscribe(e Event, h Handler) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnUnsubscribe: %s %+v\n", l.prefix, e.Name(), h)
	return err
}

func (l *LoggerWare) OnAllPreFire(e Event, args []any) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnAllPreFire: %s: %+v\n", l.prefix, e.Name(), args)
	return err
}

func (l *LoggerWare) OnPreFire(e Event, h Handler, args []any) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnPreFire: %s: %+v: %+v\n", l.prefix, e.Name(), h, args)
	return err
}

func (l *LoggerWare) OnPostFire(e Event, h Handler, args []any) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnPostFire: %s: %+v: %+v\n", l.prefix, e.Name(), h, args)
	return err
}

func (l *LoggerWare) OnAllPostFire(e Event, args []any) error {
	_, err := fmt.Fprintf(l.outputStream, "%sOnAllPostFire: %s: %+v\n", l.prefix, e.Name(), args)
	return err
}

var _ Handlerware = (*LoggerWare)(nil)
