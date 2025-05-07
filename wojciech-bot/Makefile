.PHONY: build_prod build_prod_linux

PACKAGE_JSON := ../package.json
VERSION := $(shell jq -r '.version' $(PACKAGE_JSON))
OUT := ./wojciech-bot

build_prod_linux:
	@echo "Building with version: $(VERSION)"
	CGO_ENABLED=0 GOOS=linux go build -ldflags "-X lib/metadata.Version=$(VERSION)" -o $(OUT)

build_prod:
	@echo "Building with version: $(VERSION)"
	CGO_ENABLED=0  go build -ldflags "-X lib/metadata.Version=$(VERSION)" -o $(OUT)