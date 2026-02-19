BINARY := graylogctl

.PHONY: build test lint fmt

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/$(BINARY) ./cmd/graylogctl

test:
	CGO_ENABLED=0 go test ./...

lint:
	CGO_ENABLED=0 go vet ./...

fmt:
	gofmt -w $$(rg --files -g '*.go')
