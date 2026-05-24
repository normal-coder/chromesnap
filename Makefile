BINARY     := chromesnap
CMD        := ./cmd/chromesnap
BIN_DIR    := ./bin

VERSION      := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
NEXT_VERSION := $(shell git-cliff --bumped-version 2>/dev/null)
BUILD_DATE   := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_VERSION   := $(shell go version | awk '{print $$3}')
LDFLAGS      := -ldflags "-s -w \
  -X main.version=$(VERSION) \
  -X main.buildDate=$(BUILD_DATE) \
  -X main.goVersion=$(GO_VERSION)"

.PHONY: build install run clean fmt vet tidy check snapshot next-version release

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

## next-version: show what the next version would be based on commits
next-version:
	@echo "current: $(VERSION)"
	@echo "next:    $(NEXT_VERSION)"
	@echo "---"
	@git-cliff --unreleased

## release: tag and push to trigger CI release
##   auto-bumps version from commits; override with: make release VERSION=v2.0.0
release:
	$(eval TAG := $(if $(VERSION_OVERRIDE),$(VERSION_OVERRIDE),$(NEXT_VERSION)))
	@if [ -z "$(TAG)" ]; then echo "error: could not determine next version" >&2; exit 1; fi
	@echo "==> releasing $(TAG)"
	@echo ""
	@git-cliff --unreleased --tag $(TAG)
	@echo ""
	@printf "confirm release $(TAG)? [y/N] "; read ans; [ "$$ans" = "y" ] || { echo "aborted"; exit 1; }
	@git-cliff --unreleased --tag $(TAG) -o /tmp/chromesnap-tag-msg.txt
	@if [ -f CHANGELOG.md ]; then \
		git-cliff --unreleased --tag $(TAG) --prepend CHANGELOG.md; \
	else \
		git-cliff --tag $(TAG) -o CHANGELOG.md; \
	fi
	@git add CHANGELOG.md
	@git commit -m "chore(release): $(TAG)"
	@git tag -s $(TAG) -F /tmp/chromesnap-tag-msg.txt
	@git push origin main --follow-tags
	@echo "==> pushed $(TAG), CI will build and publish the release"
