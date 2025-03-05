#!/bin/bash

# Build server and builder
echo "Building server and builder..."
mkdir -p bin
go build -o bin/server cmd/server/main.go
go build -o bin/builder cmd/builder/main.go

# Run integration tests
echo "Running integration tests..."
go test -v ./tests/integration/...

# Save test results
echo "Integration test results:"
go test -v ./tests/integration/... > integration_test_results.txt

echo "Integration test results saved to integration_test_results.txt"
