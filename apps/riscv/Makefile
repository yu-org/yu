# Makefile for RISC-V VM Integration

# Variables
CKB_VM_REPO := https://github.com/nervosnetwork/ckb-vm.git
CKB_VM_DIR := ckb-vm
WRAPPER_DIR := ckb-vm-wrapper
CKB_VM_BUILD_DIR := $(WRAPPER_DIR)/target/release
CKB_VM_LIB_DIR := $(CKB_VM_BUILD_DIR)

# Go build tags
BUILD_TAGS := cgo

.PHONY: all clean build test deps build-wrapper

# Default target
all: build-wrapper build

# Build the Rust wrapper
build-wrapper:
	@echo "Building CKB-VM wrapper..."
	cd $(WRAPPER_DIR) && cargo build --release
	@echo "CKB-VM wrapper build complete!"

# Build the Go module
build: build-wrapper
	@echo "Building RISC-V module..."
	CGO_ENABLED=1 go build -tags $(BUILD_TAGS) .

# Compile test programs
compile-test-programs:
	@echo "Compiling RISC-V test programs..."
	@mkdir -p test_programs/compiled
	@cd test_programs && ./build_test_programs.sh

# Run tests
test: build-wrapper compile-test-programs
	@echo "Running tests..."
	CGO_ENABLED=1 go test -tags $(BUILD_TAGS) -v .

# Install dependencies
deps:
	@echo "Installing Go dependencies..."
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(WRAPPER_DIR)/target
	go clean

# Check if CKB-VM wrapper is properly built
check-wrapper:
	@echo "Checking CKB-VM wrapper..."
	@if [ ! -f "$(CKB_VM_LIB_DIR)/libckb_vm_wrapper.a" ]; then \
		echo "ERROR: CKB-VM wrapper static library not found. Run 'make build-wrapper' first."; \
		exit 1; \
	fi
	@echo "CKB-VM wrapper static library is correctly built!"

# Show build information
info:
	@echo "Build Information:"
	@echo "  CKB-VM Repository: $(CKB_VM_REPO)"
	@echo "  Wrapper Directory: $(WRAPPER_DIR)"
	@echo "  Build Tags: $(BUILD_TAGS)"
	@echo "  Go Version: $$(go version)"
	@echo "  CGO Enabled: $$(go env CGO_ENABLED)"
	@echo "  Rust Version: $$(rustc --version 2>/dev/null || echo 'Rust not found')"

# Help target
help:
	@echo "Available targets:"
	@echo "  all          - Build wrapper and Go module (default)"
	@echo "  build-wrapper- Build the Rust wrapper"
	@echo "  build        - Build the Go module"
	@echo "  test         - Run tests"
	@echo "  deps         - Install Go dependencies"
	@echo "  clean        - Clean build artifacts"
	@echo "  check-wrapper- Check if CKB-VM wrapper is properly built"
	@echo "  info         - Show build information"
	@echo "  help         - Show this help message"
