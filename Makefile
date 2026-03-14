CUSTOM_GCL ?= ./custom-gcl
GOLANGCI ?= golangci-lint
CUSTOM_GCL_CONFIG ?= .custom-gcl.yaml

.PHONY: pre-push fmt-check vet test build lint ensure-custom-gcl

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

ensure-custom-gcl:
	@if [ -x "$(CUSTOM_GCL)" ]; then \
		echo "Using $(CUSTOM_GCL)"; \
	elif command -v custom-gcl >/dev/null 2>&1; then \
		echo "Using custom-gcl from PATH"; \
	elif command -v "$(GOLANGCI)" >/dev/null 2>&1; then \
		echo "Building $(CUSTOM_GCL) via $(GOLANGCI) custom"; \
		"$(GOLANGCI)" custom --config "$(CUSTOM_GCL_CONFIG)"; \
		if [ ! -x "$(CUSTOM_GCL)" ]; then \
			echo "Failed to build $(CUSTOM_GCL)"; \
			exit 1; \
		fi; \
	else \
		echo "Missing custom-gcl and $(GOLANGCI)."; \
		echo "Download custom-gcl from Releases or install golangci-lint and run:"; \
		echo "  $(GOLANGCI) custom --config $(CUSTOM_GCL_CONFIG)"; \
		exit 1; \
	fi

lint:
	@bin=""; \
	if [ -x "$(CUSTOM_GCL)" ]; then \
		bin="$(CUSTOM_GCL)"; \
	elif command -v custom-gcl >/dev/null 2>&1; then \
		bin="$$(command -v custom-gcl)"; \
	else \
		$(MAKE) --no-print-directory ensure-custom-gcl; \
		bin="$(CUSTOM_GCL)"; \
	fi; \
	"$$bin" run --timeout=5m ./...
