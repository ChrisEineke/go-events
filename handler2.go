package events

import (
	"fmt"
	"reflect"
	"sync"
)

type Handler2[T1, T2 any] interface {
	// apply invokes the callable with the given arguments. This variant of apply tries to match as many arguments of
	// the event payload to the parameter list of the callable (in order as fired only). The callable will not be
	// invoked with more parameters than it supports. If the callable has too many arguments, the remaining parameters
	// will be invoked with the parameters' zero values.
	apply(args ...any) error
	// apply2 invokes the callable with the exact argument(s).
	apply2(arg1 T1, arg2 T2)
	// getCallable returns the callable Value.
	getCallable() reflect.Value
	// isOnce returns whether or not this Handler is to be invoked only once and then removed from the handler list.
	isOnce() bool
	// isOnce returns whether or not this Handler is to be invoked asynchronously.
	isAsync() bool
}

type handler2[T1, T2 any] struct {
	// callable is the object that will be invoked.
	callable reflect.Value
	// callableArgs is the argument list that the callable will be invoked with. This eliminates allocating a new slice
	// & slice header every time the callable is invoked.
	callableArgs []reflect.Value
	// mutex ensures that the callable is only ever invoked sequentially.
	mutex             sync.Mutex
	subscriptionFlags SubscriptionFlag
}

func (h *handler2[T1, T2]) apply(args ...any) error {
	if len(args) != 2 {
		return fmt.Errorf("expected exactly 2 arguments; %d provided", len(args))
	}
	h.apply2(args[0].(T1), args[1].(T2))
	return nil
}

func (h *handler2[T1, T2]) apply2(arg1 T1, arg2 T2) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.callableArgs[0] = reflect.ValueOf(arg1)
	h.callableArgs[1] = reflect.ValueOf(arg2)
	h.callable.Call(h.callableArgs)
}

func (h *handler2[T1, T2]) getCallable() reflect.Value {
	return h.callable
}

func (h *handler2[T1, T2]) isOnce() bool {
	return h.subscriptionFlags&SubscriptionOnce == SubscriptionOnce
}

func (h *handler2[T1, T2]) isAsync() bool {
	return h.subscriptionFlags&SubscriptionAsync == SubscriptionAsync
}

type Callable2[T1, T2 any] = func(T1, T2)

func newHandler2[T1, T2 any](callable Callable2[T1, T2], options ...SubscriptionModifier) (Handler2[T1, T2], error) {
	callableValue := reflect.ValueOf(callable)
	if kind := callableValue.Kind(); kind != reflect.Func {
		return nil, fmt.Errorf("%s is not of type reflect.Func", kind)
	}
	callableType := callableValue.Type()
	callableNumIn := callableType.NumIn()
	if callableNumIn != 2 {
		return nil, fmt.Errorf("The callable doesn't have exactly two parameters: %d", callableNumIn)
	}
	if callableType.In(0) != reflect.TypeFor[T1]() {
		return nil, fmt.Errorf("The callable's first parameter doesn't match first generic type: %v", callableType.In(0))
	}
	if callableType.In(1) != reflect.TypeFor[T2]() {
		return nil, fmt.Errorf("The callable's second parameter doesn't match second generic type: %v", callableType.In(1))
	}

	h := &handler2[T1, T2]{
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

var _ Handler2[any, any] = (*handler2[any, any])(nil)
var _ Handler = (*handler2[any, any])(nil)
