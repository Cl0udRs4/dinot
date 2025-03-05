package listener

import (
	"context"
	"net"
	"os"
	"testing"
	"time"
)

func TestICMPListener_Start(t *testing.T) {
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

	// Clean up
	defer listener.Stop()

	// Give the listener time to start
	time.Sleep(100 * time.Millisecond)

	// Note: We do not actually try to connect to the ICMP listener in this test
	// because it requires creating raw ICMP packets, which is complex and requires root privileges.
	// Instead, we just verify that the listener starts successfully.
}
