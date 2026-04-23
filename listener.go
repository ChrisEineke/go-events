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

// listener abstracts the callback-calling machinery.
type listener interface {
	sync.Locker

	// apply invokes the callable with the given arguments. This variant of apply tries to match as many arguments of
	// the event payload to the parameter list of the callable (in order as fired only). The callable will not be
	// invoked with more parameters than it supports. If the callable has too many arguments, the remaining parameters
	// will be invoked with the parameters' zero values.
	apply(args ...any)
	// safeApply invokes the callable with the given arguments.
	// If the callable's parameter list doesn't match the event payload exactly, it will return an error.
	safeApply(args ...any) error

	getListener() reflect.Value
	isOnce() bool
	setOnce(bool)
	isAsync() bool
	setAsync(bool)
	isTransactional() bool
	setTransactional(bool)
}

func newListener(callable any, options ...SubscriptionModifier) (listener, error) {
	callableValue := reflect.ValueOf(callable)
	if kind := callableValue.Kind(); kind != reflect.Func {
		return nil, fmt.Errorf("%s is not of type reflect.Func", kind)
	}
	callableType := callableValue.Type()
	callableNumIn := callableType.NumIn()
	var l listener
	switch callableNumIn {
	case 0:
		l = &nullaryListener{
			callable:          callableValue,
			mutex:             sync.Mutex{},
			subscriptionFlags: 0,
		}
	default:
		nilArgs := make([]reflect.Value, callableNumIn)
		for i := range callableNumIn {
			nilArgs[i] = reflect.New(callableType.In(i)).Elem()
		}
		l = &nAryListener{
			callable:          callableValue,
			callableArgs:      make([]reflect.Value, callableNumIn),
			nilArgs:           nilArgs,
			mutex:             sync.Mutex{},
			subscriptionFlags: 0,
		}
	}
	for _, option := range options {
		option(l)
	}
	return l, nil
}

// nullaryListener is a listener that's optimized for callables without any parameters.
type nullaryListener struct {
	callable          reflect.Value
	mutex             sync.Mutex
	subscriptionFlags SubscriptionFlag
}

func (l *nullaryListener) Lock() {
	l.mutex.Lock()
}

func (l *nullaryListener) Unlock() {
	l.mutex.Unlock()
}

func (l *nullaryListener) apply(args ...any) {
	l.callable.Call(nil)
}

func (l *nullaryListener) safeApply(args ...any) error {
	l.callable.Call(nil)
	return nil
}

func (l *nullaryListener) getListener() reflect.Value {
	return l.callable
}

func (l *nullaryListener) isOnce() bool {
	return l.subscriptionFlags&SubscriptionOnce == SubscriptionOnce
}

func (l *nullaryListener) setOnce(val bool) {
	l.subscriptionFlags |= SubscriptionOnce
}

func (l *nullaryListener) isAsync() bool {
	return l.subscriptionFlags&SubscriptionAsync == SubscriptionAsync
}

func (l *nullaryListener) setAsync(val bool) {
	l.subscriptionFlags |= SubscriptionAsync
}

func (l *nullaryListener) isTransactional() bool {
	return l.subscriptionFlags&SubscriptionTransactional == SubscriptionTransactional
}

func (l *nullaryListener) setTransactional(val bool) {
	l.subscriptionFlags |= SubscriptionTransactional
}

type nAryListener struct {
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

func (l *nAryListener) Lock() {
	l.mutex.Lock()
}

func (l *nAryListener) Unlock() {
	l.mutex.Unlock()
}

func (d *nAryListener) apply(args ...any) {
	d.callableArgs = d.nilArgs
	for i := range d.callableArgs {
		if args[i] == nil {
			continue
		}
		d.callableArgs[i] = reflect.ValueOf(args[i])
	}
	d.callable.Call(d.callableArgs)
}

func (d *nAryListener) safeApply(args ...any) error {
	callableArgsLen := len(d.nilArgs)
	payloadArgsLen := len(args)
	if callableArgsLen != payloadArgsLen {
		return fmt.Errorf("length of callable parameter list (%d) doesn't match event payload (%d)",
			callableArgsLen, payloadArgsLen)
	}
	d.callableArgs = d.nilArgs
	for i, arg := range args {
		d.callableArgs[i] = reflect.ValueOf(arg)
	}
	d.callable.Call(d.callableArgs)
	return nil
}

func (l *nAryListener) getListener() reflect.Value {
	return l.callable
}

func (l *nAryListener) isOnce() bool {
	return l.subscriptionFlags&SubscriptionOnce == SubscriptionOnce
}

func (l *nAryListener) setOnce(val bool) {
	l.subscriptionFlags |= SubscriptionOnce
}

func (l *nAryListener) isAsync() bool {
	return l.subscriptionFlags&SubscriptionAsync == SubscriptionAsync
}

func (l *nAryListener) setAsync(val bool) {
	l.subscriptionFlags |= SubscriptionAsync
}

func (l *nAryListener) isTransactional() bool {
	return l.subscriptionFlags&SubscriptionTransactional == SubscriptionTransactional
}

func (l *nAryListener) setTransactional(val bool) {
	l.subscriptionFlags |= SubscriptionTransactional
}

type SubscriptionModifier func(listener)

// Sync invokes the listener synchronously (the default).
func Sync() SubscriptionModifier {
	return func(l listener) {
		l.setAsync(false)
	}
}

// Async invokes the listener asynchronously.
func Async() SubscriptionModifier {
	return func(l listener) {
		l.setAsync(true)
	}
}

// Always keeps the listener registered after being called (the default).
func Always() SubscriptionModifier {
	return func(l listener) {
		l.setOnce(false)
	}
}

// Once removes the listener after being called once.
func Once() SubscriptionModifier {
	return func(l listener) {
		l.setOnce(true)
	}
}

// NonTransactional invokes subsequent listeners for a topic concurrently (the default).
func NonTransactional() SubscriptionModifier {
	return func(l listener) {
		l.setTransactional(false)
	}
}

// Transactional invokes subsequent listeners for a topic serially (true).
func Transactional() SubscriptionModifier {
	return func(l listener) {
		l.setTransactional(true)
	}
}
