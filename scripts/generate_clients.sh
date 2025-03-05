#!/bin/bash

# Script to generate client executables with different configurations
# Usage: ./generate_clients.sh [output_directory]

# Default output directory
OUTPUT_DIR=${1:-"./bin/clients"}

# Ensure the builder is built
if [ ! -f "./bin/builder" ]; then
    echo "Building builder..."
    mkdir -p bin
    go build -o bin/builder cmd/builder/main.go
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Generate TCP client
echo "Generating TCP client..."
./bin/builder -protocol=tcp -domain=localhost -servers=127.0.0.1:8080 -modules=shell,system -encryption=aes -output="$OUTPUT_DIR/client_tcp_aes"

# Generate UDP client
echo "Generating UDP client..."
./bin/builder -protocol=udp -domain=localhost -servers=127.0.0.1:8081 -modules=shell,system -encryption=chacha20 -output="$OUTPUT_DIR/client_udp_chacha20"

# Generate WebSocket client
echo "Generating WebSocket client..."
./bin/builder -protocol=ws -domain=localhost -servers=127.0.0.1:8082 -modules=shell,system,file -encryption=aes -output="$OUTPUT_DIR/client_ws_aes"

# Generate ICMP client (requires sudo for running)
echo "Generating ICMP client..."
./bin/builder -protocol=icmp -domain=localhost -servers=127.0.0.1 -modules=shell,system -encryption=chacha20 -output="$OUTPUT_DIR/client_icmp_chacha20"

# Generate DNS client
echo "Generating DNS client..."
./bin/builder -protocol=dns -domain=example.com -servers=127.0.0.1:8053 -modules=shell,system -encryption=aes -output="$OUTPUT_DIR/client_dns_aes"

# Generate multi-protocol client
echo "Generating multi-protocol client..."
./bin/builder -protocol=tcp,udp,ws -domain=localhost -servers=127.0.0.1:8080,127.0.0.1:8081,127.0.0.1:8082 -modules=shell,system,file -encryption=aes -output="$OUTPUT_DIR/client_multi_aes"

echo "Client generation complete. Clients are available in $OUTPUT_DIR"
