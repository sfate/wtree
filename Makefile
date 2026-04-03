BINARY  := wtree
VERSION_FILE := VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/sfate/wtree/internal/version.Value=$(VERSION)"
BUMP    ?= patch

INSTALL_DIR ?= /usr/local/bin

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
	if [ -n "$$(git status --porcelain)" ]; then \
	  echo "Working tree is not clean. Commit or stash changes before releasing."; \
	  exit 1; \
	fi; \
	CURRENT=$$(cat $(VERSION_FILE) 2>/dev/null || echo "v0.0.0"); \
	git fetch --tags; \
	LATEST=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	if [ "$$CURRENT" != "$$LATEST" ]; then \
	  echo "VERSION ($$CURRENT) does not match latest tag ($$LATEST)."; \
	  exit 1; \
	fi; \
	MAJOR=$$(echo $$LATEST | cut -d. -f1 | tr -d v); \
	MINOR=$$(echo $$LATEST | cut -d. -f2); \
	PATCH=$$(echo $$LATEST | cut -d. -f3); \
	case "$(BUMP)" in \
	  major) NEW="v$$((MAJOR+1)).0.0" ;; \
	  minor) NEW="v$$MAJOR.$$((MINOR+1)).0" ;; \
	  *)     NEW="v$$MAJOR.$$MINOR.$$((PATCH+1))" ;; \
	esac; \
	echo "Latest: $$LATEST  →  Releasing: $$NEW"; \
	printf '%s\n' "$$NEW" > $(VERSION_FILE); \
	git add $(VERSION_FILE); \
	git commit -m "Release $$NEW"; \
	git tag -a "$$NEW" -m "Release $$NEW"; \
	git push origin HEAD; \
	git push origin "$$NEW"; \
