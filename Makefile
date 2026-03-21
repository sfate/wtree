BINARY  := wtree
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
BUMP    ?= patch

.PHONY: build lint audit test clean release

build:
	go build $(LDFLAGS) -o $(BINARY) .

test:
	go test ./...

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4.0 run --allow-parallel-runners

audit:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

release:
	@set -e; \
	git fetch --tags; \
	LATEST=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	MAJOR=$$(echo $$LATEST | cut -d. -f1 | tr -d v); \
	MINOR=$$(echo $$LATEST | cut -d. -f2); \
	PATCH=$$(echo $$LATEST | cut -d. -f3); \
	case "$(BUMP)" in \
	  major) NEW="v$$((MAJOR+1)).0.0" ;; \
	  minor) NEW="v$$MAJOR.$$((MINOR+1)).0" ;; \
	  *)     NEW="v$$MAJOR.$$MINOR.$$((PATCH+1))" ;; \
	esac; \
	echo "Latest: $$LATEST  →  Releasing: $$NEW"; \
	git tag -a "$$NEW" -m "Release $$NEW"; \
	git push origin "$$NEW"; \
