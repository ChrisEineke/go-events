package events

type Handlerware interface {
	// OnUse will be called when the Handlerware is attached to the Event.
	OnUse(e Event) error
	// OnDisuse will be called when the Handlerware is detached from the Event.
	OnDisuse(e Event) error

	// OnSubscribe will be called after a hanlder was attached to the Event.
	OnSubscribe(e Event, h Handler) error
	// OnUnsubscribe will be called after handler was detached from the Event.
	OnUnsubscribe(e Event, h Handler) error

	// OnAllPreFire will be called before all regular handlers.
	OnAllPreFire(e Event, args []any) error
	// OnPreFire will be called before a specific handler is called.
	OnPreFire(e Event, h Handler, args []any) error

	// OnPostFire will be called after a specific handler is called.
	OnPostFire(e Event, h Handler, args []any) error
	// OnAllPostFire will be called after all regular handlers have been called and the subscription list have been
	// updated.
	OnAllPostFire(e Event, args []any) error
}
