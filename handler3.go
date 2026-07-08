package events

import (
	"fmt"
	"reflect"
	"sync"
)

type Handler3[T1, T2, T3 any] interface {
	// apply invokes the callable with the given arguments. This variant of apply tries to match as many arguments of
	// the event payload to the parameter list of the callable (in order as fired only). The callable will not be
	// invoked with more parameters than it supports. If the callable has too many arguments, the remaining parameters
	// will be invoked with the parameters' zero values.
	apply(args ...any) error
	// apply3 invokes the callable with the exact argument(s).
	apply3(arg1 T1, arg2 T2, arg3 T3)
	// getCallable returns the callable Value.
	getCallable() reflect.Value
	// isOnce returns whether or not this Handler is to be invoked only once and then removed from the handler list.
	isOnce() bool
	// isOnce returns whether or not this Handler is to be invoked asynchronously.
	isAsync() bool
}

type handler3[T1, T2, T3 any] struct {
	// callable is the object that will be invoked.
	callable reflect.Value
	// callableArgs is the argument list that the callable will be invoked with. This eliminates allocating a new slice
	// & slice header every time the callable is invoked.
	callableArgs []reflect.Value
	// mutex ensures that the callable is only ever invoked sequentially.
	mutex             sync.Mutex
	subscriptionFlags SubscriptionFlag
}

func (h *handler3[T1, T2, T3]) apply(args ...any) error {
	if len(args) != 3 {
		return fmt.Errorf("expected exactly 3 argument; %d provided", len(args))
	}
	h.apply3(args[0].(T1), args[1].(T2), args[2].(T3))
	return nil
}

func (h *handler3[T1, T2, T3]) apply3(arg1 T1, arg2 T2, arg3 T3) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.callableArgs[0] = reflect.ValueOf(arg1)
	h.callableArgs[1] = reflect.ValueOf(arg2)
	h.callableArgs[2] = reflect.ValueOf(arg3)
	h.callable.Call(h.callableArgs)
}

func (h *handler3[T1, T2, T3]) getCallable() reflect.Value {
	return h.callable
}

func (h *handler3[T1, T2, T3]) isOnce() bool {
	return h.subscriptionFlags&SubscriptionOnce == SubscriptionOnce
}

func (h *handler3[T1, T2, T3]) isAsync() bool {
	return h.subscriptionFlags&SubscriptionAsync == SubscriptionAsync
}

type Callable3[T1, T2, T3 any] = func(T1, T2, T3)

func newHandler3[T1, T2, T3 any](callable Callable3[T1, T2, T3], options ...SubscriptionModifier) (Handler3[T1, T2, T3], error) {
	callableValue := reflect.ValueOf(callable)
	if kind := callableValue.Kind(); kind != reflect.Func {
		return nil, fmt.Errorf("%s is not of type reflect.Func", kind)
	}
	callableType := callableValue.Type()
	callableNumIn := callableType.NumIn()
	if callableNumIn != 3 {
		return nil, fmt.Errorf("The callable doesn't have exactly three parameters: %d", callableNumIn)
	}
	if callableType.In(0) != reflect.TypeFor[T1]() {
		return nil, fmt.Errorf("The callable's first parameter doesn't match first generic type: %v", callableType.In(0))
	}
	if callableType.In(1) != reflect.TypeFor[T2]() {
		return nil, fmt.Errorf("The callable's second parameter doesn't match second generic type: %v", callableType.In(1))
	}
	if callableType.In(2) != reflect.TypeFor[T3]() {
		return nil, fmt.Errorf("The callable's third parameter doesn't match third generic type: %v", callableType.In(2))
	}

	h := &handler3[T1, T2, T3]{
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

var _ Handler3[any, any, any] = (*handler3[any, any, any])(nil)
var _ Handler = (*handler3[any, any, any])(nil)
