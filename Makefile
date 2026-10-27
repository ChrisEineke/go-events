.PHONY := build test vet coverage bench cpubench membench all docs example install clean
.DEFAULT_GOAL := all

all: build test vet

clean:
	@rm -f cpu.pprof mem.pprof coverage.out
	@rm -rf cmd/*/pkg pkg/internal/*

build:
	@go build 
	@make -C examples build

test:
	@gotestsum -f standard-verbose -- -race -coverprofile=coverage.out

vet:
	@go vet ./...

coverage: build
	@go tool cover -html=coverage.out

bench: test
	@go test -bench . -benchtime 3s

cpubench: test
	@go test -bench . -benchtime 3s -cpuprofile=cpu.pprof

membench: test
	@go test -bench . -benchtime 3s -memprofile=mem.pprof

install:
	@go install .

docs:
	@echo "Generating documentation..."
	@go doc .
