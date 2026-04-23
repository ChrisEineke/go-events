package EventBus

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

func (l *LoggerWare) OnUse(e *Event) error {
	fmt.Fprintf(l.outputStream, "%sOnUse: %s\n", l.prefix, e.N)
	return nil
}

func (l *LoggerWare) OnDisuse(e *Event) error {
	fmt.Fprintf(l.outputStream, "%sOnDisuse: %s\n", l.prefix, e.N)
	return nil
}

func (l *LoggerWare) OnPreFire(e *Event, args ...any) {
	fmt.Fprintf(l.outputStream, "%sOnPreFire: %s: %v\n", l.prefix, e.N, args)
}

func (l *LoggerWare) OnPostFire(e *Event, args ...any) {
	fmt.Fprintf(l.outputStream, "%sOnPostFire: %s: %v\n", l.prefix, e.N, args)
}
