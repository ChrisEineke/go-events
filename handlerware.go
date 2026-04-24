package events

type Handlerware interface {
	// OnUse will be called when the Handlerware is attached to the Event.
	OnUse(e *Event) error
	// OnDisuse will be called when the Handlerware is detached from the Event.
	OnDisuse(e *Event) error

	// OnAllPreFire will be called before all regular handlers.
	OnAllPreFire(e *Event, args []any)
	// OnAllPostFire will be called after all regular handlers have been called and the subscription list have been
	// updated.
	OnAllPostFire(e *Event, args []any)
}
