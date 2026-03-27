BINARY=codebench
VERSION=0.1.0
GOFLAGS=-ldflags "-s -w"

.PHONY: build test clean install lint coverage

build:
	go build $(GOFLAGS) -o $(BINARY) ./cmd/codebench

test:
	go test ./...

clean:
	rm -f $(BINARY)
	rm -rf .codebench/

install: build
	cp $(BINARY) $(GOPATH)/bin/$(BINARY)

lint:
	go vet ./...

coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
