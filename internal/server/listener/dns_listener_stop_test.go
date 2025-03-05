package listener

import (
	"context"
	"net"
	"testing"
)

func TestDNSListener_Stop(t *testing.T) {
	// Create a test config
	config := DNSConfig{
		Config: Config{
			Address:        "127.0.0.1:8054", // Use a non-standard port for testing
			BufferSize:     1024,
			MaxConnections: 10,
			Timeout:        30,
		},
		Domain:      "example.com",
		TTL:         60,
		RecordTypes: []string{"A", "TXT"},
	}

	// Create a DNS listener
	listener := NewDNSListener(config)

	// Create a mock connection handler
	handler := func(conn net.Conn) {
		// Just close the connection in the test
		conn.Close()
	}

	// Start the listener
	ctx := context.Background()
	err := listener.Start(ctx, handler)
	if err != nil {
		t.Fatalf("Failed to start DNS listener: %v", err)
	}

	// Ensure the listener is running
	if listener.GetStatus() != StatusRunning {
		t.Errorf("Expected status running, got %s", listener.GetStatus())
	}

	// Stop the listener
	err = listener.Stop()
	if err != nil {
		t.Fatalf("Failed to stop DNS listener: %v", err)
	}

	// Ensure the listener is stopped
	if listener.GetStatus() != StatusStopped {
		t.Errorf("Expected status stopped, got %s", listener.GetStatus())
	}
}
