package events

type Handlerware interface {
	// OnUse will be called when the Handlerware is attached to the Event.
	OnUse(e *Event) error
	// OnDisuse will be called when the Handlerware is detached from the Event.
	OnDisuse(e *Event) error

	// OnAllPreFire will be called before all regular handlers.
	OnAllPreFire(e *Event, args []any)
	// OnPreFire will be called before a specific handler is called.
	OnPreFire(e *Event, l Listener, args []any)
	// OnPstFire will be called after a specific handler is called.
	OnPostFire(e *Event, l Listener, args []any)
	// OnAllPostFire will be called after all regular handlers have been called and the subscription list have been
	// updated.
	OnAllPostFire(e *Event, args []any)
}

type Nilware struct{}

func (_ *Nilware) OnUse(e *Event) error                        { return nil }
func (_ *Nilware) OnDisuse(e *Event) error                     { return nil }
func (_ *Nilware) OnAllPreFire(e *Event, args []any)           {}
func (_ *Nilware) OnPreFire(e *Event, l Listener, args []any)  {}
func (_ *Nilware) OnPostFire(e *Event, l Listener, args []any) {}
func (_ *Nilware) OnAllPostFire(e *Event, args []any)          {}
