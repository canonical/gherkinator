# justfile
set shell := ["bash", "-c"]

[private]
default: help

# Build the gherkinator binary
build: fmt
    @echo "Building gherkinator..."
    go build -o bin/gherkinator .

# Apply formatting standards to project
fmt:
    @echo "Formatting Go code..."
    go fmt ./...

# Check project against coding style standards
lint:
    @echo "Running linter..."
    golangci-lint run ./...

# Run unit tests with testify and generate coverage profile
unit:
    @echo "Running unit tests..."
    go test -v -coverprofile=coverage.out ./...

# View HTML coverage report
coverage: unit
    @echo "Generating coverage report..."
    go tool cover -html=coverage.out

# Install the binary to the system GOPATH
install: build
    @echo "Installing gherkinator..."
    go install .

# Build the snap package using snapcraft
snap:
    @echo "Building snap package..."
    snapcraft

# Clean build artifacts, temporary doc server files, and snapcraft artifacts
clean:
    @echo "Cleaning up workspace..."
    rm -rf bin/ coverage.out
    rm -rf .gherkindocs
    @echo "Cleaning snapcraft cache and artifacts..."
    snapcraft clean || true
    rm -f *.snap

# Show available recipes
help:
    @just --list --unsorted
