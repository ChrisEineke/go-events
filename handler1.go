package events

import (
	"fmt"
	"reflect"
	"sync"
)

type Func1[T1 any] = func(T1) error

type Applicable1[T1 any] interface {
	// apply1 invokes the function with the given payload.
	apply1(arg1 T1) error
}

type handler1[T1 any] struct {
	event             *E1[T1]
	fn                Func1[T1]
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

	isOnce := h.subscriptionFlags&SubscriptionOnce != 0
	isAsync := h.subscriptionFlags&SubscriptionAsync != 0

	if isOnce {
		h.event.handlersToRemove = append(h.event.handlersToRemove, h)
	}
	if len(h.event.handlerwares) == 0 {
		if isAsync {
			h.event.wg.Go(func() {
				h.fn(arg1)
			})
		} else {
			h.fn(arg1)
		}
	} else {
		for _, hw := range h.event.handlerwares {
			if err := hw.OnPreFire(h.event, h, arg1); err != nil {
				return err
			}
		}
		if isAsync {
			h.event.wg.Go(func() {
				h.fn(arg1)
			})
		} else {
			h.fn(arg1)
		}
		for _, hw := range h.event.handlerwares {
			if err := hw.OnPostFire(h.event, h, arg1); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *handler1[T1]) funcValue() reflect.Value {
	return reflect.ValueOf(h.fn)
}

func newHandler1[T1 any](event *E1[T1], fn Func1[T1], options ...SubscriptionModifier) (*handler1[T1], error) {
	h := &handler1[T1]{
		event:             event,
		fn:                fn,
		mutex:             sync.Mutex{},
		subscriptionFlags: 0,
	}
	for _, option := range options {
		option(&h.subscriptionFlags)
	}
	return h, nil
}

var _ Applicable1[any] = (*handler1[any])(nil)
var _ Handler = (*handler1[any])(nil)
