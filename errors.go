package events

import "errors"

var ErrNoHandlers = errors.New("no handlers")
var ErrInvalidCallable = errors.New("invalid callable")
var ErrHandlerNotFound = errors.New("handler not found")
