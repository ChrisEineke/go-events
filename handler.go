package events

import (
	"reflect"
)

type SubscriptionFlag int

const (
	SubscriptionOnce SubscriptionFlag = 1 << iota
	SubscriptionAsync
)

// Handler is the interface shared between generic Handlers (nullaryHandler, etc) and typ-specific Handlers (handler1,
// etc.).
type Handler interface {
	Applicable

	// funcValue returns this Handler's encapsulated function value.
	funcValue() reflect.Value
}

// Applicable abstracts the callback-calling machinery.
type Applicable interface {
	// apply invokes the function with the given arguments. This variant of apply tries to match as many arguments of
	// the event payload to the parameter list of the function (in order as fired only). The function will not be
	// invoked with more parameters than it supports. If the function has too many arguments, the remaining parameters
	// will be invoked with the parameters' zero values.
	apply(args ...any) error
}

type SubscriptionModifier func(*SubscriptionFlag)

// Sync invokes the Handler synchronously (the default).
func Sync() SubscriptionModifier {
	return func(flags *SubscriptionFlag) {
		*flags &^= SubscriptionAsync
	}
}

// Async invokes the Handler asynchronously.
func Async() SubscriptionModifier {
	return func(flags *SubscriptionFlag) {
		*flags |= SubscriptionAsync
	}
}

// Always keeps the Handler registered after being called (the default).
func Always() SubscriptionModifier {
	return func(flags *SubscriptionFlag) {
		*flags &^= SubscriptionOnce
	}
}

// Once removes the Handler after being called once.
func Once() SubscriptionModifier {
	return func(flags *SubscriptionFlag) {
		*flags |= SubscriptionOnce
	}
}
