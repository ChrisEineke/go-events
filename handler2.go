package events

import (
	"fmt"
	"reflect"
	"sync"
)

type Callable2[T1, T2 any] = func(T1, T2) error
type Callable2NoError[T1, T2 any] = func(T1, T2)

type Handler2[T1, T2 any] interface {
	// apply invokes the callable with the given arguments. This variant of apply tries to match as many arguments of
	// the event payload to the parameter list of the callable (in order as fired only). The callable will not be
	// invoked with more parameters than it supports. If the callable has too many arguments, the remaining parameters
	// will be invoked with the parameters' zero values.
	apply(args ...any) error
	// apply2 invokes the callable with the exact argument(s).
	apply2(arg1 T1, arg2 T2) error
	// getCallable returns the callable Value.
	getCallable() reflect.Value
	// isOnce returns whether or not this Handler is to be invoked only once and then removed from the handler list.
	isOnce() bool
	// isOnce returns whether or not this Handler is to be invoked asynchronously.
	isAsync() bool
}

type handler2[T1, T2 any] struct {
	// callable is the object that will be invoked.
	callable      Callable2[T1, T2]
	callableValue reflect.Value
	// mutex ensures that the callable is only ever invoked sequentially.
	mutex             sync.Mutex
	subscriptionFlags SubscriptionFlag
}

func (h *handler2[T1, T2]) apply(args ...any) error {
	if len(args) != 2 {
		return fmt.Errorf("expected exactly 2 arguments; %d provided", len(args))
	}
	return h.apply2(args[0].(T1), args[1].(T2))
}

func (h *handler2[T1, T2]) apply2(arg1 T1, arg2 T2) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.callable(arg1, arg2)
}

func (h *handler2[T1, T2]) getCallable() reflect.Value {
	return h.callableValue
}

func (h *handler2[T1, T2]) isOnce() bool {
	return h.subscriptionFlags&SubscriptionOnce == SubscriptionOnce
}

func (h *handler2[T1, T2]) isAsync() bool {
	return h.subscriptionFlags&SubscriptionAsync == SubscriptionAsync
}

func newHandler2[T1, T2 any](callable any, options ...SubscriptionModifier) (Handler2[T1, T2], error) {
	var specificCallable Callable2[T1, T2]
	switch v := callable.(type) {
	case Callable2[T1, T2]:
		specificCallable = v
	case Callable2NoError[T1, T2]:
		specificCallable = func(arg1 T1, arg2 T2) error {
			v(arg1, arg2)
			return nil
		}
	default:
		return nil, fmt.Errorf("The callable's parameter list doesn't match the event's generic type list: %v", reflect.TypeOf(callable))
	}
	h := &handler2[T1, T2]{
		callable:          specificCallable,
		callableValue:     reflect.ValueOf(callable),
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
