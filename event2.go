package events

import (
	"fmt"
	"reflect"
	"slices"
	"sync"
)

type Event2[T1, T2 any] interface {
	// Fire dispatches the given payload(s) to all subscribed handlers taking into account the modifiers that they were
	// registered with. If a Handler's function signature contains more parameters than provided arguments, zero values
	// will be filled in. If a Handler's function contains less parameters than provided arguments, the Handler will
	// be invoked will less arguments.
	Fire2(arg1 T1, arg2 T2) error
	// HasHandlers returns true if at least one Handler is registered, false otherwise.
	HasHandlers() bool
	// Use adds the Handlerware to this Event.
	Use(Handlerware) error
	// Disuse emoves the Handlerware from this Event.
	Disuse(Handlerware) error
	// On registers the given callable with the given modifiers. Returns an error if the callable is not a function.
	On(callable any, options ...SubscriptionModifier) error
	// Off cancels the given callable. Returns an error if the callable is not subscribed to this Event.
	Off(callable any) error
	// WaitAsync waits for all registered async handlers of this Event to complete.
	WaitAsync()
}

// E2 is a an Event whose HandlerS receive exactly two arguments of the Event's generic types.
type E2[T1, T2 any] struct {
	N                EventName
	handlers         []Handler2[T1, T2]
	handlersToRemove []Handler2[T1, T2]
	handlerwares     []Handlerware
	lock             sync.RWMutex
	wg               sync.WaitGroup
}

func (e *E2[T1, T2]) Fire(args ...any) error {
	return e.Fire2(args[0].(T1), args[1].(T2))
}

func (e *E2[T1, T2]) Fire2(arg1 T1, arg2 T2) error {
	e.lock.RLock()
	defer e.lock.RUnlock()

	if len(e.handlerwares) == 0 {
		for _, handler := range e.handlers {
			if handler.isOnce() {
				e.handlersToRemove = append(e.handlersToRemove, handler)
			}
			if !handler.isAsync() {
				handler.apply2(arg1, arg2)
			} else {
				e.wg.Go(func() {
					handler.apply2(arg1, arg2)
				})
			}
		}
		if len(e.handlersToRemove) > 0 {
			for _, handler := range e.handlersToRemove {
				e.removeCallable(handler.getCallable())
			}
			e.handlersToRemove = e.handlersToRemove[:0]
		}
	} else {
		args := []any{arg1, arg2}

		for _, hw := range e.handlerwares {
			if err := hw.OnAllPreFire(e, args); err != nil {
				return err
			}
		}
		for _, handler := range e.handlers {
			if handler.isOnce() {
				e.handlersToRemove = append(e.handlersToRemove, handler)
			}
			if !handler.isAsync() {
				for _, hw := range e.handlerwares {
					if err := hw.OnPreFire(e, handler.(Handler), args); err != nil {
						return err
					}
				}
				handler.apply2(arg1, arg2)
				for _, hw := range e.handlerwares {
					if err := hw.OnPostFire(e, handler.(Handler), args); err != nil {
						return err
					}
				}
			} else {
				e.wg.Go(func() {
					for _, hw := range e.handlerwares {
						_ = hw.OnPreFire(e, handler.(Handler), args)
					}
					handler.apply2(arg1, arg2)
					for _, hw := range e.handlerwares {
						_ = hw.OnPostFire(e, handler.(Handler), args)
					}
				})
			}
		}
		if len(e.handlersToRemove) > 0 {
			for _, handler := range e.handlersToRemove {
				e.removeCallable(handler.getCallable())
			}
			e.handlersToRemove = e.handlersToRemove[:0]
		}
		for _, hw := range e.handlerwares {
			if err := hw.OnAllPostFire(e, args); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *E2[T1, T2]) removeCallable(h reflect.Value) (Handler2[T1, T2], error) {
	var result Handler2[T1, T2]
	e.handlers = slices.DeleteFunc(e.handlers, func(it Handler2[T1, T2]) bool {
		if it.getCallable().Pointer() == h.Pointer() {
			if result != nil {
				return false
			}
			result = it
			return true
		}
		return false
	})
	if result == nil {
		return nil, fmt.Errorf("handler %v not found", h)
	}
	return result, nil
}

func (e *E2[T1, T2]) HasHandlers() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()

	return len(e.handlers) > 0
}

func (e *E2[T1, T2]) Use(hw Handlerware) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.handlerwares = append(e.handlerwares, hw)
	return hw.OnUse(e)
}

func (e *E2[T1, T2]) Disuse(hw Handlerware) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	foundOne := false
	e.handlerwares = slices.DeleteFunc(e.handlerwares, func(it Handlerware) bool {
		if it == hw {
			if foundOne {
				return false
			}
			foundOne = true
			return true
		}
		return false
	})
	if foundOne {
		return hw.OnDisuse(e)
	}
	return nil
}

func (e *E2[T1, T2]) On(callable any, options ...SubscriptionModifier) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	handler, err := newHandler2[T1, T2](callable, options...)
	if err != nil {
		return err
	}
	e.handlers = append(e.handlers, handler)
	for _, hw := range e.handlerwares {
		if err := hw.OnSubscribe(e, handler.(Handler)); err != nil {
			return err
		}
	}
	return nil
}

func (e *E2[T1, T2]) Off(callable any) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	if len(e.handlers) == 0 {
		return fmt.Errorf("event doesn't have any handlers")
	}
	value := reflect.ValueOf(callable)
	handler, err := e.removeCallable(value)
	if err != nil {
		return fmt.Errorf("function %v is not subscribed to event %w", callable, err)
	}
	for _, hw := range e.handlerwares {
		hw.OnUnsubscribe(e, handler.(Handler))
	}
	return nil
}

func (e *E2[T1, T2]) WaitAsync() {
	e.wg.Wait()
}

func (e *E2[T1, T2]) Name() EventName {
	return e.N
}

func (e *E2[T1, T2]) Handlers() []Handler {
	var result []Handler
	for _, handler := range e.handlers {
		result = append(result, handler.(Handler))
	}
	return result
}

var _ Event2[any, any] = (*E2[any, any])(nil)
var _ Event = (*E2[any, any])(nil)
