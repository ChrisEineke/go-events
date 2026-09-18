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

	// funcValue returns this Handler's encapsulated function value.
	funcValue() reflect.Value
}

// Applicable abstracts the callback-calling machinery.
type Applicable interface {
	// apply invokes the function with the given arguments. This variant of apply tries to match as many arguments of
	// the event payload to the parameter list of the function (in order as fired only). The function will not be
	// invoked with more parameters than it supports. If the function has too many arguments, the remaining parameters
	// will be invoked with the parameters' zero values.
	apply(args ...any) error
}

func newHandler(e *E, fn any, options ...SubscriptionModifier) (Handler, error) {
	v := reflect.ValueOf(fn)
	if kind := v.Kind(); kind != reflect.Func {
		return nil, fmt.Errorf("%s: %s is not of type reflect.Func", fn, kind)
	}
	fnType := v.Type()

	if fnNumOut := fnType.NumOut(); fnNumOut != 1 {
		return nil, fmt.Errorf("%s: must return exactly one value: %d", fn, fnNumOut)
	}
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !fnType.Out(0).Implements(errorType) {
		return nil, fmt.Errorf("%s: must return exactly one value of type error: %s", fn, fnType.Out(0))
	}

	fnNumIn := fnType.NumIn()
	var h Handler
	switch fnNumIn {
	case 0:
		nh := &nullaryHandler{
			event:             e,
			fn:                v,
			mutex:             sync.Mutex{},
			subscriptionFlags: 0,
		}
		for _, option := range options {
			option(&nh.subscriptionFlags)
		}
		h = nh
	default:
		nilArgs := make([]reflect.Value, fnNumIn)
		for i := range fnNumIn {
			nilArgs[i] = reflect.New(fnType.In(i)).Elem()
		}
		nh := &nAryHandler{
			event:             e,
			fn:                v,
			fnArgs:            make([]reflect.Value, fnNumIn),
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

// nullaryHandler is a Handler that is optimized for functions without any parameters.
type nullaryHandler struct {
	event             *E
	fn                reflect.Value
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
				h.fn.Call(nil)
			})
		} else {
			h.fn.Call(nil)
		}
	} else {
		for _, hw := range h.event.handlerwares {
			if err := hw.OnPreFire(h.event, h, args...); err != nil {
				return err
			}
		}
		if isAsync {
			h.event.wg.Go(func() {
				h.fn.Call(nil)
			})
		} else {
			h.fn.Call(nil)
		}
		for _, hw := range h.event.handlerwares {
			if err := hw.OnPostFire(h.event, h, args...); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *nullaryHandler) funcValue() reflect.Value {
	return h.fn
}

type nAryHandler struct {
	event *E
	fn    reflect.Value
	// fnArgs is the argument list that the function will be invoked with. This eliminates allocating a new slice
	// & slice header every time the function is invoked.
	fnArgs []reflect.Value
	// nilArgs is a list of zero-initialized values that the argument list is initialized with. This eliminates
	// re-creating zero values for unused parameters every time the function is invoked.
	nilArgs []reflect.Value
	// mutex ensures that the function is only ever invoked sequentially.
	mutex             sync.Mutex
	subscriptionFlags SubscriptionFlag
}

func (d *nAryHandler) apply(args ...any) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// len(d.callabaleArgs) and len(d.nilArgs) are guaranteed to be the same length.
	_ = copy(d.fnArgs, d.nilArgs)
	for i := range d.fnArgs {
		if i >= len(args) || args[i] == nil {
			continue
		}
		d.fnArgs[i] = reflect.ValueOf(args[i])
	}

	isOnce := d.subscriptionFlags&SubscriptionOnce != 0
	isAsync := d.subscriptionFlags&SubscriptionAsync != 0

	if isOnce {
		d.event.handlersToRemove = append(d.event.handlersToRemove, d)
	}
	if len(d.event.handlerwares) == 0 {
		if isAsync {
			d.event.wg.Go(func() {
				d.fn.Call(d.fnArgs)
			})
		} else {
			d.fn.Call(d.fnArgs)
		}
	} else {
		for _, hw := range d.event.handlerwares {
			if err := hw.OnPreFire(d.event, d, args...); err != nil {
				return err
			}
		}
		if isAsync {
			d.event.wg.Go(func() {
				d.fn.Call(d.fnArgs)
			})
		} else {
			d.fn.Call(d.fnArgs)
		}
		for _, hw := range d.event.handlerwares {
			if err := hw.OnPostFire(d.event, d, args...); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *nAryHandler) funcValue() reflect.Value {
	return h.fn
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
