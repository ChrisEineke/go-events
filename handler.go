package events

import (
	"fmt"
	"reflect"
	"sync"
)

type SubscriptionFlag int

const (
	SubscriptionOnce SubscriptionFlag = 1 << iota
	SubscriptionAsync
)

// Handler is the interface shared between generic Handlers (nullaryHandler, etc) and typ-specific Handlers (handler1,
// etc.).
type Handler interface {
	Applicable

	// callable returns this Handler's encapsulated callable Value.
	callable() reflect.Value
}

// Applicable abstracts the callback-calling machinery.
type Applicable interface {
	// apply invokes the callable with the given arguments. This variant of apply tries to match as many arguments of
	// the event payload to the parameter list of the callable (in order as fired only). The callable will not be
	// invoked with more parameters than it supports. If the callable has too many arguments, the remaining parameters
	// will be invoked with the parameters' zero values.
	apply(args ...any) error
}

func newHandler(e *E, callable any, options ...SubscriptionModifier) (Handler, error) {
	call := reflect.ValueOf(callable)
	if kind := call.Kind(); kind != reflect.Func {
		return nil, fmt.Errorf("%s: %s is not of type reflect.Func", call, kind)
	}
	callableType := call.Type()

	if callableNumOut := callableType.NumOut(); callableNumOut != 1 {
		return nil, fmt.Errorf("%s: must return exactly one value: %d", call, callableNumOut)
	}
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !callableType.Out(0).Implements(errorType) {
		return nil, fmt.Errorf("%s: must return exactly one value of type error: %s", call, callableType.Out(0))
	}

	callableNumIn := callableType.NumIn()
	var h Handler
	switch callableNumIn {
	case 0:
		nh := &nullaryHandler{
			event:             e,
			call:              call,
			mutex:             sync.Mutex{},
			subscriptionFlags: 0,
		}
		for _, option := range options {
			option(&nh.subscriptionFlags)
		}
		h = nh
	default:
		nilArgs := make([]reflect.Value, callableNumIn)
		for i := range callableNumIn {
			nilArgs[i] = reflect.New(callableType.In(i)).Elem()
		}
		nh := &nAryHandler{
			event:             e,
			call:              call,
			callableArgs:      make([]reflect.Value, callableNumIn),
			nilArgs:           nilArgs,
			mutex:             sync.Mutex{},
			subscriptionFlags: 0,
		}
		for _, option := range options {
			option(&nh.subscriptionFlags)
		}
		h = nh
	}
	return h, nil
}

// nullaryHandler is a Handler that is optimized for callables without any parameters.
type nullaryHandler struct {
	event             *E
	call              reflect.Value
	mutex             sync.Mutex
	subscriptionFlags SubscriptionFlag
}

func (h *nullaryHandler) apply(args ...any) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	isOnce := h.subscriptionFlags&SubscriptionOnce != 0
	isAsync := h.subscriptionFlags&SubscriptionAsync != 0

	if isOnce {
		h.event.handlersToRemove = append(h.event.handlersToRemove, h)
	}
	if len(h.event.handlerwares) == 0 {
		if isAsync {
			h.event.wg.Go(func() {
				h.call.Call(nil)
			})
		} else {
			h.call.Call(nil)
		}
	} else {
		for _, hw := range h.event.handlerwares {
			if err := hw.OnPreFire(h.event, h, args...); err != nil {
				return err
			}
		}
		if isAsync {
			h.event.wg.Go(func() {
				h.call.Call(nil)
			})
		} else {
			h.call.Call(nil)
		}
		for _, hw := range h.event.handlerwares {
			if err := hw.OnPostFire(h.event, h, args...); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *nullaryHandler) callable() reflect.Value {
	return h.call
}

type nAryHandler struct {
	event *E
	call  reflect.Value
	// callableArgs is the argument list that the callable will be invoked with. This eliminates allocating a new slice
	// & slice header every time the callable is invoked.
	callableArgs []reflect.Value
	// nilArgs is a list of zero-initialized values that the argument list is initialized with. This eliminates
	// re-creating zero values for unused parameters every time the callable is invoked.
	nilArgs []reflect.Value
	// mutex ensures that the callable is only ever invoked sequentially.
	mutex             sync.Mutex
	subscriptionFlags SubscriptionFlag
}

func (d *nAryHandler) apply(args ...any) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// len(d.callabaleArgs) and len(d.nilArgs) are guaranteed to be the same length.
	_ = copy(d.callableArgs, d.nilArgs)
	for i := range d.callableArgs {
		if i >= len(args) || args[i] == nil {
			continue
		}
		d.callableArgs[i] = reflect.ValueOf(args[i])
	}

	isOnce := d.subscriptionFlags&SubscriptionOnce != 0
	isAsync := d.subscriptionFlags&SubscriptionAsync != 0

	if isOnce {
		d.event.handlersToRemove = append(d.event.handlersToRemove, d)
	}
	if len(d.event.handlerwares) == 0 {
		if isAsync {
			d.event.wg.Go(func() {
				d.call.Call(d.callableArgs)
			})
		} else {
			d.call.Call(d.callableArgs)
		}
	} else {
		for _, hw := range d.event.handlerwares {
			if err := hw.OnPreFire(d.event, d, args...); err != nil {
				return err
			}
		}
		if isAsync {
			d.event.wg.Go(func() {
				d.call.Call(d.callableArgs)
			})
		} else {
			d.call.Call(d.callableArgs)
		}
		for _, hw := range d.event.handlerwares {
			if err := hw.OnPostFire(d.event, d, args...); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *nAryHandler) callable() reflect.Value {
	return h.call
}

type SubscriptionModifier func(*SubscriptionFlag)

// Sync invokes the Handler synchronously (the default).
func Sync() SubscriptionModifier {
	return func(flags *SubscriptionFlag) {
		*flags &^= SubscriptionAsync
	}
}

// Async invokes the Handler asynchronously.
func Async() SubscriptionModifier {
	return func(flags *SubscriptionFlag) {
		*flags |= SubscriptionAsync
	}
}

// Always keeps the Handler registered after being called (the default).
func Always() SubscriptionModifier {
	return func(flags *SubscriptionFlag) {
		*flags &^= SubscriptionOnce
	}
}

// Once removes the Handler after being called once.
func Once() SubscriptionModifier {
	return func(flags *SubscriptionFlag) {
		*flags |= SubscriptionOnce
	}
}
