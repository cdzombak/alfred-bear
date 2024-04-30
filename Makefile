SHELL:=/usr/bin/env bash
VERSION:=$(shell ./.version.sh)

default: help
.PHONY: help  # via https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Print help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: clean build package

.PHONY: clean
clean: ## Remove all build outputs
	rm -rf .pkg
	rm -rf out

.PHONY: build
build: ## Build the alfred-bear binary
	mkdir -p out
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./out/alfred-bear-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ./out/alfred-bear-arm64 .
	lipo -create -output ./out/alfred-bear ./out/alfred-bear-amd64 ./out/alfred-bear-arm64

.PHONY: package  # TODO(cdzombak):
package: ## Package the workflow for distribution
	rm -rf ./.pkg
	mkdir -p ./.pkg/bin
	cp -v ./out/alfred-bear ./.pkg/bin/alfred-bear
	ln ./images/icon.png ./.pkg/32A6D04F-624E-4CC2-8D52-DA218FA43111.png
	ln ./images/icon.png ./.pkg/icon.png
	cp -v ./workflow/info.plist ./.pkg/info.plist
	cp -v ./workflow/prefs.plist ./.pkg/prefs.plist
	sed -i '' -e 's/__WORKFLOW_VERSION__/${VERSION}/g' ./.pkg/info.plist
	mkdir -p ./out
	cd ./.pkg && zip -r workflow.zip * && mv -v workflow.zip ../out/alfred-bear-${VERSION}.alfredworkflow

.PHONY: lint
lint: ## Lint all source files in this repository (requires nektos/act: https://nektosact.com)
	act --artifact-server-path /tmp/artifacts -j lint

.PHONY: update-lint
update-lint: ## Pull updated images supporting the lint target (may fetch >10 GB!)
	docker pull catthehacker/ubuntu:full-latest
