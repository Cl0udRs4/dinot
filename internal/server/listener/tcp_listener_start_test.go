package listener

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestTCPListener_Start(t *testing.T) {
	// Create a test config with a random available port
	config := Config{
		Address:        "127.0.0.1:0", // Use port 0 to get a random available port
		BufferSize:     1024,
		MaxConnections: 10,
		Timeout:        30,
	}

	// Create a TCP listener
	listener := NewTCPListener(config)

	// Create a mock connection handler
	handler := func(conn net.Conn) {
		// Just close the connection in the test
		conn.Close()
	}

	// Start the listener
	ctx := context.Background()
	err := listener.Start(ctx, handler)
	if err != nil {
		t.Fatalf("Failed to start TCP listener: %v", err)
	}

	// Ensure the listener is running
	if listener.GetStatus() != StatusRunning {
		t.Errorf("Expected status running, got %s", listener.GetStatus())
	}

	// Get the actual address from the listener internal field
	actualAddr := listener.listener.Addr().String()
	if actualAddr == "127.0.0.1:0" || actualAddr == "" {
		t.Errorf("Expected a real port to be assigned, but got %s", actualAddr)
	}

	// Clean up
	defer listener.Stop()

	// Test connection to the listener
	conn, err := net.Dial("tcp", actualAddr)
	if err != nil {
		t.Fatalf("Failed to connect to TCP listener: %v", err)
	}
	defer conn.Close()

	// Give the handler time to process the connection
	time.Sleep(100 * time.Millisecond)
}
