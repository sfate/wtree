BINARY  := wtree
VERSION_CMD := go run ./internal/version/cmd
CURRENT_VERSION := $(VERSION_CMD) get
BUMP    ?= patch

INSTALL_DIR ?= /usr/local/bin

.PHONY: build lint audit test clean release

build:
	go build -o $(BINARY) .

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
	git fetch --tags; \
	LATEST_VERSION=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	if [ "$$CURRENT_VERSION" != "$$LATEST_VERSION" ]; then \
	  echo "VERSION ($$CURRENT_VERSION) does not match latest tag ($$LATEST_VERSION)."; \
	  exit 1; \
	fi; \
	NEW_VERSION=$$($(VERSION_CMD) bump $(BUMP)); \
	echo "Latest: $$LATEST_VERSION  →  Releasing: $$NEW_VERSION"; \
	$(VERSION_CMD) set "$$NEW_VERSION"; \
	git add internal/version/VERSION; \
	git commit -m "chore(deps): bump to $$NEW_VERSION"; \
	git tag -a "$$NEW_VERSION" -m "chore(deps): bump to $$NEW_VERSION"; \
	git push origin HEAD; \
	git push origin "$$NEW_VERSION"; \
