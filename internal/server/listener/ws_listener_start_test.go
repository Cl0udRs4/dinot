package listener

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestWSListener_Start(t *testing.T) {
	// Create a test config with a fixed port for testing
	config := Config{
		Address:        "127.0.0.1:8082",
		BufferSize:     1024,
		MaxConnections: 10,
		Timeout:        30,
	}

	// Create a WebSocket listener
	listener := NewWSListener(config)

	// Create a mock connection handler
	handler := func(conn net.Conn) {
		// Just close the connection in the test
		conn.Close()
	}

	// Start the listener
	ctx := context.Background()
	err := listener.Start(ctx, handler)
	if err != nil {
		t.Fatalf("Failed to start WebSocket listener: %v", err)
	}

	// Ensure the listener is running
	if listener.GetStatus() != StatusRunning {
		t.Errorf("Expected status running, got %s", listener.GetStatus())
	}

	// Clean up
	defer listener.Stop()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	// Test connection to the listener
	// For WebSocket, we will just make a simple HTTP request
	resp, err := http.Get("http://127.0.0.1:8082")
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket listener: %v", err)
	}
	defer resp.Body.Close()

	// Verify that we got a response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
