GOLANGCI ?= ./custom-gcl

.PHONY: pre-push fmt-check vet test build lint

pre-push: fmt-check vet test build lint

fmt-check:
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "These files are not formatted with gofmt:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

vet:
	go vet ./...

test:
	go test -race -covermode=atomic -coverprofile=coverage.out ./...

build:
	go build ./...
	mkdir -p dist
	go build -o dist/loglint ./cmd/loglint

lint:
	@if [ ! -x "$(GOLANGCI)" ]; then \
		echo "Missing executable $(GOLANGCI). Download custom-gcl release binary first."; \
		exit 1; \
	fi
	$(GOLANGCI) run --timeout=5m ./...
