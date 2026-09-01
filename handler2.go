package events

import (
	"fmt"
	"reflect"
	"sync"
)

type Callable2[T1, T2 any] = func(T1, T2) error

type Applicable2[T1, T2 any] interface {
	// apply2 invokes the callable with the exact argument(s).
	apply2(arg1 T1, arg2 T2) error
}

type handler2[T1, T2 any] struct {
	event             *E2[T1, T2]
	call              Callable2[T1, T2]
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

	isOnce := h.subscriptionFlags&SubscriptionOnce != 0
	isAsync := h.subscriptionFlags&SubscriptionAsync != 0

	if isOnce {
		h.event.handlersToRemove = append(h.event.handlersToRemove, h)
	}
	if len(h.event.handlerwares) == 0 {
		if isAsync {
			h.event.wg.Go(func() {
				h.call(arg1, arg2)
			})
		} else {
			h.call(arg1, arg2)
		}
	} else {
		for _, hw := range h.event.handlerwares {
			if err := hw.OnPreFire(h.event, h, arg1, arg2); err != nil {
				return err
			}
		}
		if isAsync {
			h.event.wg.Go(func() {
				h.call(arg1, arg2)
			})
		} else {
			h.call(arg1, arg2)
		}
		for _, hw := range h.event.handlerwares {
			if err := hw.OnPostFire(h.event, h, arg1, arg2); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *handler2[T1, T2]) callable() reflect.Value {
	return reflect.ValueOf(h.call)
}

func newHandler2[T1, T2 any](event *E2[T1, T2], callable Callable2[T1, T2], options ...SubscriptionModifier) (*handler2[T1, T2], error) {
	h := &handler2[T1, T2]{
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

var _ Applicable2[any, any] = (*handler2[any, any])(nil)
var _ Handler = (*handler2[any, any])(nil)
