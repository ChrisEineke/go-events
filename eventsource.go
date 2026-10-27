package events

type EventName = string

// EventSource is the interface shared between type-specific Events (E0, E1, E2, ...).
type EventSource interface {
	// Name returns the name of the Event.
	Name() EventName

	// HasHandlers returns true if at least one Handler is registered, false otherwise. It is not strictly necessary to
	// call this method before calling Fire.
	HasHandlers() bool

	// Handlers returns the Handlers of the Event.
	Handlers() []Handler

	// Use adds the Handlerware to this Event.
	Use(Handlerware) error

	// Disuse emoves the Handlerware from this Event.
	Disuse(Handlerware) error

	// WaitAsync waits for all registered async handlers of this Event to complete.
	WaitAsync()
}
