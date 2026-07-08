package events

import (
	"fmt"
	"reflect"
	"sync"
)

type Handler1[T1 any] interface {
	// apply invokes the callable with the given arguments. This variant of apply tries to match as many arguments of
	// the event payload to the parameter list of the callable (in order as fired only). The callable will not be
	// invoked with more parameters than it supports. If the callable has too many arguments, the remaining parameters
	// will be invoked with the parameters' zero values.
	apply(args ...any) error
	// apply1 invokes the callable with the exact argument(s).
	apply1(arg1 T1)
	// getCallable returns the callable Value.
	getCallable() reflect.Value
	// isOnce returns whether or not this Handler is to be invoked only once and then removed from the handler list.
	isOnce() bool
	// isOnce returns whether or not this Handler is to be invoked asynchronously.
	isAsync() bool
}

type handler1[T1 any] struct {
	// callable is the object that will be invoked.
	callable reflect.Value
	// callableArgs is the argument list that the callable will be invoked with. This eliminates allocating a new slice
	// & slice header every time the callable is invoked.
	callableArgs []reflect.Value
	// mutex ensures that the callable is only ever invoked sequentially.
	mutex             sync.Mutex
	subscriptionFlags SubscriptionFlag
}

func (h *handler1[T1]) apply(args ...any) error {
	if len(args) != 1 {
		return fmt.Errorf("expected exactly 1 argument; %d provided", len(args))
	}
	h.apply1(args[0].(T1))
	return nil
}

func (h *handler1[T1]) apply1(arg1 T1) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.callableArgs[0] = reflect.ValueOf(arg1)
	h.callable.Call(h.callableArgs)
}

func (h *handler1[T1]) getCallable() reflect.Value {
	return h.callable
}

func (h *handler1[T1]) isOnce() bool {
	return h.subscriptionFlags&SubscriptionOnce == SubscriptionOnce
}

func (h *handler1[T1]) isAsync() bool {
	return h.subscriptionFlags&SubscriptionAsync == SubscriptionAsync
}

type Callable1[T1 any] = func(T1)

func newHandler1[T1 any](callable Callable1[T1], options ...SubscriptionModifier) (Handler1[T1], error) {
	callableValue := reflect.ValueOf(callable)
	if kind := callableValue.Kind(); kind != reflect.Func {
		return nil, fmt.Errorf("%s is not of type reflect.Func", kind)
	}
	callableType := callableValue.Type()
	callableNumIn := callableType.NumIn()
	if callableNumIn != 1 {
		return nil, fmt.Errorf("The callable doesn't have exactly one parameter: %d", callableNumIn)
	}
	if callableType.In(0) != reflect.TypeFor[T1]() {
		return nil, fmt.Errorf("The callable's first parameter doesn't match first generic type: %v", callableType.In(0))
	}

	h := &handler1[T1]{
		callable:          callableValue,
		callableArgs:      make([]reflect.Value, callableNumIn),
		mutex:             sync.Mutex{},
		subscriptionFlags: 0,
	}
	for _, option := range options {
		option(&h.subscriptionFlags)
	}
	return h, nil
}

var _ Handler1[any] = (*handler1[any])(nil)
var _ Handler = (*handler1[any])(nil)
