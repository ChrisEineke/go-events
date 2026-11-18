# go-events - Agent Guide

This guide helps AI agents work with the go-events repository efficiently.

## Repository Structure

```
go-events/         # Core event library
├── docs/          # API documentation
├── examples/      # Usage examples
│   ├── basic/     # Basic event usage
│   └── e1/        # Generic event type examples
├── internal/      # Private implementation details
├── event1.go          # E1 generic event type with 1 type parameter
├── event2.go          # E2 generic event type with 2 type parameters
├── event3.go          # E3 generic event type with 3 type parameters
├── event4.go          # E4 generic event type with 4 type parameters
├── eventsource.go     # EventSource interface (shared by E and E1-E4)
├── errors.go          # Error types
├── handler.go         # Common handler infrastructure (nullaryHandler, nAryHandler, SubscriptionModifier)
├── handler1.go        # Handler1 implementation for E1
├── handler2.go        # Handler2 implementation for E2
├── handler3.go        # Handler3 implementation for E3
├── handler4.go        # Handler4 implementation for E4
├── handlerware.go     # Handlerware interface
├── handlerware_context.go    # Context handlerware
├── handlerware_context_test.go # Context handlerware tests
├── handlerware_logger.go    # Logger handlerware
├── handlerware_logger_test.go # Logger handlerware tests
├── handlerware_nilware.go    # Nilware handlerware
├── nilware_test.go     # Nil handler tests
├── package.go          # Package declarations (stub)
├── registry.go         # Event registration (stub)
├── registry_test.go    # Registration tests (stub)
└── coverage.out        # Test coverage report
├── Makefile       # Build automation
└── go.mod/go.sum  # Go module definitions
```

## Quick Start

```bash
# Build
make build

# Run tests
make test
```

## Core Concepts

### Generic Event Types

The library provides four generic event types, each accepting a different number of generic type parameters. These determine how many arguments handlers receive when the event is fired:

- **E1[T1]** - Accepts 1 generic type parameter. Handlers receive exactly 1 argument of type T1.
- **E2[T1, T2]** - Accepts 2 generic type parameters. Handlers receive exactly 2 arguments of types T1 and T2.
- **E3[T1, T2, T3]** - Accepts 3 generic type parameters. Handlers receive exactly 3 arguments of types T1, T2, and T3.
- **E4[T1, T2, T3, T4]** - Accepts 4 generic type parameters. Handlers receive exactly 4 arguments of types T1, T2, T3, and T4.

All generic event types share common features:
- **Handlerware support** via `Use()` and `Disuse()` for middleware
- **Subscription modifiers** via `On()` callable with options
- **Unsubscription** via `Off()`
- **Async support** via `WaitAsync()` for waiting on async handlers
- **Thread-safe** Fire() calls (E3 and E4 use RWMutex for better concurrency)

### Usage Pattern

```go
import "github.com/ChrisEineke/go-events/pkg/events"

// Create an event with 2 generic type parameters
e := &events.E2[string, int]{}

// Subscribe with a handler that accepts exactly 2 arguments
e.On(func(msg string, count int) error {
    // Handle event
    return nil
})

// Fire the event with 2 arguments
e.Fire("hello", 42)
```

## Development Workflow

1. **Read** the relevant files in `pkg/events/`
2. **Test** with `make test` before any changes (and to find concurrency issues)
3. **Build** to verify compilation: `make build`

## Common Tasks

### Running Examples

```bash
make examples  # Build all examples
cd examples/basic && go run main.go
```

### Coverage

```bash
go test -coverprofile=coverage.out ./pkg/events/
go tool cover -html=coverage.out
```

## API Documentation

See `docs/api.md` for the complete API reference.

## Testing Guidelines

- All event types must have corresponding test files
- Test coverage target: ~90%
- Run `make test` to verify all tests pass and check for race conditions

## Important Notes

- Go 1.25+ required
- The library moved from root to `pkg/events/` - imports must be updated
- Generic event types replace reflection-based approaches
- Handlerware pattern enforces context.Context as first parameter
