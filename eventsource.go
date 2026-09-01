package events

// EventSource is the interface shared between generic Events (E) and type-specific Events (E1, E2, ...).
type EventSource interface {
	// Name returns the name of the Event.
	Name() EventName

	// HasHandlers returns true if at least one Handler is registered, false otherwise. It is not strictly necessary to
	// call this method before calling Fire.
	HasHandlers() bool

	// Handlers returns the Handlers of the Event.
	Handlers() []Handler
}
