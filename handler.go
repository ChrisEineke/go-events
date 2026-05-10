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
	SubscriptionTransactional
)

// Handler abstracts the callback-calling machinery.
type Handler interface {
	sync.Locker

	// apply invokes the callable with the given arguments. This variant of apply tries to match as many arguments of
	// the event payload to the parameter list of the callable (in order as fired only). The callable will not be
	// invoked with more parameters than it supports. If the callable has too many arguments, the remaining parameters
	// will be invoked with the parameters' zero values.
	apply(args ...any)
	// safeApply invokes the callable with the given arguments.
	// If the callable's parameter list doesn't match the event payload exactly, it will return an error.
	safeApply(args ...any) error
	// getCallable returns the callable Value.
	getCallable() reflect.Value
	// isOnce returns whether or not this Handler is to be invoked only once and then removed from the handler list.
	isOnce() bool
	// isOnce returns whether or not this Handler is to be invoked asynchronously.
	isAsync() bool
	// isOnce returns whether or not this Handler is to be invoked transactionally.
	isTransactional() bool
}

func newHandler(callable any, options ...SubscriptionModifier) (Handler, error) {
	callableValue := reflect.ValueOf(callable)
	if kind := callableValue.Kind(); kind != reflect.Func {
		return nil, fmt.Errorf("%s is not of type reflect.Func", kind)
	}
	callableType := callableValue.Type()
	callableNumIn := callableType.NumIn()
	var h Handler
	switch callableNumIn {
	case 0:
		nh := &nullaryHandler{
			callable:          callableValue,
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
			callable:          callableValue,
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

// nullaryHandler is a Handler that's optimized for callables without any parameters.
type nullaryHandler struct {
	callable          reflect.Value
	mutex             sync.Mutex
	subscriptionFlags SubscriptionFlag
}

func (h *nullaryHandler) Lock() {
	h.mutex.Lock()
}

func (h *nullaryHandler) Unlock() {
	h.mutex.Unlock()
}

func (h *nullaryHandler) apply(args ...any) {
	h.callable.Call(nil)
}

func (h *nullaryHandler) safeApply(args ...any) error {
	h.callable.Call(nil)
	return nil
}

func (h *nullaryHandler) getCallable() reflect.Value {
	return h.callable
}

func (h *nullaryHandler) isOnce() bool {
	return h.subscriptionFlags&SubscriptionOnce != 0
}

func (h *nullaryHandler) isAsync() bool {
	return h.subscriptionFlags&SubscriptionAsync != 0
}

func (h *nullaryHandler) isTransactional() bool {
	return h.subscriptionFlags&SubscriptionTransactional != 0
}

type nAryHandler struct {
	callable reflect.Value
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

func (h *nAryHandler) Lock() {
	h.mutex.Lock()
}

func (h *nAryHandler) Unlock() {
	h.mutex.Unlock()
}

func (d *nAryHandler) apply(args ...any) {
	// len(d.callabaleArgs) and len(d.nilArgs) are guaranteed to be the same length.
	_ = copy(d.callableArgs, d.nilArgs)
	d.callableArgs = d.nilArgs
	for i := range d.callableArgs {
		if args[i] == nil {
			continue
		}
		d.callableArgs[i] = reflect.ValueOf(args[i])
	}
	d.callable.Call(d.callableArgs)
}

func (d *nAryHandler) safeApply(args ...any) error {
	if len(d.callableArgs) != len(args) {
		return fmt.Errorf("length of callable parameter list (%d) doesn't match event payload (%d)",
			len(d.callableArgs), len(args))
	}
	// len(d.callabaleArgs) and len(d.nilArgs) are guaranteed to be the same length.
	_ = copy(d.callableArgs, d.nilArgs)
	for i := range args {
		if args[i] == nil {
			continue
		}
		d.callableArgs[i] = reflect.ValueOf(args[i])
	}
	d.callable.Call(d.callableArgs)
	return nil
}

func (h *nAryHandler) getCallable() reflect.Value {
	return h.callable
}

func (h *nAryHandler) isOnce() bool {
	return h.subscriptionFlags&SubscriptionOnce == SubscriptionOnce
}

func (h *nAryHandler) isAsync() bool {
	return h.subscriptionFlags&SubscriptionAsync == SubscriptionAsync
}

func (h *nAryHandler) isTransactional() bool {
	return h.subscriptionFlags&SubscriptionTransactional == SubscriptionTransactional
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

// NonTransactional invokes subsequent handlers concurrently (the default).
func NonTransactional() SubscriptionModifier {
	return func(flags *SubscriptionFlag) {
		*flags &^= SubscriptionTransactional
	}
}

// Transactional invokes subsequent handlers serially.
func Transactional() SubscriptionModifier {
	return func(flags *SubscriptionFlag) {
		*flags |= SubscriptionTransactional
	}
}
