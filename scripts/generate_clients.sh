#!/bin/bash

# Build builder if not already built
if [ ! -f "bin/builder" ]; then
    echo "Building builder..."
    mkdir -p bin
    go build -o bin/builder cmd/builder/main.go
fi

# Generate TCP client
echo "Generating TCP client..."
./bin/builder -protocol tcp -domain localhost -servers localhost:8080 -modules shell -output bin/client_tcp

# Generate UDP client
echo "Generating UDP client..."
./bin/builder -protocol udp -domain localhost -servers localhost:8081 -modules shell -output bin/client_udp

# Generate WebSocket client
echo "Generating WebSocket client..."
./bin/builder -protocol ws -domain localhost -servers localhost:8082 -modules shell -output bin/client_ws

# Generate ICMP client (requires sudo)
echo "Generating ICMP client..."
sudo ./bin/builder -protocol icmp -domain localhost -servers localhost -modules shell -output bin/client_icmp

# Generate DNS client
echo "Generating DNS client..."
./bin/builder -protocol dns -domain example.com -servers localhost:53 -modules shell -output bin/client_dns

echo "Client executables generated in bin/ directory"
