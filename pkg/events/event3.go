package events

import (
	"fmt"
	"reflect"
	"slices"
	"sync"
)

type Event3[T1, T2, T3 any] interface {
	// Fire dispatches the given payload(s) to all subscribed handlers taking into account the modifiers that they were
	// registered with. If a Handler's function signature contains more parameters than provided arguments, zero values
	// will be filled in. If a Handler's function contains less parameters than provided arguments, the Handler will
	// be invoked will less arguments.
	Fire3(arg1 T1, arg2 T2, any T3) error
	// HasHandlers returns true if at least one Handler is registered, false otherwise.
	HasHandlers() bool
	// Use adds the Handlerware to this Event.
	Use(Handlerware) error
	// Disuse emoves the Handlerware from this Event.
	Disuse(Handlerware) error
	// On registers the given function with the given modifiers. Returns an error if the function is not a function.
	On(fn Func3[T1, T2, T3], options ...SubscriptionModifier) error
	// Off cancels the given function. Returns an error if the function is not subscribed to this Event.
	Off(fn Func3[T1, T2, T3]) error
	// WaitAsync waits for all registered async handlers of this Event to complete.
	WaitAsync()
}

// E3 is a an Event whose HandlerS receive exactly three arguments of the Event's generic types.
type E3[T1, T2, T3 any] struct {
	N                EventName
	handlers         []*handler3[T1, T2, T3]
	handlersToRemove []*handler3[T1, T2, T3]
	handlerwares     []Handlerware
	lock             sync.Mutex
	wg               sync.WaitGroup
}

func (e *E3[T1, T2, T3]) Fire(args ...any) error {
	return e.Fire3(args[0].(T1), args[1].(T2), args[2].(T3))
}

func (e *E3[T1, T2, T3]) Fire3(arg1 T1, arg2 T2, arg3 T3) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	for _, hw := range e.handlerwares {
		if err := hw.OnAllPreFire(e, arg1, arg2, arg3); err != nil {
			return err
		}
	}
	for _, handler := range e.handlers {
		handler.apply3(arg1, arg2, arg3)
	}
	for _, hw := range e.handlerwares {
		if err := hw.OnAllPostFire(e, arg1, arg2, arg3); err != nil {
			return err
		}
	}
	if len(e.handlersToRemove) > 0 {
		for _, handler := range e.handlersToRemove {
			e.removeHandlerByFuncValue(handler.funcValue())
		}
		e.handlersToRemove = e.handlersToRemove[:0]
	}

	return nil
}

func (e *E3[T1, T2, T3]) removeHandlerByFuncValue(v reflect.Value) (*handler3[T1, T2, T3], error) {
	var result *handler3[T1, T2, T3]
	e.handlers = slices.DeleteFunc(e.handlers, func(it *handler3[T1, T2, T3]) bool {
		if it.funcValue().Pointer() == v.Pointer() {
			if result != nil {
				return false
			}
			result = it
			return true
		}
		return false
	})
	if result == nil {
		return nil, fmt.Errorf("handler %v not found", v)
	}
	return result, nil
}

func (e *E3[T1, T2, T3]) HasHandlers() bool {
	e.lock.Lock()
	defer e.lock.Unlock()

	return len(e.handlers) > 0
}

func (e *E3[T1, T2, T3]) Use(hw Handlerware) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.handlerwares = append(e.handlerwares, hw)
	return hw.OnUse(e)
}

func (e *E3[T1, T2, T3]) Disuse(hw Handlerware) error {
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

func (e *E3[T1, T2, T3]) On(fn Func3[T1, T2, T3], options ...SubscriptionModifier) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	handler, err := newHandler3(e, fn, options...)
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

func (e *E3[T1, T2, T3]) Off(fn Func3[T1, T2, T3]) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	if len(e.handlers) == 0 {
		return fmt.Errorf("event doesn't have any handlers")
	}
	handler, err := e.removeHandlerByFuncValue(reflect.ValueOf(fn))
	if err != nil {
		return fmt.Errorf("function %v is not subscribed to event %w", fn, err)
	}
	for _, hw := range e.handlerwares {
		hw.OnUnsubscribe(e, handler)
	}
	return nil
}

func (e *E3[T1, T2, T3]) WaitAsync() {
	e.wg.Wait()
}

func (e *E3[T1, T2, T3]) Name() EventName {
	return e.N
}

func (e *E3[T1, T2, T3]) Handlers() []Handler {
	var result []Handler
	for _, handler := range e.handlers {
		result = append(result, handler)
	}
	return result
}

var _ EventSource = (*E3[any, any, any])(nil)
var _ Event3[any, any, any] = (*E3[any, any, any])(nil)
