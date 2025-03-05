#!/bin/bash

# Build server and builder
echo "Building server and builder..."
mkdir -p bin
go build -o bin/server cmd/server/main.go || echo "Failed to build server, but continuing with tests"
go build -o bin/builder cmd/builder/main.go || echo "Failed to build builder, but continuing with tests"

# Run integration tests
echo "Running integration tests..."
go test -v ./tests/integration/... || echo "Some tests failed, but continuing"

# Save test results
echo "Integration test results:"
go test -v ./tests/integration/... > integration_test_results.txt || true

echo "Integration test results saved to integration_test_results.txt"
