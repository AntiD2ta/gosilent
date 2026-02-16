VERSION ?= dev
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test install vet lint fix-check gosec quality test-unit test-e2e

build:
	go build $(LDFLAGS) ./cmd/gosilent/

test:
	go test ./...

install:
	go install $(LDFLAGS) ./cmd/gosilent/

vet:
	go vet ./...

lint:
	golangci-lint run

fix-check:
	@test -z "$$(go fix -diff ./...)" || { echo "go fix found un-applied modernizations:"; go fix -diff ./...; exit 1; }

gosec:
	gosec -exclude-dir=testdata ./...

quality: build vet lint fix-check gosec

test-unit:
	go test -race ./internal/...

test-e2e:
	go test -race ./e2e/...
