package EventBus

import (
	"fmt"
	"reflect"
	"slices"
	"sync"
)

type Bus interface {
	Subscribe(topic string, callback any, options ...SubscribeOption) error
	Unsubscribe(topic string, callback any) error
	Publish(topic string, args ...any)
	HasCallback(topic string) bool
	WaitAsync()
}

// EventBus - box for dispatchers and callbacks.
type EventBus struct {
	registry map[string]*subscription
	lock     sync.RWMutex // a lock for the map
	wg       sync.WaitGroup
}

type subscription struct {
	dispatchers []*dispatcher
}

func (es *subscription) hasDispatchers() bool {
	return len(es.dispatchers) > 0
}

func (es *subscription) addDispatcher(d *dispatcher) {
	es.dispatchers = append(es.dispatchers, d)
}

func (es *subscription) removeDispatcher(callback reflect.Value) error {
	dispatcherIdx, _ := es.findDispatcher(callback)
	if dispatcherIdx == -1 {
		return fmt.Errorf("dispatcher %v not found", callback)
	}
	es.removeDispatcherIdx(dispatcherIdx)
	return nil
}

func (es *subscription) findDispatcher(callback reflect.Value) (int, *dispatcher) {
	for i, dispatcher := range es.dispatchers {
		if dispatcher.callback.Type() == callback.Type() &&
			dispatcher.callback.Pointer() == callback.Pointer() {
			return i, dispatcher
		}
	}
	return -1, nil
}

func (es *subscription) removeDispatcherIdx(idx int) error {
	numDispatchers := len(es.dispatchers)

	if idx < 0 || idx >= numDispatchers {
		return fmt.Errorf("dispatcher index out of range: %v", idx)
	}

	es.dispatchers = slices.Delete(es.dispatchers, idx, 1)
	return nil
}

type dispatcher struct {
	callback      reflect.Value
	callbackType  reflect.Type
	callbackArgs  []reflect.Value
	nilArgs       []reflect.Value
	once          bool
	async         bool
	transactional bool
	sync.Mutex    // lock for an event dispatcher - useful for running async callbacks serially
}

func (d *dispatcher) dispatch(args ...any) {
	for i := range d.callbackArgs {
		arg := args[i]
		if arg == nil {
			d.callbackArgs[i] = d.nilArgs[i]
		} else {
			d.callbackArgs[i] = reflect.ValueOf(arg)
		}
	}
	d.callback.Call(d.callbackArgs)
}

// New returns a new EventBus with no subscriptions.
func New() Bus {
	b := &EventBus{
		registry: map[string]*subscription{},
		lock:     sync.RWMutex{},
		wg:       sync.WaitGroup{},
	}
	return Bus(b)
}

type SubscribeOption func(*dispatcher)

// Async invokes the callback asynchronously.
func Async() SubscribeOption {
	return func(eh *dispatcher) {
		eh.async = true
	}
}

// Once removes the callback after being called once.
func Once() SubscribeOption {
	return func(eh *dispatcher) {
		eh.once = true
	}
}

// Transactional determines whether subsequent callbacks for a topic are run serially (true) or concurrently (false).
func Transactional() SubscribeOption {
	return func(eh *dispatcher) {
		eh.transactional = true
	}
}

// Subscribe subscribes a callback to a topic with the given subscription options:
// Returns error if the length of `topic` is 0.
// Returns error if `callback` is not a function.
func (bus *EventBus) Subscribe(topic string, callback any, options ...SubscribeOption) error {
	if len(topic) == 0 {
		return fmt.Errorf("cannot subscribe to empty topic name")
	}
	callbackValue := reflect.ValueOf(callback)
	if kind := callbackValue.Kind(); kind != reflect.Func {
		return fmt.Errorf("%s is not of type reflect.Func", kind)
	}
	callbackType := callbackValue.Type()
	nilArgs := make([]reflect.Value, callbackType.NumIn())
	for i := range callbackType.NumIn() {
		nilArgs[i] = reflect.New(callbackType.In(i)).Elem()
	}
	d := &dispatcher{
		callback:      callbackValue,
		callbackType:  callbackType,
		callbackArgs:  make([]reflect.Value, callbackType.NumIn()),
		nilArgs:       nilArgs,
		once:          false,
		async:         false,
		transactional: false,
		Mutex:         sync.Mutex{},
	}
	for _, option := range options {
		option(d)
	}

	bus.lock.Lock()
	defer bus.lock.Unlock()

	es, ok := bus.registry[topic]
	if ok {
		es.addDispatcher(d)
	} else {
		bus.registry[topic] = &subscription{[]*dispatcher{d}}
	}
	return nil
}

// Unsubscribe removes a callback defined for a topic.
// Returns error if there are no callbacks subscribed to the topic.
func (bus *EventBus) Unsubscribe(topic string, callback any) error {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	es, ok := bus.registry[topic]
	if !ok {
		return fmt.Errorf("topic %s doesn't exist", topic)
	}
	if len(es.dispatchers) == 0 {
		return fmt.Errorf("topic %s doesn't have any dispatchers", topic)
	}
	value := reflect.ValueOf(callback)
	err := es.removeDispatcher(value)
	if err != nil {
		return fmt.Errorf("function %v is not subscribed to topic %s %w", callback, topic, err)
	}
	return nil
}

// HasCallback returns true if exists any callback subscribed to the topic.
func (bus *EventBus) HasCallback(topic string) bool {
	bus.lock.RLock()
	defer bus.lock.RUnlock()

	if es, ok := bus.registry[topic]; ok {
		return es.hasDispatchers()
	}
	return false
}

// Publish executes all callback subscribed to a topic. Any additional argument will be transferred to the callback.
func (bus *EventBus) Publish(topic string, args ...any) {
	bus.lock.RLock() // will unlock if dispatcher is not found or always after setUpPublish
	defer bus.lock.RUnlock()

	es, ok := bus.registry[topic]
	if !ok || !es.hasDispatchers() {
		return
	}

	dispatchersToRemove := []*dispatcher{}

	for _, dispatcher := range es.dispatchers {
		if dispatcher.once {
			dispatchersToRemove = append(dispatchersToRemove, dispatcher)
		}
		if !dispatcher.async {
			dispatcher.dispatch(args...)
		} else {
			bus.wg.Add(1)
			if dispatcher.transactional {
				bus.lock.RUnlock()
				dispatcher.Lock()
				bus.lock.RLock()
			}
			go func() {
				defer bus.wg.Done()
				if dispatcher.transactional {
					defer dispatcher.Unlock()
				}
				dispatcher.dispatch(args...)
			}()
		}
	}
	for _, dispatcher := range dispatchersToRemove {
		es.removeDispatcher(dispatcher.callback)
	}
}

// WaitAsync waits for all async callbacks to complete
func (bus *EventBus) WaitAsync() {
	bus.wg.Wait()
}
