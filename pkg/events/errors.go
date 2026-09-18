package events

import "errors"

var ErrHandlerAlreadyExists = errors.New("handler already exists")
var ErrHandlerNotFound = errors.New("handler not found")
var ErrInvalidFunction = errors.New("invalid function")
var ErrInvalidEvent = errors.New("invalid event")
var ErrInvalidEventName = errors.New("invalid event name")
var ErrInvalidHandler = errors.New("invalid handler")
var ErrNoHandlers = errors.New("no handlers")
