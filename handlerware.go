package events

import (
	"fmt"
	"sync"
)

type Registration[T any] struct {
	event *Event
	data  T
}

type Registry[T any] struct {
	registry     map[EventName]*Registration[T]
	registryLock sync.RWMutex
}

func (r *Registry[T]) Registration(name EventName) (*Registration[T], error) {
	r.registryLock.RLock()
	defer r.registryLock.RUnlock()

	reg, ok := r.registry[name]
	if !ok {
		return nil, fmt.Errorf("event %s has not been registered", name)
	}
	return reg, nil
}

func (r *Registry[T]) Register(e *Event, data T) error {
	r.registryLock.Lock()
	defer r.registryLock.Unlock()

	if e == nil {
		return fmt.Errorf("event cannot be nil")
	}
	if e.N == "" {
		return fmt.Errorf("event's name cannot be blank")
	}
	if _, ok := r.registry[e.N]; ok {
		return fmt.Errorf("event %s has already been registered", e.N)
	}
	reg := &Registration[T]{
		event: e,
		data:  data,
	}
	r.registry[e.N] = reg
	return nil
}

func (r *Registry[T]) Deregister(e *Event) (T, error) {
	r.registryLock.Lock()
	defer r.registryLock.Unlock()

	var res T

	if e == nil {
		return res, fmt.Errorf("event cannot be nil")
	}
	if e.N == "" {
		return res, fmt.Errorf("event's name cannot be blank")
	}
	reg, ok := r.registry[e.N]
	if !ok {
		return res, fmt.Errorf("event %s has not been mounted", e.N)
	}
	delete(r.registry, e.N)
	return reg.data, nil
}
