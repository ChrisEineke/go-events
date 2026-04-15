package EventBus

import (
	"reflect"
	"sync"
)

type listener struct {
	listener      reflect.Value
	listenerNumIn int
	listenerArgs  []reflect.Value
	nilArgs       []reflect.Value
	sync.Mutex
	once          bool
	async         bool
	transactional bool
}

func (d *listener) apply(args ...any) {
	if d.listenerNumIn > 0 {
		d.listenerArgs = d.nilArgs
		for i := range d.listenerArgs {
			if args[i] == nil {
				continue
			}
			d.listenerArgs[i] = reflect.ValueOf(args[i])
		}
	}
	d.listener.Call(d.listenerArgs)
}

type SubscriptionModifier func(*listener)

// sync invokes the listener synchronously (the default).
func Sync() SubscriptionModifier {
	return func(eh *listener) {
		eh.async = false
	}
}

// Async invokes the listener asynchronously.
func Async() SubscriptionModifier {
	return func(eh *listener) {
		eh.async = true
	}
}

// Always keeps the listener registered after being called (the default).
func Always() SubscriptionModifier {
	return func(eh *listener) {
		eh.once = false
	}
}

// Once removes the listener after being called once.
func Once() SubscriptionModifier {
	return func(eh *listener) {
		eh.once = true
	}
}

// NonTransactional invokes subsequent listeners for a topic concurrently (the default).
func NonTransactional() SubscriptionModifier {
	return func(eh *listener) {
		eh.transactional = false
	}
}

// Transactional invokes subsequent listeners for a topic serially (true).
func Transactional() SubscriptionModifier {
	return func(eh *listener) {
		eh.transactional = true
	}
}
