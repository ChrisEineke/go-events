package events

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"sync"
)

type Fireable interface {
	// Fire dispatches the given payload(s) to all subscribed handlers taking into account the modifiers that they were
	// registered with. If a Handler's function signature contains more parameters than provided arguments, zero values
	// will be filled in. If a Handler's function contains less parameters than provided arguments, the Handler will
	// be invoked will less arguments.
	Fire(args ...any)
	// Fire dispatches the given payload(s) to all subscribed handlers taking into account the modifiers that they were
	// registered with and the provided deadline. If a Handler's function signature contains more parameters than
	// provided arguments, zero values will be filled in. If a Handler's function contains less parameters than
	// provided arguments, the Handler will be invoked will less arguments.
	FireContext(ctx context.Context, args ...any)
	// HasHandlers returns true if at least one Handler is registered, false otherwise.
	HasHandlers() bool
	// Use adds the Handlerware to this Event.
	Use(Handlerware)
	// Disuse emoves the Handlerware from this Event.
	Disuse(Handlerware)
}

type Subscribable interface {
	// On registers the given callable with the given modifiers. Returns an error if the callable is not a function.
	On(callable any, options ...SubscriptionModifier) error
	// Off cancels the given callable. Returns an error if the callable is not subscribed to this Event.
	Off(callable any) error
}

type Waitable interface {
	// WaitAsync waits for all registered async handlers of this Event to complete.
	WaitAsync()
}
type EventName = string

type Event struct {
	Fireable
	Waitable

	N                EventName
	handlers         []Handler
	handlersToRemove []Handler
	handlerwares     []Handlerware
	lock             sync.RWMutex
	wg               sync.WaitGroup
}

func (e *Event) Fire(args ...any) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	for _, hw := range e.handlerwares {
		hw.OnAllPreFire(e, args)
	}
	for _, handler := range e.handlers {
		if handler.isOnce() {
			e.handlersToRemove = append(e.handlersToRemove, handler)
		}
		if !handler.isAsync() {
			for _, hw := range e.handlerwares {
				hw.OnPreFire(e, handler, args)
			}
			handler.apply(args...)
			for _, hw := range e.handlerwares {
				hw.OnPostFire(e, handler, args)
			}
		} else {
			e.wg.Add(1)
			if handler.isTransactional() {
				e.lock.RUnlock()
				handler.Lock()
				e.lock.RLock()
			}
			go func() {
				defer e.wg.Done()
				if handler.isTransactional() {
					defer handler.Unlock()
				}
				for _, hw := range e.handlerwares {
					hw.OnPreFire(e, handler, args)
				}
				handler.apply(args...)
				for _, hw := range e.handlerwares {
					hw.OnPostFire(e, handler, args)
				}
			}()
		}
	}
	if len(e.handlersToRemove) > 0 {
		for _, handler := range e.handlersToRemove {
			e.removeCallable(handler.getCallable())
		}
		e.handlersToRemove = e.handlersToRemove[:0]
	}
	for _, hw := range e.handlerwares {
		hw.OnAllPostFire(e, args)
	}
}

func (e *Event) removeCallable(h reflect.Value) error {
	foundOne := false
	e.handlers = slices.DeleteFunc(e.handlers, func(it Handler) bool {
		if it.getCallable().Pointer() == h.Pointer() {
			if foundOne {
				return false
			}
			foundOne = true
			return true
		}
		return false
	})
	if !foundOne {
		return fmt.Errorf("handler %v not found", h)
	}
	return nil
}

func (e *Event) FireContext(ctx context.Context, args ...any) {
}

func (e *Event) HasHandlers() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()

	return len(e.handlers) > 0
}

func (e *Event) Use(hw Handlerware) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.handlerwares = append(e.handlerwares, hw)
	hw.OnUse(e)
}

func (e *Event) Disuse(hw Handlerware) {
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
		hw.OnDisuse(e)
	}
}

func (e *Event) On(callable any, options ...SubscriptionModifier) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	handler, err := newHandler(callable, options...)
	if err != nil {
		return err
	}
	e.handlers = append(e.handlers, handler)
	return nil
}

func (e *Event) Off(callable any) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	if len(e.handlers) == 0 {
		return fmt.Errorf("event doesn't have any handlers")
	}
	value := reflect.ValueOf(callable)
	err := e.removeCallable(value)
	if err != nil {
		return fmt.Errorf("function %v is not subscribed to event %w", callable, err)
	}
	return nil
}

func (e *Event) WaitAsync() {
	e.wg.Wait()
}
