package events

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"sync"
)

type Fireable interface {
	// Fire dispatches the given payload(s) to all subscribed listeners taking into account the modifiers that they were
	// registered with. If a Listener's function signature contains more parameters than provided arguments, zero values
	// will be filled in. If a Listener's function contains less parameters than provided arguments, the Listener will
	// be invoked will less arguments.
	Fire(args ...any)
	// Fire dispatches the given payload(s) to all subscribed listeners taking into account the modifiers that they were
	// registered with and the provided deadline. If a Listener's function signature contains more parameters than
	// provided arguments, zero values will be filled in. If a Listener's function contains less parameters than
	// provided arguments, the Listener will be invoked will less arguments.
	FireContext(ctx context.Context, args ...any)
	// HasListeners returns true if at least one Listener is registered, false otherwise.
	HasListeners() bool
	// Use adds the Handlerware to this Event.
	Use(Handlerware)
	// Disuse emoves the Handlerware from this Event.
	Disuse(Handlerware)
}

type Subscribable interface {
	// On registers the given callable with the given modifiers. Returns an error if the callable is not a function.
	On(callable any, options ...SubscriptionModifier) error
	// Off cancels the given callable. Returns an error if the callable is not subscribed to this topic.
	Off(callable any) error
}

type Waitable interface {
	// WaitAsync waits for all registered async listeners of this Topic to complete.
	WaitAsync()
}
type EventName = string

type Event struct {
	Fireable
	Waitable

	N                 EventName
	listeners         []Listener
	listenersToRemove []Listener
	handlerwares      []Handlerware
	lock              sync.RWMutex
	wg                sync.WaitGroup
}

func (e *Event) Fire(args ...any) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	for _, hw := range e.handlerwares {
		hw.OnAllPreFire(e, args)
	}
	for _, listener := range e.listeners {
		if listener.isOnce() {
			e.listenersToRemove = append(e.listenersToRemove, listener)
		}
		if !listener.isAsync() {
			for _, hw := range e.handlerwares {
				hw.OnPreFire(e, listener, args)
			}
			listener.apply(args...)
			for _, hw := range e.handlerwares {
				hw.OnPostFire(e, listener, args)
			}
		} else {
			e.wg.Add(1)
			if listener.isTransactional() {
				e.lock.RUnlock()
				listener.Lock()
				e.lock.RLock()
			}
			go func() {
				defer e.wg.Done()
				if listener.isTransactional() {
					defer listener.Unlock()
				}
				for _, hw := range e.handlerwares {
					hw.OnPreFire(e, listener, args)
				}
				listener.apply(args...)
				for _, hw := range e.handlerwares {
					hw.OnPostFire(e, listener, args)
				}
			}()
		}
	}
	if len(e.listenersToRemove) > 0 {
		for _, listener := range e.listenersToRemove {
			e.removeListener(listener.getListener())
		}
		e.listenersToRemove = e.listenersToRemove[:0]
	}
	for _, hw := range e.handlerwares {
		hw.OnAllPostFire(e, args)
	}
}

func (e *Event) removeListener(l reflect.Value) error {
	foundOne := false
	e.listeners = slices.DeleteFunc(e.listeners, func(it Listener) bool {
		if it.getListener().Pointer() == l.Pointer() {
			if foundOne {
				return false
			}
			foundOne = true
			return true
		}
		return false
	})
	if !foundOne {
		return fmt.Errorf("listener %v not found", l)
	}
	return nil
}

func (e *Event) FireContext(ctx context.Context, args ...any) {
}

func (e *Event) HasListeners() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()

	return len(e.listeners) > 0
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

	listener, err := newListener(callable, options...)
	if err != nil {
		return err
	}
	e.listeners = append(e.listeners, listener)
	return nil
}

func (e *Event) Off(callable any) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	if len(e.listeners) == 0 {
		return fmt.Errorf("event doesn't have any listeners")
	}
	value := reflect.ValueOf(callable)
	err := e.removeListener(value)
	if err != nil {
		return fmt.Errorf("function %v is not subscribed to event %w", callable, err)
	}
	return nil
}

func (e *Event) WaitAsync() {
	e.wg.Wait()
}
