package events

import (
	"fmt"
	"reflect"
	"sync"
)

type Callable3[T1, T2, T3 any] = func(T1, T2, T3) error
type Callable3NoError[T1, T2, T3 any] = func(T1, T2, T3)

type Handler3[T1, T2, T3 any] interface {
	// apply invokes the callable with the given arguments. This variant of apply tries to match as many arguments of
	// the event payload to the parameter list of the callable (in order as fired only). The callable will not be
	// invoked with more parameters than it supports. If the callable has too many arguments, the remaining parameters
	// will be invoked with the parameters' zero values.
	apply(args ...any) error
	// apply3 invokes the callable with the exact argument(s).
	apply3(arg1 T1, arg2 T2, arg3 T3) error
	// getCallable returns the callable Value.
	getCallable() reflect.Value
	// isOnce returns whether or not this Handler is to be invoked only once and then removed from the handler list.
	isOnce() bool
	// isOnce returns whether or not this Handler is to be invoked asynchronously.
	isAsync() bool
}

type handler3[T1, T2, T3 any] struct {
	// callable is the object that will be invoked.
	callable      Callable3[T1, T2, T3]
	callableValue reflect.Value
	// mutex ensures that the callable is only ever invoked sequentially.
	mutex             sync.Mutex
	subscriptionFlags SubscriptionFlag
}

func (h *handler3[T1, T2, T3]) apply(args ...any) error {
	if len(args) != 3 {
		return fmt.Errorf("expected exactly 3 argument; %d provided", len(args))
	}
	return h.apply3(args[0].(T1), args[1].(T2), args[2].(T3))
}

func (h *handler3[T1, T2, T3]) apply3(arg1 T1, arg2 T2, arg3 T3) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.callable(arg1, arg2, arg3)
}

func (h *handler3[T1, T2, T3]) getCallable() reflect.Value {
	return h.callableValue
}

func (h *handler3[T1, T2, T3]) isOnce() bool {
	return h.subscriptionFlags&SubscriptionOnce == SubscriptionOnce
}

func (h *handler3[T1, T2, T3]) isAsync() bool {
	return h.subscriptionFlags&SubscriptionAsync == SubscriptionAsync
}

func newHandler3[T1, T2, T3 any](callable any, options ...SubscriptionModifier) (Handler3[T1, T2, T3], error) {
	var specificCallable Callable3[T1, T2, T3]
	switch v := callable.(type) {
	case Callable3[T1, T2, T3]:
		specificCallable = v
	case Callable3NoError[T1, T2, T3]:
		specificCallable = func(arg1 T1, arg2 T2, arg3 T3) error {
			v(arg1, arg2, arg3)
			return nil
		}
	default:
		return nil, fmt.Errorf("The callable's parameter list doesn't match the event's generic type list: %v", reflect.TypeOf(callable))
	}
	h := &handler3[T1, T2, T3]{
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

var _ Handler3[any, any, any] = (*handler3[any, any, any])(nil)
var _ Handler = (*handler3[any, any, any])(nil)
