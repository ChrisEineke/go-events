# EventBus

[![GoDoc](https://godoc.org/github.com/ChrisEineke/EventBus?status.svg)](https://godoc.org/github.com/ChrisEineke/EventBus)
[![Coverage Status](https://img.shields.io/coveralls/ChrisEineke/EventBus.svg)](https://coveralls.io/r/ChrisEineke/EventBus?branch=master)
[![Build Status](https://travis-ci.org/ChrisEineke/EventBus.svg)](https://travis-ci.org/ChrisEineke/EventBus)

Package EventBus is a light-weight event bus for Go with async compatibility.

## Contribution & Support
Contributions are welcome! If you have a bugfix, go ahead and create a pull request. For feature requests or
enhancements, create a discussion ticket first.

## Prerequisites
* go >= v1.25.0

## Installation
Run the following command in your terminal to add the package to your Go project:
```
go get github.com/ChrisEineke/EventBus@latest
```

Then add following import statement to your code file(s):
```go
import "github.com/ChrisEineke/EventBus"
```

You can shorten the package name as well:
```go
import (
	evt "github.com/ChrisEineke/EventBus"
)
```

## Example
```go
func calculator(a int, b int) {
	fmt.Printf("%d\n", a + b)
}

func main() {
	bus := EventBus.New(); // or: bus := EventBus.Singleton()
	bus.Topic("calculator.addition").On(calculator)
    bus.Topic("calculator.addition").Fire(20, 40)
	bus.Topic("calculator.addition", calculator)
}
```

## Documentation
Available over at [GoDoc](https://godoc.org/github.com/ChrisEineke/EventBus).

## Inter-Process Event Processing
Works with two rpc services:
- a client service to listen to remotely published events from a server
- a server service to listen to client subscriptions

server.go
```go
func main() {
    server := NewServer(":2010", "/_server_bus_", New())
    server.Start()
    // ...
    server.EventBus().Publish("main:calculator", 4, 6)
    // ...
    server.Stop()
}
```

client.go
```go
func main() {
    client := NewClient(":2015", "/_client_bus_", New())
    client.Start()
    client.Subscribe("main:calculator", calculator, ":2010", "/_server_bus_")
    // ...
    client.Stop()
}
```

## Special thanks
* To [Aliaksei Saskevich](https://github.com/asaskevich/EventBus) for the original implementation.
* To [the original EventBus contributors](https://github.com/asaskevich/EventBus/graphs/contributors).
