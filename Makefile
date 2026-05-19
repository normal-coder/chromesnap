BINARY     := chromesnap
CMD        := ./cmd/chromesnap
BIN_DIR    := ./bin

VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_VERSION := $(shell go version | awk '{print $$3}')
LDFLAGS    := -ldflags "-s -w \
  -X main.version=$(VERSION) \
  -X main.buildDate=$(BUILD_DATE) \
  -X main.goVersion=$(GO_VERSION)"

.PHONY: build install run clean fmt vet tidy check snapshot release

## build: compile binary for current platform → ./bin/chromesnap
build:
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY) $(CMD)
	@echo "built $(BIN_DIR)/$(BINARY) $(VERSION)"

## install: compile and install to GOPATH/bin
install:
	go install $(LDFLAGS) $(CMD)
	@echo "installed $(BINARY) $(VERSION)"

## run url="<url and flags>": build and run in one step
##   example: make run url="https://example.com -e iPhone-15 -o out.png"
run: build
	$(BIN_DIR)/$(BINARY) $(url)

## clean: remove build artifacts
clean:
	@rm -rf $(BIN_DIR) dist
	@echo "cleaned"

## fmt: format all Go source files
fmt:
	gofmt -w -l .

## vet: run go vet
vet:
	go vet ./...

## tidy: tidy go.mod and go.sum
tidy:
	go mod tidy

## check: fmt + vet + tidy (run before committing)
check: fmt vet tidy

## snapshot: local multi-platform build via goreleaser (no publish)
snapshot:
	goreleaser build --snapshot --clean

## release: full release via goreleaser (intended for CI on tag push)
release:
	goreleaser release --clean
