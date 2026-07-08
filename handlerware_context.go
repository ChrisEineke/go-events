package events

import (
	"context"
	"fmt"
	"reflect"
)

// StrictContextWare adds Context-awareness to EventS.
//
// * All existing and new Handlers' first parameter is enforced to be of type context.Context,
//
// * Fire() enforces the first argument to be of type context.Context,
//
// * All current and future Handlers will be checked to have a context.Context as the first parameter when the
//   Handlerware is Use()'d with an Event.

type StrictContextWare struct {
	registry *Registry[context.Context]
}

// Creates a new StrictContextWare.
func NewStrictContextWare() *StrictContextWare {
	return &StrictContextWare{registry: NewRegistry[context.Context]()}
}

func (s *StrictContextWare) OnUse(e Event) error {
	var err error
	for _, handler := range e.Handlers() {
		err = s.ensureCallableFirstArgIsContext(handler.getCallable())
		if err != nil {
			break
		}
	}
	return err
}

func (s *StrictContextWare) OnDisuse(e Event) error {
	return nil
}

func (s *StrictContextWare) OnSubscribe(e Event, h Handler) error {
	return s.ensureCallableFirstArgIsContext(h.getCallable())
}

func (s *StrictContextWare) OnUnsubscribe(e Event, h Handler) error {
	return nil
}

func (s *StrictContextWare) OnAllPreFire(e Event, args []any) error {
	if len(args) == 0 {
		return fmt.Errorf("event payload must contain at least 1 arg (a Context)")
	}
	c, ok := args[0].(context.Context)
	if !ok {
		return fmt.Errorf("event payload's first argument is not a Context: %v", args[0])
	}
	return s.registry.Put(e, c)
}

func (s *StrictContextWare) OnPreFire(e Event, h Handler, args []any) error {
	reg, _ := s.registry.Get(e.Name())
	select {
	case <-reg.data.Done():
		return reg.data.Err()
	default:
		break
	}
	return nil
}

func (s *StrictContextWare) OnPostFire(e Event, h Handler, args []any) error {
	reg, _ := s.registry.Get(e.Name())
	select {
	case <-reg.data.Done():
		return reg.data.Err()
	default:
		break
	}
	return nil
}

func (s *StrictContextWare) OnAllPostFire(e Event, args []any) error {
	_, err := s.registry.Delete(e)
	return err
}

func (s *StrictContextWare) ensureCallableFirstArgIsContext(callable reflect.Value) error {
	callableType := callable.Type()
	callableNumIn := callableType.NumIn()
	if callableNumIn < 1 {
		return fmt.Errorf("handler must have at least 1 arg (a Context)")
	}
	firstArgType := callableType.In(0)
	if firstArgType != reflect.TypeFor[context.Context]() {
		return fmt.Errorf("handler's first arg is not a Context")
	}
	return nil
}

var _ Handlerware = (*StrictContextWare)(nil)
