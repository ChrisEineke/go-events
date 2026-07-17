package events

import (
	"fmt"
	"reflect"
	"sync"
)

type Callable4[T1, T2, T3, T4 any] = func(T1, T2, T3, T4) error
type Callable4NoError[T1, T2, T3, T4 any] = func(T1, T2, T3, T4)

type Handler4[T1, T2, T3, T4 any] interface {
	// apply invokes the callable with the given arguments. This variant of apply tries to match as many arguments of
	// the event payload to the parameter list of the callable (in order as fired only). The callable will not be
	// invoked with more parameters than it supports. If the callable has too many arguments, the remaining parameters
	// will be invoked with the parameters' zero values.
	apply(args ...any) error
	// apply4 invokes the callable with the exact argument(s).
	apply4(arg1 T1, arg2 T2, arg3 T3, arg4 T4) error
	// getCallable returns the callable Value.
	getCallable() reflect.Value
	// isOnce returns whether or not this Handler is to be invoked only once and then removed from the handler list.
	isOnce() bool
	// isOnce returns whether or not this Handler is to be invoked asynchronously.
	isAsync() bool
}

type handler4[T1, T2, T3, T4 any] struct {
	// callable is the object that will be invoked.
	callable      Callable4[T1, T2, T3, T4]
	callableValue reflect.Value
	// mutex ensures that the callable is only ever invoked sequentially.
	mutex             sync.Mutex
	subscriptionFlags SubscriptionFlag
}

func (h *handler4[T1, T2, T3, T4]) apply(args ...any) error {
	if len(args) != 4 {
		return fmt.Errorf("expected exactly 4 argument; %d provided", len(args))
	}
	return h.apply4(args[0].(T1), args[1].(T2), args[2].(T3), args[3].(T4))
}

func (h *handler4[T1, T2, T3, T4]) apply4(arg1 T1, arg2 T2, arg3 T3, arg4 T4) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.callable(arg1, arg2, arg3, arg4)
}

func (h *handler4[T1, T2, T3, T4]) getCallable() reflect.Value {
	return h.callableValue
}

func (h *handler4[T1, T2, T3, T4]) isOnce() bool {
	return h.subscriptionFlags&SubscriptionOnce == SubscriptionOnce
}

func (h *handler4[T1, T2, T3, T4]) isAsync() bool {
	return h.subscriptionFlags&SubscriptionAsync == SubscriptionAsync
}

func newHandler4[T1, T2, T3, T4 any](callable any, options ...SubscriptionModifier) (Handler4[T1, T2, T3, T4], error) {
	var specificCallable Callable4[T1, T2, T3, T4]
	switch v := callable.(type) {
	case Callable4[T1, T2, T3, T4]:
		specificCallable = v
	case Callable4NoError[T1, T2, T3, T4]:
		specificCallable = func(arg1 T1, arg2 T2, arg3 T3, arg4 T4) error {
			v(arg1, arg2, arg3, arg4)
			return nil
		}
	default:
		return nil, fmt.Errorf("The callable's parameter list doesn't match the event's generic type list: %v", reflect.TypeOf(callable))
	}
	h := &handler4[T1, T2, T3, T4]{
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

var _ Handler4[any, any, any, any] = (*handler4[any, any, any, any])(nil)
var _ Handler = (*handler4[any, any, any, any])(nil)
