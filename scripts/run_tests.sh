#!/bin/bash

# Run all unit tests and generate coverage report
echo "Running unit tests..."
go test -v ./internal/... -coverprofile=coverage.out

# Display coverage report
echo "Coverage report:"
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

echo "Test results saved to coverage.out and coverage.html"
