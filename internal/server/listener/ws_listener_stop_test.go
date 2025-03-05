package listener

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestWSListener_Stop(t *testing.T) {
	// Create a test config with a random available port
	config := Config{
		Address:        "127.0.0.1:0", // Use port 0 to get a random available port
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

	// Get the server address before stopping
	serverAddr := listener.server.Addr

	// Stop the listener
	err = listener.Stop()
	if err != nil {
		t.Fatalf("Failed to stop WebSocket listener: %v", err)
	}

	// Ensure the listener is stopped
	if listener.GetStatus() != StatusStopped {
		t.Errorf("Expected status stopped, got %s", listener.GetStatus())
	}

	// Give the server time to stop
	time.Sleep(100 * time.Millisecond)

	// Verify that we cannot connect to the stopped listener
	client := &http.Client{
		Timeout: 500 * time.Millisecond,
	}
	_, err = client.Get("http://" + serverAddr)
	if err == nil {
		t.Errorf("Expected connection to fail after listener stopped, but it succeeded")
	}
}
