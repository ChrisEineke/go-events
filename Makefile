.PHONY := build test bench cpubench membench
.DEFAULT_GOAL := build

clean:
	@rm -f cpu.pprof mem.pprof

build:
	@GOAMD64=v4 go build

test: build
	@gotestsum -- -coverprofile=coverage.out

coverage:
	@go tool cover -html=coverage.out

bench: test
	@go test -bench . -benchtime 3s

cpubench: test
	@go test -bench . -benchtime 3s -cpuprofile=cpu.pprof

membench: test
	@go test -bench -membench . -benchtime 3s -memprofile=mem.pprof
