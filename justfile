# Copyright 2026 Canonical Ltd.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

go := require("go")

[private]
default:
    @just help

# Show available recipes
help:
    @just --list --unsorted

# Prepare the local environment
setup:
    go mod tidy

# Clean project directory
clean:
    @echo "Cleaning up workspace..."
    rm -rf bin/ coverage.out
    rm -rf .gherkindocs
    @echo "Cleaning snapcraft cache and artifacts..."
    rm -rf snap/ *.snap

# Apply static checks
check: lint vet

# Run tests for specified targets, or all tests if none specified
test *targets:
    #!/usr/bin/env bash
    if [ "{{targets}}" = "" ]; then
        just test-all
        exit 0
    fi

    for target in {{targets}}; do
        if just --show $target > /dev/null 2>&1; then
            echo "Running $target tests..."
            just $target
        else
            echo "$target tests not found, skipping."
            exit 1
        fi
    done

# Run unit tests for specified artifacts, or all artifacts if none specified
unit *args:
    @echo "Running unit tests..."
    go test -v -coverprofile=coverage.out {{args}} ./...

# Run integration tests for specified artifacts, or all artifacts if none specified
integration *args:
    @echo "Integration tests not applicable for this project."

# Apply formatting standards
fmt:
    @echo "Formatting Go code..."
    go fmt ./...

# Check against style standards
lint:
    @echo "Running linter..."
    golangci-lint run ./...

# Vet Go source code
vet:
    @echo "Running go vet..."
    go vet ./...

# Build specified artifacts, or all artifacts if none specified
build *args:
    @echo "Building gherkinator..."
    go build -o bin/gherkinator {{args}} ./cmd/gherkinator/

# Install the binary to the system GOPATH
install: build
    @echo "Installing gherkinator..."
    go install ./cmd/gherkinator/

# View HTML coverage report
coverage: unit
    @echo "Generating coverage report..."
    go tool cover -html=coverage.out

# Build the snap package using snapcraft
snap:
    @echo "Building snap package..."
    cp -r build/snap ./snap
    snapcraft pack
    rm -rf snap/
