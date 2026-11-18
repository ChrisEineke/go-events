// Package main demonstrates type-safe events using E1 (single argument).
package main

import (
	"fmt"

	"github.com/ChrisEineke/go-events"
)

type TemperatureSensor struct {
	temperateChangedEvt events.E1[float64]
}

func (t *TemperatureSensor) ReadTemperature() {
	t.temperateChangedEvt.Fire1(23.5)
}

func main() {
	sensor := &TemperatureSensor{}
	sensor.temperateChangedEvt.On(func(temp float64) error {
		fmt.Printf("Temperature: %.1f°C\n", temp)
		return nil
	})
	sensor.ReadTemperature()
}
