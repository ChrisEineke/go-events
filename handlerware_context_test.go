package events

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStrictContextWareImplementsInterface(t *testing.T) {
	var _ Handlerware = &StrictContextWare{}
}

func TestEventFireContext(t *testing.T) {
	StrictContextWare := NewStrictContextWare()

	e := E{N: "testEvent"}
	e.Use(StrictContextWare)
	numHandlersCalled := 0
	firstHandler := func(ctx context.Context, a int) {
		assert.NotNil(t, ctx)
		assert.Equal(t, 10, a)
		numHandlersCalled++
	}
	e.On(firstHandler)

	ctx := context.Background()
	err := e.Fire(ctx, 10)
	assert.NoError(t, err, "unexpected error during Fire")
	assert.Equal(t, 1, numHandlersCalled, "expected exactly 1 handler to be called")

	e.Off(firstHandler)
	e.Disuse(StrictContextWare)
}

func TestEventFireContextWithDeadline(t *testing.T) {
	StrictContextWare := NewStrictContextWare()

	e := E{N: "testEvent"}
	e.Use(StrictContextWare)
	numHandlersCalled := 0
	firstHandler := func(ctx context.Context, a int) {
		assert.NotNil(t, ctx)
		assert.Equal(t, 10, a)
		numHandlersCalled++
	}
	e.On(firstHandler)
	secondHandler := func(ctx context.Context, a int) {
		assert.NotNil(t, ctx)
		assert.Equal(t, 10, a)
		time.Sleep(100 * time.Millisecond)
		numHandlersCalled++
	}
	e.On(secondHandler)
	thirdHandler := func(ctx context.Context, a int) {
		assert.Fail(t, "shouldn't be reachable because the previous handler exceeded the deadline")
	}
	e.On(thirdHandler)

	ctx := context.Background()
	ctx, cancelFn := context.WithDeadline(ctx, time.Now().Add(50*time.Millisecond))
	defer cancelFn()
	err := e.Fire(ctx, 10)
	assert.Error(t, err, "Fire should return an error because of missing Context arg")
	assert.Equal(t, 2, numHandlersCalled, "expected exactly 2 handlers to be called")

	e.Off(thirdHandler)
	e.Off(secondHandler)
	e.Off(firstHandler)
	e.Disuse(StrictContextWare)
}

func TestEventFireNoArg(t *testing.T) {
	e := E{N: "testEvent"}
	e.Use(NewStrictContextWare())
	assert.Error(t, e.Fire(), "Fire should return an error because of no args")
}

func TestEventFireNoContext(t *testing.T) {
	e := E{N: "testEvent"}
	e.Use(NewStrictContextWare())
	e.On(func(ctx context.Context, a int) {
		assert.NotNil(t, ctx)
		assert.Equal(t, 10, a)
	})
	assert.Error(t, e.Fire(10, nil), "Fire should return an error because of missing Context arg")
}

func TestEventFireNoContextInHandler(t *testing.T) {
	e := E{N: "testEvent"}
	e.Use(NewStrictContextWare())
	err := e.On(func(a int) {})
	assert.Error(t, err, "On should return an error because of missing Context arg")
}

func TestEventFireNoContextInExistingHandler(t *testing.T) {
	e := E{N: "testEvent"}
	e.On(func(a int) {})
	err := e.Use(NewStrictContextWare())
	assert.Error(t, err, "Use should return an error because of missing Context arg in an existing handler")
}
