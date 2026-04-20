package EventBus

import (
	"context"
)

type Fireable interface {
	// Fire dispatches the given payload(s) to all subscribed listeners taking into account the modifiers that they were
	// registered with. If a listener's function signature contains more parameters than provided arguments, zero values
	// will be filled in. If a listener's function contains less parameters than provided arguments, the listener will
	// be invoked will less arguments.
	Fire(args ...any)
	// Fire dispatches the given payload(s) to all subscribed listeners taking into account the modifiers that they were
	// registered with and the provided deadline. If a listener's function signature contains more parameters than
	// provided arguments, zero values will be filled in. If a listener's function contains less parameters than
	// provided arguments, the listener will be invoked will less arguments.
	FireContext(ctx context.Context, args ...any)
	// HasListeners returns true if at least one listener is registered, false otherwise.
	HasListeners() bool
	// Use adds the Handlerware to this Event.
	Use(Handlerware)
	// Discard removes the Handlerware from this Event.
	Discard(Handlerware)
}

type Subscribable interface {
	// On registers the given listener with the given modifiers. Returns an error if `listener` is not a function.
	On(listener any, options ...SubscriptionModifier) error
	// Off cancels the given listener. Returns an error if `listener` is not subscribed to this topic.
	Off(listener any) error
}

type Waitable interface {
	// WaitAsync waits for all registered async listeners of this Topic to complete.
	WaitAsync()
}

type Handlerware interface {
	// OnUse will be called when the Handlerware is attached to the Event.
	OnUse(e *Event) error
	// OnDisuse will be called when the Handlerware is detached from the Event.
	OnDisuse(e *Event) error

	// OnPreFire will be called before regular handlers.
	OnPreFire(e *Event, args ...any)
	// OnPostFire will be called after regular listeners have been called and any Once()'lers have been removed.
	OnPostFire(e *Event, args ...any)
}
