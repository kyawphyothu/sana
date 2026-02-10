# Project variables
BINARY_NAME=sana
MAIN_PATH=./main.go
BUILD_DIR=./bin
GO=go
GOFLAGS=-v

# Version info (can be overridden)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: run
run: ## Run the application
	SANA_ENV=development $(GO) run $(MAIN_PATH)

.PHONY: run-prod
run-prod: ## Run the production application
	$(GO) run $(MAIN_PATH)

.PHONY: install
install: ## Install the binary to $GOPATH/bin (production mode by default)
	$(GO) install $(LDFLAGS)
	@echo "✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"
	@echo "  To run in development mode: SANA_ENV=development sana"
	@echo "  To run in production mode: sana (or SANA_ENV=production sana)"

.PHONY: clean
clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf ./data
	@echo "✓ Cleaned"

.PHONY: clean-prod
clean-prod: ## Remove production data
	@echo "⚠️  This will delete your production database!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		CONFIG_DIR=$$(go run ./scripts/get-config-dir.go); \
		rm -f "$$CONFIG_DIR/sana.db"; \
		echo "✓ Production database cleaned from $$CONFIG_DIR"; \
	fi

.PHONY: seed
seed: ## Seed the database
	@echo "Seeding database..."
	SANA_ENV=development go run ./seeder/seeder.go
	@echo "✓ Database seeded"

.PHONY: test
test: ## Run tests
	$(GO) test -v ./...

.PHONY: fmt
fmt: ## Format code
	$(GO) fmt ./...

.PHONY: tidy
tidy: ## Tidy go modules
	$(GO) mod tidy

.PHONY: check
check: fmt vet test ## Run fmt, vet, and test

.PHONY: all
all: clean deps check build ## Clean, download deps, check, and build

.PHONY: release
release: ## Build for multiple platforms
	@echo "Building releases..."
	@mkdir -p $(BUILD_DIR)/releases
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✓ Releases built in $(BUILD_DIR)/releases"

.DEFAULT_GOAL := help