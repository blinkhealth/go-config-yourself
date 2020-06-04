.PHONY: build-deps test-deps test unit-test integration-test release os-dependencies build test/gcy build-xgo compress-binaries

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
REPORTS ?= ./test/reports
XGO_IMAGE ?= unrob/xgo-upx-gpgme:latest

BUILD_HOST := $(shell uname -s | tr '[[:upper:]]' '[[:lower:]]')
BINARY := dist/$(BUILD_HOST)/go-config-yourself
export GO111MODULE=on

autocomplete-test: build-local
	mv dist/local/gcy dist/local/qwer
	rm /usr/local/share/zsh/site-functions/_qwer || true
	sed 's/gcy/qwer/' bin/autocomplete/completion.zsh > /usr/local/share/zsh/site-functions/_qwer
	rm ~/.zcompdump*
	# now run
	# ---------------------------
	# rehash && compinit
	# source /usr/local/share/zsh/site-functions/_qwer
	# export PATH="$(pwd)/dist/local:$PATH"
	# compdef _qwer_zsh_autocomplete qwer
	# ---------------------------
	# then test with `qwer [TAB]`

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
	mkdir -p /tmp/gcy-test-dict
	echo "electrodoméstico" > /tmp/gcy-test-dict/words
	CGO=1 go build -tags test \
		-ldflags "-X main.version=test -X github.com/blinkhealth/go-config-yourself/pkg/crypto/password.validationDictionaryFolder=/tmp/gcy-test-dict" \
		-o test/gcy

coverage:
	gotestsum --format short-verbose -- ./... -tags test -coverprofile=$(REPORTS)/coverage.out --coverpkg=./...

test-deps:
	# install outside package dir so go.sum is not affected
	cd / && GO111MODULE=on go get -u gotest.tools/gotestsum github.com/golangci/golangci-lint/cmd/golangci-lint

# --------------
# Build
# --------------
dist:
	mkdir -p dist
	./bin/build/version.sh > dist/VERSION

debian:
	$(ROOT_DIR)/bin/build/debian

build: dist build-deps dist/gcy-macos-amd64.tgz dist/gcy-linux-amd64.tgz dist/gcy-linux-arm.tgz debian

build-xgo: dist
	# produce debug-symbol stripped binaries
	xgo --image $(XGO_IMAGE) \
		--targets 'linux/amd64 linux/arm-7 darwin-10.10/amd64' \
		--out 'dist/gcy' \
		--ldflags "-s -w -X main.version=$(shell sed 's/^v//' dist/VERSION)" \
		$(ROOT_DIR)
	mv dist/gcy-darwin-10.10-amd64 dist/gcy-macos-amd64
	$(MAKE) compress-binaries

compress-binaries:
	docker run --rm --tty \
		-v $(ROOT_DIR)/dist:/target \
		--entrypoint upx \
		$(XGO_IMAGE) -9 --no-progress \
		/target/gcy-linux-amd64 /target/gcy-macos-amd64 /target/gcy-linux-arm-7

build-local: dist
	mkdir -p dist/local
	CGO=1 go build -ldflags "-s -w -X main.version=$(shell sed 's/^v//' dist/VERSION)" -o dist/local/gcy
	# skip compression on local builds
	# upx -9 dist/local/gcy

dist/gcy-macos-amd64.tgz: docs build-xgo
	mkdir -p dist/macos
	cp dist/gcy-macos-amd64 dist/macos/gcy
	cp -r bin/autocomplete dist/macos
	cp -r dist/docs/man dist/macos
	tar -czf dist/gcy-macos-amd64.tgz -C "$(ROOT_DIR)/dist/macos" .
	openssl dgst -sha256 dist/gcy-macos-amd64.tgz | awk '{print $$2}' > dist/gcy-macos-amd64.shasum
	rm -rf dist/macos

dist/gcy-linux-amd64.tgz: docs build-xgo
	mkdir -p dist/linux
	cp dist/gcy-linux-amd64 dist/linux/gcy
	cp -r bin/autocomplete dist/linux
	cp -r dist/docs/man dist/linux
	cp bin/build/linux.Makefile dist/linux/Makefile
	tar -czf dist/gcy-linux-amd64.tgz -C "$(ROOT_DIR)/dist/linux" .
	openssl dgst -sha256 dist/gcy-linux-amd64.tgz | awk '{print $$2}' > dist/gcy-linux-amd64.shasum
	rm -rf dist/linux

dist/gcy-linux-arm.tgz: docs build-xgo
	mkdir -p dist/arm
	cp dist/gcy-linux-arm-7 dist/arm/gcy
	cp -r bin/autocomplete dist/arm
	cp -r dist/docs/man dist/arm
	cp bin/build/linux.Makefile dist/arm/Makefile
	tar -czf dist/gcy-linux-arm.tgz -C "$(ROOT_DIR)/dist/arm" .
	openssl dgst -sha256 dist/gcy-linux-arm.tgz | awk '{print $$2}' > dist/gcy-linux-arm.shasum
	rm -rf dist/arm

build-deps:
	cd / && GO111MODULE=auto go get -u src.techknowlogick.com/xgo

# --------------
# Documentation
# --------------
DOCS := $(shell find pkg/crypto -name '*.md')
MAN_PAGES := $(patsubst pkg/crypto/%/README.md,dist/docs/man/gcy-%.5,$(DOCS))
dist/docs/man/gcy-%.5: pkg/crypto/%/README.md
	mkdir -p $(@D)
	pandoc --standalone \
		--wrap=preserve \
		--lua-filter './bin/docs/pandoc-filters/man.lua' \
		--metadata 'adjusting=l' \
		--metadata 'header=gcy' \
		--metadata "date=$(shell cat dist/VERSION)" \
		--metadata 'hyphenate=false' \
		--metadata='title=gcy-$*(5) gcy help' \
		--to man $< -o $@

dist/docs/man/gcy.1: README.md
	mkdir -p $(@D)
	pandoc --standalone \
		--wrap=preserve \
		--lua-filter './bin/docs/pandoc-filters/man.lua' \
		--metadata 'adjusting=l' \
		--metadata 'header=gcy' \
		--metadata "date=$(shell cat dist/VERSION)" \
		--metadata 'hyphenate=false' \
		--metadata='title=gcy(1) gcy help' \
		--to man $< -o $@


MDFILES := $(shell find . -path ./.github -prune -o -name '*.md' -print)
HTML_PAGES := $(patsubst %.md,dist/docs/html/%.html,$(MDFILES))
dist/docs/html/%.html: %.md
	mkdir -p $(@D)
	pandoc --standalone \
		--wrap=preserve \
		--lua-filter './bin/docs/pandoc-filters/html.lua' \
		--metadata 'hyphenate=false' \
		--metadata pagetitle="gcy: $(patsubst .,$(subst .md,,$(notdir $<)),$(notdir $(patsubst %/,%,$(dir $<))))" \
		--template bin/docs/template.html \
		--to html $< -o $(@:README.html=index.html)

docs: dist/docs/man/gcy.1 $(MAN_PAGES) $(HTML_PAGES)
