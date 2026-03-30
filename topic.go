package EventBus

import (
	"fmt"
	"reflect"
	"slices"
	"sync"
)

// Topic manages listeners and dispatches event payloads to said listeners.
type Topic interface {
	// Fire dispatches the given payload(s) to all subscribed listeners taking into account the modifiers that they were
	// registered with. If a listener's function signature contains more parameters than provided arguments, zero values
	// will be filled in. If a listener's function contains less parameters than provided arguments, the listener will
	// be invoked will less arguments.
	Fire(args ...any)
	// Fireable returns true if at least one listener is registered, false otherwise.
	Fireable() bool
	// On registers the given listener with the given modifiers. Returns an error if `listener` is not a function.
	On(listener any, options ...SubscriptionModifier) error
	// Off cancels the given listener. Returns an error if `listener` is not subscribed to this topic.
	Off(listener any) error
	// WaitAsync waits for all registered async listeners of this Topic to complete.
	WaitAsync()
}

type topic struct {
	name      string
	listeners []*listener
	lock      sync.RWMutex
	wg        sync.WaitGroup
	wgBus     *sync.WaitGroup
}

func newTopic(name string, wgBus *sync.WaitGroup) Topic {
	return Topic(&topic{
		name:      name,
		listeners: []*listener{},
		lock:      sync.RWMutex{},
		wg:        sync.WaitGroup{},
		wgBus:     wgBus,
	})
}

func (t *topic) Fire(args ...any) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if len(t.listeners) == 0 {
		return
	}

	listenersToRemove := []*listener{}

	for _, listener := range t.listeners {
		if listener.once {
			listenersToRemove = append(listenersToRemove, listener)
		}
		if !listener.async {
			listener.apply(args...)
		} else {
			t.wg.Add(1)
			t.wgBus.Add(1)
			if listener.transactional {
				t.lock.RUnlock()
				listener.Lock()
				t.lock.RLock()
			}
			go func() {
				defer t.wg.Done()
				defer t.wgBus.Done()
				if listener.transactional {
					defer listener.Unlock()
				}
				listener.apply(args...)
			}()
		}
	}
	for _, listener := range listenersToRemove {
		t.removeListener(listener.listener)
	}
}

func (t *topic) Fireable() bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return len(t.listeners) > 0
}

func (t *topic) On(callable any, options ...SubscriptionModifier) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	listenerValue := reflect.ValueOf(callable)
	if kind := listenerValue.Kind(); kind != reflect.Func {
		return fmt.Errorf("%s is not of type reflect.Func", kind)
	}
	listenerType := listenerValue.Type()
	listenerNumIn := listenerType.NumIn()
	nilArgs := make([]reflect.Value, listenerNumIn)
	for i := range listenerNumIn {
		nilArgs[i] = reflect.New(listenerType.In(i)).Elem()
	}
	d := &listener{
		listener:      listenerValue,
		listenerNumIn: listenerNumIn,
		listenerArgs:  make([]reflect.Value, listenerNumIn),
		nilArgs:       nilArgs,
		Mutex:         sync.Mutex{},
		once:          false,
		async:         false,
		transactional: false,
	}
	for _, option := range options {
		option(d)
	}
	t.listeners = append(t.listeners, d)
	return nil
}

// Off removes a listener defined for a topic.
// Returns error if there are no listeners subscribed to the topic.
func (t *topic) Off(listener any) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if len(t.listeners) == 0 {
		return fmt.Errorf("topic %s doesn't have any listeners", t.name)
	}
	value := reflect.ValueOf(listener)
	err := t.removeListener(value)
	if err != nil {
		return fmt.Errorf("function %v is not subscribed to topic %s %w", listener, t.name, err)
	}
	return nil
}

// WaitAsync waits for all async listeners to complete
func (t *topic) WaitAsync() {
	t.wg.Wait()
}

func (t *topic) removeListener(listener reflect.Value) error {
	listenerIdx, _ := t.findListener(listener)
	if listenerIdx == -1 {
		return fmt.Errorf("listener %v not found", listener)
	}
	t.removeListenerIdx(listenerIdx)
	return nil
}

func (t *topic) findListener(callable reflect.Value) (int, *listener) {
	for i, listener := range t.listeners {
		if listener.listener.Pointer() != callable.Pointer() {
			continue
		}
		return i, listener
	}
	return -1, nil
}

func (t *topic) removeListenerIdx(idx int) error {
	numListeners := len(t.listeners)

	if idx < 0 || idx >= numListeners {
		return fmt.Errorf("listener index out of range: %v", idx)
	}

	t.listeners = slices.Delete(t.listeners, idx, 1)
	return nil
}
