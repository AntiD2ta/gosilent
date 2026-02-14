VERSION ?= dev
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test install vet

build:
	go build $(LDFLAGS) ./cmd/gosilent/

test:
	go test ./...

install:
	go install $(LDFLAGS) ./cmd/gosilent/

vet:
	go vet ./...
