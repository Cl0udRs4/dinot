package listener

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestDNSListener_Start(t *testing.T) {
	// Create a test config
	config := DNSConfig{
		Config: Config{
			Address:        "127.0.0.1:8053", // Use a non-standard port for testing
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

	// Clean up
	defer listener.Stop()

	// Give the listener time to start
	time.Sleep(100 * time.Millisecond)

	// Note: We do not actually try to connect to the DNS listener in this test
	// because it requires creating DNS queries, which is complex.
	// Instead, we just verify that the listener starts successfully.
}
