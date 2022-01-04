export GO111MODULE=on


BIN := src-fingerprint

GO ?= go
GOFLAGS := -v
EXTRA_GOFLAGS ?=
LDFLAGS := $(LDFLAGS) -X main.version=dev -X main.builtBy=makefile`
SOURCES ?= $(shell find ./* -name "*.go" -type f ! -path "./vendor/*")

.PHONY: default
default: build

.PHONY: clean
clean:
	$(GO) clean $(GOFLAGS) -i ./...
	rm -rf $(BIN)

.PHONY: lint
lint:
	golangci-lint run

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test:
	@test -z "$$(gofmt -l $(SOURCES))" || (echo "Files need to be linted. Use make fmt" && false)
	GIT_TERMINAL_PROMPT=0 $(GO) test $(GOFLAGS) ./... -coverprofile=.coverage.out
	go tool cover -func=.coverage.out

.PHONY: build
build: $(BIN)

# Use this target to test packaging
.PHONY: dist
dist:
	goreleaser release --skip-publish --skip-validate --rm-dist

$(BIN): $(SOURCES)
	$(GO) build $(GOFLAGS) -ldflags '-s -w $(LDFLAGS)' $(EXTRA_GOFLAGS) -o $@ ./cmd/$(BIN)
