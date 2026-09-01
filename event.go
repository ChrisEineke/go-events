package events

import (
	"fmt"
	"reflect"
	"slices"
	"sync"
)

type Event interface {
	// Fire dispatches the given payload(s) to all subscribed Handlers taking into account the modifiers that they were
	// registered with.
	// If there are no Handlers registered, Fire will return nil or an error depending on the AllowNoHandlers flag.
	// If there are Handlers registered, Fire will iterate through them and call them with the given arguments using the
	// following semantics:
	// - If a Handler's function signature contains more parameters than provided arguments, zero values will be filled
	// in.
	// - If a Handler's function contains less parameters than provided arguments, the Handler will be invoked will less
	// arguments.
	// Any errors returned by Handlers are currently ignored.
	// Any errors returned by Handlerwares will be returned immediately and iteration will stop.
	// If a Handler is registered with the Once modifier, it will be removed after being invoked.
	// If a Handler is registered with the Async modifier, it will be invoked asynchronously. The WaitAsync method can
	// be used to wait for all async Handlers to complete.
	// It is safe to call Fire from multiple Goroutines.
	Fire(args ...any) error

	// Use adds the Handlerware to this Event.
	Use(Handlerware) error

	// Disuse removes the Handlerware from this Event.
	Disuse(Handlerware) error

	// On registers the given callable with the given modifiers. Returns an error if the callable is not a function.
	On(callable any, options ...SubscriptionModifier) error

	// Off cancels the given callable. Returns an error if the callable is not subscribed to this Event.
	Off(callable any) error

	// WaitAsync waits for all registered async Handlers of this Event to complete.
	WaitAsync()

	// Name returns the name of the Event.
	Name() EventName
}

type EventName = string

// E is an Event.
type E struct {
	N                  EventName
	disallowNoHandlers bool
	handlers           []Handler
	handlersToRemove   []Handler
	handlerwares       []Handlerware
	lock               sync.Mutex
	wg                 sync.WaitGroup
}

func (e *E) Fire(args ...any) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	if len(e.handlers) == 0 && e.disallowNoHandlers {
		return ErrNoHandlers
	}

	for _, hw := range e.handlerwares {
		if err := hw.OnAllPreFire(e, args...); err != nil {
			return err
		}
	}
	for _, handler := range e.handlers {
		handler.apply(args...)
	}
	for _, hw := range e.handlerwares {
		if err := hw.OnAllPostFire(e, args...); err != nil {
			return err
		}
	}
	if len(e.handlersToRemove) > 0 {
		for _, handler := range e.handlersToRemove {
			e.removeCallable(handler.callable())
		}
		e.handlersToRemove = e.handlersToRemove[:0]
	}

	return nil
}

func (e *E) Use(hw Handlerware) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.handlerwares = append(e.handlerwares, hw)
	return hw.OnUse(e)
}

func (e *E) Disuse(hw Handlerware) error {
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

func (e *E) On(callable any, options ...SubscriptionModifier) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	handler, err := newHandler(e, callable, options...)
	if err != nil {
		return err
	}
	e.handlers = append(e.handlers, handler)
	for _, hw := range e.handlerwares {
		if err := hw.OnSubscribe(e, handler); err != nil {
			return err
		}
	}
	return nil
}

func (e *E) Off(callable any) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	if len(e.handlers) == 0 {
		return ErrNoHandlers
	}
	value := reflect.ValueOf(callable)
	handler, err := e.removeCallable(value)
	if err != nil {
		return fmt.Errorf("function %v is not subscribed to event %w", callable, err)
	}
	for _, hw := range e.handlerwares {
		hw.OnUnsubscribe(e, handler)
	}
	return nil
}

func (e *E) WaitAsync() {
	e.wg.Wait()
}

func (e *E) Name() EventName {
	return e.N
}

func (e *E) HasHandlers() bool {
	return len(e.handlers) > 0
}

func (e *E) Handlers() []Handler {
	return e.handlers
}

func (e *E) removeCallable(h reflect.Value) (Handler, error) {
	var result Handler
	e.handlers = slices.DeleteFunc(e.handlers, func(it Handler) bool {
		if it.callable().Pointer() == h.Pointer() {
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

// ensure that E implements EventSource and Event.
var _ EventSource = (*E)(nil)
var _ Event = (*E)(nil)
