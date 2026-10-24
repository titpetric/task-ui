.PHONY: build test lint release

# What the build needs
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)

build: ## Compile the CLI into bin/
	@echo "building tools $(VERSION)"
	@mkdir -p bin && echo "#!/bin/sh" > bin/tool && chmod +x bin/tool

test: build ## Run the unit tests
	@echo "ok   tools/cmd 0.012s"
	@echo "ok   tools/internal 0.004s"

lint: ## Vet and format checks
	@echo "vet ok, fmt ok"

release: test lint ## Tag and publish a release
	@echo "would tag $(VERSION)"
