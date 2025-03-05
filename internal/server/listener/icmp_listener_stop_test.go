package listener

import (
	"context"
	"net"
	"os"
	"testing"
)

func TestICMPListener_Stop(t *testing.T) {
	// Skip this test if not running as root
	if os.Getuid() != 0 {
		t.Skip("Skipping test that requires root privileges")
	}

	// Create a test config
	config := Config{
		Address:        "127.0.0.1",
		BufferSize:     1024,
		MaxConnections: 10,
		Timeout:        30,
	}

	// Create an ICMP listener
	listener := NewICMPListener(config)

	// Create a mock connection handler
	handler := func(conn net.Conn) {
		// Just close the connection in the test
		conn.Close()
	}

	// Start the listener
	ctx := context.Background()
	err := listener.Start(ctx, handler)
	if err != nil {
		t.Fatalf("Failed to start ICMP listener: %v", err)
	}

	// Ensure the listener is running
	if listener.GetStatus() != StatusRunning {
		t.Errorf("Expected status running, got %s", listener.GetStatus())
	}

	// Stop the listener
	err = listener.Stop()
	if err != nil {
		t.Fatalf("Failed to stop ICMP listener: %v", err)
	}

	// Ensure the listener is stopped
	if listener.GetStatus() != StatusStopped {
		t.Errorf("Expected status stopped, got %s", listener.GetStatus())
	}
}
