// Package main demonstrates basic event usage with context.
package main

import (
	"fmt"

	"github.com/ChrisEineke/go-events"
)

func main() {
	// Create an event
	e := &events.E3[int, int, *int]{N: "add"}

	// Define the Add method with context
	addFunc := func(a, b int, result *int) error {
		*result = a + b
		return nil
	}

	// Subscribe the calculator to the addition event
	e.On(addFunc)

	// Fire the event with context and arguments
	var result int
	err := e.Fire3(20, 40, &result)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(result)
}
