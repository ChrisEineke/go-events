# go-events

[![GoDoc](https://godoc.org/github.com/ChrisEineke/go-events?status.svg)](https://godoc.org/github.com/ChrisEineke/go-events)
[![Build Status](https://travis-ci.org/ChrisEineke/go-events.svg)](https://travis-ci.org/ChrisEineke/go-events)

## Summary
`go-events` is an event handling library for Go that supports middleware ("handlerware") to extend functionality.

## Contribution & Support
* Contributions produced by humans only.
* If you have a bugfix, create a pull request.
* If you have a feature or enhancement request, create an Issue ticket.

## Prerequisites
* go >= v1.25.0

## Installation
Run the following command in your terminal to add the package to your Go project:
```
go get github.com/ChrisEineke/go-events@latest
```

Then add following import statement to your code file(s):
```go
import "github.com/ChrisEineke/go-events"
```

## Example
```go
type PocketCalculator struct {}

func (p *PocketCalculator) Add(a int, b int) {
	fmt.Printf("%d\n", a + b)
}

type PocketCalculatorOperator struct {
    OnAddition event.E
}

func (p *PocketCalculatorOperator) Calculate() {
    p.OnAddition.Fire(20, 40)
}

func main() {
    calculator := &PocketCalculator{}
    operator := &PocketCalculatorOperator{}
    operator.OnAddition.On(calculator.Add)
    operator.Calculate()
}
```

## Documentation
See [GoDoc](https://godoc.org/github.com/ChrisEineke/go-events).

## Special thanks
* To [Aliaksei Saskevich](https://github.com/asaskevich/EventBus) for the original implementation.
* To [the original EventBus contributors](https://github.com/asaskevich/EventBus/graphs/contributors).
