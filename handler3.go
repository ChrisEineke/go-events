package events

import (
	"fmt"
	"reflect"
	"sync"
)

type Callable3[T1, T2, T3 any] = func(T1, T2, T3) error

type Applicable3[T1, T2, T3 any] interface {
	// apply3 invokes the callable with the exact argument(s).
	apply3(arg1 T1, arg2 T2, arg3 T3) error
}

type handler3[T1, T2, T3 any] struct {
	event             *E3[T1, T2, T3]
	call              Callable3[T1, T2, T3]
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

	isOnce := h.subscriptionFlags&SubscriptionOnce != 0
	isAsync := h.subscriptionFlags&SubscriptionAsync != 0

	if isOnce {
		h.event.handlersToRemove = append(h.event.handlersToRemove, h)
	}
	if len(h.event.handlerwares) == 0 {
		if isAsync {
			h.event.wg.Go(func() {
				h.call(arg1, arg2, arg3)
			})
		} else {
			h.call(arg1, arg2, arg3)
		}
	} else {
		for _, hw := range h.event.handlerwares {
			if err := hw.OnPreFire(h.event, h, arg1, arg2, arg3); err != nil {
				return err
			}
		}
		if isAsync {
			h.event.wg.Go(func() {
				h.call(arg1, arg2, arg3)
			})
		} else {
			h.call(arg1, arg2, arg3)
		}
		for _, hw := range h.event.handlerwares {
			if err := hw.OnPostFire(h.event, h, arg1, arg2, arg3); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *handler3[T1, T2, T3]) callable() reflect.Value {
	return reflect.ValueOf(h.call)
}

func newHandler3[T1, T2, T3 any](event *E3[T1, T2, T3], callable Callable3[T1, T2, T3], options ...SubscriptionModifier) (*handler3[T1, T2, T3], error) {
	h := &handler3[T1, T2, T3]{
		event:             event,
		call:              callable,
		mutex:             sync.Mutex{},
		subscriptionFlags: 0,
	}
	for _, option := range options {
		option(&h.subscriptionFlags)
	}
	return h, nil
}

var _ Applicable3[any, any, any] = (*handler3[any, any, any])(nil)
var _ Handler = (*handler3[any, any, any])(nil)
