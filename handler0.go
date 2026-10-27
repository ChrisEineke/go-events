package events

import (
	"fmt"
	"reflect"
	"sync"
)

type Func0 = func() error

type Applicable0 interface {
	// apply0 invokes the function with no payload.
	apply0() error
}

type handler0 struct {
	event             *E0
	fn                Func0
	mutex             sync.Mutex
	subscriptionFlags SubscriptionFlag
}

func (h *handler0) apply(args ...any) error {
	if len(args) != 0 {
		return fmt.Errorf("expected exactly 0 arguments; %d provided", len(args))
	}
	return h.apply0()
}

func (h *handler0) apply0() error {
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
				h.fn()
			})
		} else {
			h.fn()
		}
	} else {
		for _, hw := range h.event.handlerwares {
			if err := hw.OnPreFire(h.event, h); err != nil {
				return err
			}
		}
		if isAsync {
			h.event.wg.Go(func() {
				h.fn()
			})
		} else {
			h.fn()
		}
		for _, hw := range h.event.handlerwares {
			if err := hw.OnPostFire(h.event, h); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *handler0) funcValue() reflect.Value {
	return reflect.ValueOf(h.fn)
}

func newHandler0(event *E0, fn Func0, options ...SubscriptionModifier) (*handler0, error) {
	h := &handler0{
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

var _ Applicable0 = (*handler0)(nil)
var _ Handler = (*handler0)(nil)
