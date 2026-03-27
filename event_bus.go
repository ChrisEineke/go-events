package EventBus

import (
	"fmt"
	"reflect"
	"slices"
	"sync"
)

// BusSubscriber defines subscription-related bus behavior
type BusSubscriber interface {
	Subscribe(topic string, fn any) error
	SubscribeAsync(topic string, fn any, transactional bool) error
	SubscribeOnce(topic string, fn any) error
	SubscribeOnceAsync(topic string, fn any) error
	Unsubscribe(topic string, handler any) error
}

// BusPublisher defines publishing-related bus behavior
type BusPublisher interface {
	Publish(topic string, args ...any)
}

// BusController defines bus control behavior (checking handler's presence, synchronization)
type BusController interface {
	HasCallback(topic string) bool
	WaitAsync()
}

// Bus englobes global (subscribe, publish, control) bus behavior
type Bus interface {
	BusController
	BusSubscriber
	BusPublisher
}

// EventBus - box for handlers and callbacks.
type EventBus struct {
	registry map[string]*eventSubscription
	lock     sync.RWMutex // a lock for the map
	wg       sync.WaitGroup
}

type eventSubscription struct {
	handlers []*eventHandler
}

func (es *eventSubscription) hasHandlers() bool {
	return len(es.handlers) > 0
}

func (es *eventSubscription) addHandler(handler *eventHandler) {
	es.handlers = append(es.handlers, handler)
}

func (es *eventSubscription) removeHandler(callback reflect.Value) error {
	handlerIdx, _ := es.findHandler(callback)
	if handlerIdx == -1 {
		return fmt.Errorf("handler %v not found", callback)
	}
	es.removeHandlerIdx(handlerIdx)
	return nil
}

func (es *eventSubscription) findHandler(callback reflect.Value) (int, *eventHandler) {
	for i, handler := range es.handlers {
		if handler.callback.Type() == callback.Type() &&
			handler.callback.Pointer() == callback.Pointer() {
			return i, handler
		}
	}
	return -1, nil
}

func (es *eventSubscription) removeHandlerIdx(idx int) error {
	numHandlers := len(es.handlers)

	if idx < 0 || idx >= numHandlers {
		return fmt.Errorf("handler index out of range: %v", idx)
	}

	es.handlers = slices.Delete(es.handlers, idx, 1)
	return nil
}

type eventHandler struct {
	callback      reflect.Value
	flagOnce      bool
	async         bool
	transactional bool
	sync.Mutex    // lock for an event handler - useful for running async callbacks serially
}

// New returns new EventBus with empty handlers.
func New() Bus {
	b := &EventBus{
		registry: map[string]*eventSubscription{},
		lock:     sync.RWMutex{},
		wg:       sync.WaitGroup{},
	}
	return Bus(b)
}

// doSubscribe handles the subscription logic and is utilized by the public Subscribe functions
func (bus *EventBus) doSubscribe(topic string, handler *eventHandler) error {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	if kind := handler.callback.Kind(); kind != reflect.Func {
		return fmt.Errorf("%s is not of type reflect.Func", kind)
	}
	es, ok := bus.registry[topic]
	if ok {
		es.addHandler(handler)
	} else {
		bus.registry[topic] = &eventSubscription{[]*eventHandler{handler}}
	}
	return nil
}

// Subscribe subscribes to a topic.
// Returns error if `fn` is not a function.
func (bus *EventBus) Subscribe(topic string, fn any) error {
	return bus.doSubscribe(
		topic,
		&eventHandler{reflect.ValueOf(fn), false, false, false, sync.Mutex{}})
}

// SubscribeAsync subscribes to a topic with an asynchronous callback
// Transactional determines whether subsequent callbacks for a topic are
// run serially (true) or concurrently (false)
// Returns error if `fn` is not a function.
func (bus *EventBus) SubscribeAsync(topic string, fn any, transactional bool) error {
	return bus.doSubscribe(
		topic,
		&eventHandler{reflect.ValueOf(fn), false, true, transactional, sync.Mutex{}})
}

// SubscribeOnce subscribes to a topic once. Handler will be removed after executing.
// Returns error if `fn` is not a function.
func (bus *EventBus) SubscribeOnce(topic string, fn any) error {
	return bus.doSubscribe(
		topic,
		&eventHandler{reflect.ValueOf(fn), true, false, false, sync.Mutex{}})
}

// SubscribeOnceAsync subscribes to a topic once with an asynchronous callback
// Handler will be removed after executing.
// Returns error if `fn` is not a function.
func (bus *EventBus) SubscribeOnceAsync(topic string, fn any) error {
	return bus.doSubscribe(
		topic,
		&eventHandler{reflect.ValueOf(fn), true, true, false, sync.Mutex{}})
}

// HasCallback returns true if exists any callback subscribed to the topic.
func (bus *EventBus) HasCallback(topic string) bool {
	bus.lock.RLock()
	defer bus.lock.RUnlock()

	if es, ok := bus.registry[topic]; ok {
		return es.hasHandlers()
	}
	return false
}

// Unsubscribe removes callback defined for a topic.
// Returns error if there are no callbacks subscribed to the topic.
func (bus *EventBus) Unsubscribe(topic string, handler any) error {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	es, ok := bus.registry[topic]
	if !ok {
		return fmt.Errorf("topic %s doesn't exist", topic)
	}
	if len(es.handlers) == 0 {
		return fmt.Errorf("topic %s doesn't have any handlers", topic)
	}
	callback := reflect.ValueOf(handler)
	err := es.removeHandler(callback)
	if err != nil {
		return fmt.Errorf("handler %v is not subscribed to topic %s %w", callback, topic, err)
	}
	return nil
}

// Publish executes callback defined for a topic. Any additional argument will be transferred to the callback.
func (bus *EventBus) Publish(topic string, args ...any) {
	bus.lock.RLock() // will unlock if handler is not found or always after setUpPublish
	defer bus.lock.RUnlock()

	es, ok := bus.registry[topic]
	if !ok || !es.hasHandlers() {
		return
	}

	handlersToRemove := []*eventHandler{}

	for _, handler := range es.handlers {
		if handler.flagOnce {
			handlersToRemove = append(handlersToRemove, handler)
		}
		if !handler.async {
			bus.doPublish(handler, args...)
		} else {
			bus.wg.Add(1)
			if handler.transactional {
				bus.lock.RUnlock()
				handler.Lock()
				bus.lock.RLock()
			}
			go bus.doPublishAsync(handler, args...)
		}
	}
	for _, handler := range handlersToRemove {
		es.removeHandler(handler.callback)
	}
}

func (bus *EventBus) doPublish(handler *eventHandler, args ...any) {
	passedArguments := bus.setUpPublish(handler, args...)
	handler.callback.Call(passedArguments)
}

func (bus *EventBus) doPublishAsync(handler *eventHandler, args ...any) {
	defer bus.wg.Done()
	if handler.transactional {
		defer handler.Unlock()
	}
	bus.doPublish(handler, args...)
}

func (bus *EventBus) setUpPublish(callback *eventHandler, args ...any) []reflect.Value {
	funcType := callback.callback.Type()
	passedArguments := make([]reflect.Value, len(args))
	for i, v := range args {
		if v == nil {
			passedArguments[i] = reflect.New(funcType.In(i)).Elem()
		} else {
			passedArguments[i] = reflect.ValueOf(v)
		}
	}

	return passedArguments
}

// WaitAsync waits for all async callbacks to complete
func (bus *EventBus) WaitAsync() {
	bus.wg.Wait()
}
