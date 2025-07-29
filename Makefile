# Ghospel Makefile - Handles whisper.cpp dependencies and builds

.PHONY: help dev-setup build-whisper clean test build release clean-all lint lint-fix fmt vet
.DEFAULT_GOAL := help

# Variables
WHISPER_DIR := whisper_cpp_source
WHISPER_BUILD_DIR := $(WHISPER_DIR)/build
WHISPER_BIN := $(WHISPER_BUILD_DIR)/bin/whisper-cli
BINARIES_DIR := internal/binaries
GO_BINARY := ghospel

# Detect architecture
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M),arm64)
    ARCH := arm64
else ifeq ($(UNAME_M),x86_64)
    ARCH := amd64
else
    ARCH := $(UNAME_M)
endif

# Platform detection
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    PLATFORM := darwin
    CMAKE_FLAGS := -DGGML_METAL=ON -DGGML_METAL_EMBED_LIBRARY=ON -DGGML_BLAS_DEFAULT=ON
else ifeq ($(UNAME_S),Linux)
    PLATFORM := linux
    CMAKE_FLAGS := -DGGML_BLAS_DEFAULT=ON
else
    PLATFORM := unknown
endif

WHISPER_BINARY_NAME := whisper-cli-$(PLATFORM)-$(ARCH)

help: ## Show this help message
	@echo "Ghospel Build System"
	@echo "==================="
	@echo ""
	@echo "Development Commands:"
	@echo "  dev-setup          Initialize submodules and build whisper.cpp for development"
	@echo "  build-whisper      Build whisper.cpp binary with platform optimizations"
	@echo "  build              Build the Go application"
	@echo "  test               Run Go tests"
	@echo "  clean              Clean build artifacts"
	@echo ""
	@echo "Code Quality Commands:"
	@echo "  lint               Run golangci-lint (install if needed)"
	@echo "  lint-fix           Run golangci-lint with auto-fix"
	@echo "  fmt                Format code with gofmt and goimports"
	@echo "  vet                Run go vet"
	@echo ""
	@echo "Release Commands:"
	@echo "  embed-binaries     Prepare embedded binaries for distribution"
	@echo "  release            Build release version with embedded binaries"
	@echo "  clean-all          Clean everything including submodules"
	@echo ""
	@echo "Current platform: $(PLATFORM)-$(ARCH)"

dev-setup: ## Initialize development environment
	@echo "🔧 Setting up development environment..."
	@if [ ! -d "$(WHISPER_DIR)/.git" ]; then \
		echo "📦 Initializing whisper.cpp submodule..."; \
		git submodule add https://github.com/ggml-org/whisper.cpp.git $(WHISPER_DIR) || true; \
		git submodule update --init --recursive; \
	else \
		echo "📦 Updating whisper.cpp submodule..."; \
		git submodule update --recursive; \
	fi
	@$(MAKE) build-whisper
	@echo "✅ Development environment ready!"

build-whisper: ## Build whisper.cpp with platform optimizations
	@echo "🏗️  Building whisper.cpp for $(PLATFORM)-$(ARCH)..."
	@if [ ! -d "$(WHISPER_DIR)" ]; then \
		echo "❌ whisper.cpp submodule not found. Run 'make dev-setup' first."; \
		exit 1; \
	fi
	@cd $(WHISPER_DIR) && \
		cmake -B build $(CMAKE_FLAGS) \
			-DCMAKE_BUILD_TYPE=Release \
			-DWHISPER_BUILD_TESTS=OFF \
			-DWHISPER_BUILD_SERVER=OFF && \
		cmake --build build -j$(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4) --config Release
	@if [ -f "$(WHISPER_BIN)" ]; then \
		echo "✅ whisper.cpp built successfully at $(WHISPER_BIN)"; \
	else \
		echo "❌ Failed to build whisper.cpp"; \
		exit 1; \
	fi

build: build-whisper ## Build the Go application
	@echo "🏗️  Building $(GO_BINARY)..."
	@go build -o $(GO_BINARY) ./cmd/ghospel
	@echo "✅ $(GO_BINARY) built successfully!"

test: ## Run Go tests
	@echo "🧪 Running tests..."
	@go test ./...

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -f $(GO_BINARY)
	@if [ -d "$(WHISPER_BUILD_DIR)" ]; then \
		cd $(WHISPER_DIR) && make clean 2>/dev/null || rm -rf build; \
	fi
	@echo "✅ Cleaned!"

# Release-related targets for binary embedding

embed-binaries: build-whisper ## Prepare binaries for embedding
	@echo "📦 Preparing binaries for embedding..."
	@mkdir -p $(BINARIES_DIR)
	@if [ -f "$(WHISPER_BIN)" ]; then \
		cp "$(WHISPER_BIN)" "$(BINARIES_DIR)/$(WHISPER_BINARY_NAME)"; \
		echo "✅ Copied $(WHISPER_BIN) to $(BINARIES_DIR)/$(WHISPER_BINARY_NAME)"; \
	else \
		echo "❌ whisper.cpp binary not found. Run 'make build-whisper' first."; \
		exit 1; \
	fi

release: embed-binaries ## Build release version with embedded binaries
	@echo "🚀 Building release version..."
	@go build -tags release -ldflags "-s -w" -o $(GO_BINARY) ./cmd/ghospel
	@echo "✅ Release build complete!"

clean-all: clean ## Clean everything including submodules
	@echo "🧹 Deep cleaning..."
	@if [ -d "$(WHISPER_DIR)" ]; then \
		rm -rf $(WHISPER_DIR); \
	fi
	@rm -rf $(BINARIES_DIR)
	@echo "✅ Everything cleaned!"

# Utility targets

check-deps: ## Check if required dependencies are available
	@echo "🔍 Checking dependencies..."
	@command -v cmake >/dev/null 2>&1 || { echo "❌ cmake is required but not installed."; exit 1; }
	@command -v make >/dev/null 2>&1 || { echo "❌ make is required but not installed."; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "❌ go is required but not installed."; exit 1; }
	@echo "✅ All dependencies available!"

info: ## Show build information
	@echo "Build Information:"
	@echo "=================="
	@echo "Platform: $(PLATFORM)"
	@echo "Architecture: $(ARCH)"
	@echo "Go version: $(shell go version)"
	@echo "Whisper binary: $(WHISPER_BINARY_NAME)"
	@echo "CMake flags: $(CMAKE_FLAGS)"

# Code quality targets

lint: ## Run golangci-lint
	@echo "🔍 Running linters..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "📦 Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@export PATH="$$PATH:$$(go env GOPATH)/bin" && golangci-lint run

lint-fix: ## Run golangci-lint with auto-fix
	@echo "🔧 Running linters with auto-fix..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "📦 Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@export PATH="$$PATH:$$(go env GOPATH)/bin" && golangci-lint run --fix

fmt: ## Format code with gofmt and goimports
	@echo "💅 Formatting code..."
	@if ! command -v goimports >/dev/null 2>&1; then \
		echo "📦 Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	@gofmt -s -w .
	@export PATH="$$PATH:$$(go env GOPATH)/bin" && goimports -w -local github.com/pascalwhoop/ghospel .
	@echo "✅ Code formatted!"

vet: ## Run go vet
	@echo "🔬 Running go vet..."
	@go vet ./...
	@echo "✅ go vet passed!"

# Enhanced build targets with linting

build-with-lint: lint ## Build with linting first
	@$(MAKE) build

test-with-lint: lint ## Run tests with linting first  
	@$(MAKE) test