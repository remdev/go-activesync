SHELL := /bin/sh

GO       ?= go
PKG      ?= ./...
COVER    ?= cover.out
FUZZTIME ?= 30s

.PHONY: all test race vet lint spec-lint cover fuzz tidy clean

all: vet lint test

test:
	$(GO) test -count=1 $(PKG)

race:
	$(GO) test -race -count=1 $(PKG)

vet:
	$(GO) vet $(PKG)

lint:
	@command -v staticcheck >/dev/null 2>&1 || { \
	  echo "installing staticcheck"; \
	  $(GO) install honnef.co/go/tools/cmd/staticcheck@latest; \
	}
	staticcheck $(PKG)

spec-lint:
	$(GO) run ./internal/spec/cmd/speclint

cover:
	$(GO) test -race -count=1 -covermode=atomic -coverprofile=$(COVER) $(PKG)
	$(GO) tool cover -func=$(COVER) | tail -n 1

fuzz:
	$(GO) test ./wbxml -run='^$$' -fuzz=FuzzDecode -fuzztime=$(FUZZTIME)

tidy:
	$(GO) mod tidy

clean:
	rm -f $(COVER) cover.html
