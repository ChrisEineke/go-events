package events

import (
	"fmt"
	"reflect"
	"sync"
)

type Callable1[T1 any] = func(T1) error
type Callable1NoError[T1 any] = func(T1)

type Handler1[T1 any] interface {
	// apply invokes the callable with the given arguments. This variant of apply tries to match as many arguments of
	// the event payload to the parameter list of the callable (in order as fired only). The callable will not be
	// invoked with more parameters than it supports. If the callable has too many arguments, the remaining parameters
	// will be invoked with the parameters' zero values.
	apply(args ...any) error
	// apply1 invokes the callable with the exact argument(s).
	apply1(arg1 T1) error
	// getCallable returns the callable Value.
	getCallable() reflect.Value
	// isOnce returns whether or not this Handler is to be invoked only once and then removed from the handler list.
	isOnce() bool
	// isOnce returns whether or not this Handler is to be invoked asynchronously.
	isAsync() bool
}

type handler1[T1 any] struct {
	// callable is the object that will be invoked.
	callable      Callable1[T1]
	callableValue reflect.Value

	// mutex ensures that the callable is only ever invoked sequentially.
	mutex             sync.Mutex
	subscriptionFlags SubscriptionFlag
}

func (h *handler1[T1]) apply(args ...any) error {
	if len(args) != 1 {
		return fmt.Errorf("expected exactly 1 argument; %d provided", len(args))
	}
	return h.apply1(args[0].(T1))
}

func (h *handler1[T1]) apply1(arg1 T1) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.callable(arg1)
}

func (h *handler1[T1]) getCallable() reflect.Value {
	return h.callableValue
}

func (h *handler1[T1]) isOnce() bool {
	return h.subscriptionFlags&SubscriptionOnce == SubscriptionOnce
}

func (h *handler1[T1]) isAsync() bool {
	return h.subscriptionFlags&SubscriptionAsync == SubscriptionAsync
}

func newHandler1[T1 any](callable any, options ...SubscriptionModifier) (Handler1[T1], error) {
	var specificCallable Callable1[T1]
	switch v := callable.(type) {
	case Callable1[T1]:
		specificCallable = v
	case Callable1NoError[T1]:
		specificCallable = func(arg1 T1) error {
			v(arg1)
			return nil
		}
	default:
		return nil, fmt.Errorf("The callable's parameter list doesn't match the event's generic type list: %v", reflect.TypeOf(callable))
	}
	h := &handler1[T1]{
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

var _ Handler1[any] = (*handler1[any])(nil)
var _ Handler = (*handler1[any])(nil)
