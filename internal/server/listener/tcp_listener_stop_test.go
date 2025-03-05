package listener

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestTCPListener_Stop(t *testing.T) {
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

	// Get the actual address before stopping
	actualAddr := listener.GetConfig().Address

	// Stop the listener
	err = listener.Stop()
	if err != nil {
		t.Fatalf("Failed to stop TCP listener: %v", err)
	}

	// Ensure the listener is stopped
	if listener.GetStatus() != StatusStopped {
		t.Errorf("Expected status stopped, got %s", listener.GetStatus())
	}

	// Verify that we cannot connect to the stopped listener
	// Give it a short timeout to avoid hanging the test
	conn, err := net.DialTimeout("tcp", actualAddr, 500*time.Millisecond)
	if err == nil {
		conn.Close()
		t.Errorf("Expected connection to fail after listener stopped, but it succeeded")
	}
}
