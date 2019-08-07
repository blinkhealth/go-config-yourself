.PHONY: build-deps test-deps test unit-test integration-test release os-dependencies build test/gcy build-xgo compress-binaries

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
REPORTS ?= ./test/reports

BUILD_HOST := $(shell uname -s | tr '[[:upper:]]' '[[:lower:]]')
BINARY := dist/$(BUILD_HOST)/go-config-yourself
GOBIN := $(GOPATH)/bin
export GO111MODULE=on

# --------------
# Dev setup
# --------------
os-dependencies:
	brew update && brew bundle check || brew bundle install

setup-dev: os-dependencies build-deps test-deps
	git config core.hooksPath $(shell git rev-parse --show-toplevel)/bin/githooks
	go get

lint:
	golangci-lint run

# --------------
# Testing
# --------------
test: unit-test integration-test

unit-test:
	gotestsum --format short -- -tags test ./...

integration-test: test/gcy
	GNUPGHOME="$(ROOT_DIR)/test/fixtures/gnupghome" \
		INVOKE_CMD="$(ROOT_DIR)/test/gcy -v" \
		bats --recursive test/cli && rm test/gcy

test/gcy:
	CGO=1 go build -tags test -ldflags "-X main.version=test" -o test/gcy

test-deps:
	# install outside package dir so go.sum is not affected
	cd / && GO111MODULE=auto go get -u gotest.tools/gotestsum github.com/golangci/golangci-lint/cmd/golangci-lint

coverage:
	mkdir -p $(REPORTS)/coverage
	mkdir -p $(REPORTS)/reports
	gotestsum --format short-verbose --junitfile $(REPORTS)/reports/unit.xml -- ./... -tags test -coverprofile=$(REPORTS)/coverage/profile --coverpkg=./...
	grep -vE '(go-config-yourself|commands)/[^/]+\.go' $(REPORTS)/coverage/profile > $(REPORTS)/unit.out
	go tool cover -html=$(REPORTS)/unit.out -o=$(REPORTS)/coverage/unit.html
	rm $(REPORTS)/unit.out

# --------------
# Release
# --------------
release: build
	$(ROOT_DIR)/bin/deploy/github
	$(ROOT_DIR)/bin/deploy/homebrew
	$(ROOT_DIR)/bin/deploy/ppa

dist:
	mkdir -p dist
	./bin/build/version.sh > dist/VERSION

build: dist build-deps dist/gcy-macos-amd64.tgz dist/gcy-linux-amd64.tgz

build-deps:
	cd / && GO111MODULE=auto go get -u src.techknowlogick.com/xgo
	docker build --tag xgo ./bin/build/

build-xgo: dist
	# produce debug-symbol stripped binaries
	$(GOBIN)/xgo --image xgo \
		--targets 'linux/amd64 darwin-10.10/amd64' \
		--out 'dist/gcy' \
		--ldflags "-s -w -X main.version=$(shell sed 's/^v//' dist/VERSION)" \
		$(ROOT_DIR)
	mv dist/gcy-darwin-10.10-amd64 dist/gcy-macos-amd64
	$(MAKE) compress-binaries

compress-binaries:
	docker run --rm --tty \
		-v $(ROOT_DIR)/dist:/target \
		--entrypoint upx \
		xgo -9 --no-progress /target/gcy-linux-amd64 /target/gcy-macos-amd64

build-local: dist
	mkdir -p dist/local
	CGO=1 go build -ldflags "-s -w -X main.version=$(shell sed 's/^v//' dist/VERSION)" -o dist/local/gcy
	upx -9 dist/local/gcy

dist/gcy-macos-amd64.tgz: docs build-xgo
	mkdir -p dist/macos
	cp dist/gcy-macos-amd64 dist/macos/gcy
	cp -r bin/autocomplete dist/macos
	cp -r dist/docs/man dist/macos
	tar -cf dist/gcy-macos-amd64.tgz -C "$(ROOT_DIR)/dist/macos" .
	openssl dgst -sha256 dist/gcy-macos-amd64.tgz | awk '{print $$2}' > dist/gcy-macos-amd64.shasum
	rm -rf dist/macos

dist/gcy-linux-amd64.tgz: docs build-xgo
	mkdir -p dist/linux
	cp dist/gcy-linux-amd64 dist/linux/gcy
	cp -r bin/autocomplete dist/linux
	cp -r dist/docs/man dist/linux
	tar -cf dist/gcy-linux-amd64.tgz -C "$(ROOT_DIR)/dist/linux" .
	openssl dgst -sha256 dist/gcy-linux-amd64.tgz | awk '{print $$2}' > dist/gcy-linux-amd64.shasum
	rm -rf dist/linux

DOCS := $(shell find pkg/crypto -name '*.md')
MAN_PAGES := $(patsubst pkg/crypto/%/README.md,dist/docs/man/go-config-yourself-%.5,$(DOCS))
dist/docs/man/go-config-yourself-%.5: pkg/crypto/%/README.md
	mkdir -p $(@D)
	pandoc --standalone \
		--wrap=preserve \
		--lua-filter './bin/docs/pandoc-filters/man.lua' \
		--metadata 'adjusting=l' \
		--metadata 'header=go-config-yourself' \
		--metadata "date=$(shell cat dist/VERSION)" \
		--metadata 'hyphenate=false' \
		--metadata='title=go-config-yourself-$*(5) go-config-yourself help' \
		--to man $< -o $@

dist/docs/man/go-config-yourself.1: README.md
	mkdir -p $(@D)
	pandoc --standalone \
		--wrap=preserve \
		--lua-filter './bin/docs/pandoc-filters/man.lua' \
		--metadata 'adjusting=l' \
		--metadata 'header=go-config-yourself' \
		--metadata "date=$(shell cat dist/VERSION)" \
		--metadata 'hyphenate=false' \
		--metadata='title=go-config-yourself(1) gcy help' \
		--to man $< -o $@


MDFILES := $(shell find . -path ./.github -prune -o -name '*.md' -print)
HTML_PAGES := $(patsubst %.md,dist/docs/html/%.html,$(MDFILES))
dist/docs/html/%.html: %.md
	mkdir -p $(@D)
	pandoc --standalone \
		--wrap=preserve \
		--lua-filter './bin/docs/pandoc-filters/html.lua' \
		--metadata 'hyphenate=false' \
		--metadata pagetitle="go-config-yourself: $(patsubst .,$(subst .md,,$(notdir $<)),$(notdir $(patsubst %/,%,$(dir $<))))" \
		--template bin/docs/template.html \
		--to html $< -o $(@:README.html=index.html)

docs: dist/docs/man/go-config-yourself.1 $(MAN_PAGES) $(HTML_PAGES)
