package events

type Handlerware interface {
	// OnUse will be called when the Handlerware is attached to the Event.
	OnUse(e *E) error
	// OnDisuse will be called when the Handlerware is detached from the Event.
	OnDisuse(e *E) error

	// OnSubscribe will be called after a hanlder was attached to the Event.
	OnSubscribe(e *E, h Handler)
	// OnUnsubscribe will be called after handler was detached from the Event.
	OnUnsubscribe(e *E, h Handler)

	// OnAllPreFire will be called before all regular handlers.
	OnAllPreFire(e *E, args []any)
	// OnPreFire will be called before a specific handler is called.
	OnPreFire(e *E, h Handler, args []any)
	// OnPstFire will be called after a specific handler is called.
	OnPostFire(e *E, h Handler, args []any)
	// OnAllPostFire will be called after all regular handlers have been called and the subscription list have been
	// updated.
	OnAllPostFire(e *E, args []any)
}
