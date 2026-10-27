package events

type Handlerware interface {
	// OnUse will be called when the Handlerware is attached to the EventSource.
	OnUse(es EventSource) error
	// OnDisuse will be called when the Handlerware is detached from the EventSource.
	OnDisuse(e EventSource) error

	// OnSubscribe will be called after a hanlder was attached to the EventSource.
	OnSubscribe(e EventSource, h Handler) error
	// OnUnsubscribe will be called after handler was detached from the EventSource.
	OnUnsubscribe(e EventSource, h Handler) error

	// OnAllPreFire will be called before all regular handlers.
	OnAllPreFire(e EventSource, args ...any) error
	// OnPreFire will be called before a specific handler is called.
	OnPreFire(e EventSource, h Handler, args ...any) error

	// OnPostFire will be called after a specific handler is called.
	OnPostFire(e EventSource, h Handler, args ...any) error
	// OnAllPostFire will be called after all regular handlers have been called and the subscription list have been
	// updated.
	OnAllPostFire(e EventSource, args ...any) error
}
