package events

import (
	"fmt"
	"io"
)

type LoggerWare struct {
	Handlerware

	outputStream io.Writer
	prefix       string
}

func Logger(os io.Writer, prefix string) Handlerware {
	return &LoggerWare{
		outputStream: os,
		prefix:       prefix,
	}
}

func (l *LoggerWare) OnUse(e *E) error {
	fmt.Fprintf(l.outputStream, "%sOnUse: %s\n", l.prefix, e.N)
	return nil
}

func (l *LoggerWare) OnDisuse(e *E) error {
	fmt.Fprintf(l.outputStream, "%sOnDisuse: %s\n", l.prefix, e.N)
	return nil
}

func (l *LoggerWare) OnAllPreFire(e *E, args []any) {
	fmt.Fprintf(l.outputStream, "%sOnAllPreFire: %s: %v\n", l.prefix, e.N, args)
}

func (l *LoggerWare) OnPreFire(e *E, h Handler, args []any) {
	fmt.Fprintf(l.outputStream, "%sOnPreFire: %s: %v: %v\n", l.prefix, e.N, h, args)
}

func (l *LoggerWare) OnPostFire(e *E, h Handler, args []any) {
	fmt.Fprintf(l.outputStream, "%sOnPostFire: %s: %v: %v\n", l.prefix, e.N, h, args)
}

func (l *LoggerWare) OnAllPostFire(e *E, args []any) {
	fmt.Fprintf(l.outputStream, "%sOnAllPostFire: %s: %v\n", l.prefix, e.N, args)
}
