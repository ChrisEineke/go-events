package events

import (
	"fmt"
	"reflect"
	"sync"
)

type Callable4[T1, T2, T3, T4 any] = func(T1, T2, T3, T4) error

type Applicable4[T1, T2, T3, T4 any] interface {
	// apply4 invokes the callable with the exact argument(s).
	apply4(arg1 T1, arg2 T2, arg3 T3, arg4 T4) error
}

type handler4[T1, T2, T3, T4 any] struct {
	event             *E4[T1, T2, T3, T4]
	call              Callable4[T1, T2, T3, T4]
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

	isOnce := h.subscriptionFlags&SubscriptionOnce != 0
	isAsync := h.subscriptionFlags&SubscriptionAsync != 0

	if isOnce {
		h.event.handlersToRemove = append(h.event.handlersToRemove, h)
	}
	if len(h.event.handlerwares) == 0 {
		if isAsync {
			h.event.wg.Go(func() {
				h.call(arg1, arg2, arg3, arg4)
			})
		} else {
			h.call(arg1, arg2, arg3, arg4)
		}
	} else {
		for _, hw := range h.event.handlerwares {
			if err := hw.OnPreFire(h.event, h, arg1, arg2, arg3, arg4); err != nil {
				return err
			}
		}
		if isAsync {
			h.event.wg.Go(func() {
				h.call(arg1, arg2, arg3, arg4)
			})
		} else {
			h.call(arg1, arg2, arg3, arg4)
		}
		for _, hw := range h.event.handlerwares {
			if err := hw.OnPostFire(h.event, h, arg1, arg2, arg3, arg4); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *handler4[T1, T2, T3, T4]) callable() reflect.Value {
	return reflect.ValueOf(h.call)
}

func newHandler4[T1, T2, T3, T4 any](event *E4[T1, T2, T3, T4], callable Callable4[T1, T2, T3, T4], options ...SubscriptionModifier) (*handler4[T1, T2, T3, T4], error) {
	h := &handler4[T1, T2, T3, T4]{
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

var _ Applicable4[any, any, any, any] = (*handler4[any, any, any, any])(nil)
var _ Handler = (*handler4[any, any, any, any])(nil)
