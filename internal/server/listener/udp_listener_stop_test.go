package listener

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestUDPListener_Stop(t *testing.T) {
	// Create a test config with a random available port
	config := Config{
		Address:        "127.0.0.1:0", // Use port 0 to get a random available port
		BufferSize:     1024,
		MaxConnections: 10,
		Timeout:        30,
	}

	// Create a UDP listener
	listener := NewUDPListener(config)

	// Create a mock connection handler
	handler := func(conn net.Conn) {
		// Just close the connection in the test
		conn.Close()
	}

	// Start the listener
	ctx := context.Background()
	err := listener.Start(ctx, handler)
	if err != nil {
		t.Fatalf("Failed to start UDP listener: %v", err)
	}

	// Get the actual address before stopping
	actualAddr := listener.conn.LocalAddr().String()

	// Stop the listener
	err = listener.Stop()
	if err != nil {
		t.Fatalf("Failed to stop UDP listener: %v", err)
	}

	// Ensure the listener is stopped
	if listener.GetStatus() != StatusStopped {
		t.Errorf("Expected status stopped, got %s", listener.GetStatus())
	}

	// Verify that we cannot connect to the stopped listener
	// For UDP, we can still "connect" but sending data should fail
	conn, err := net.Dial("udp", actualAddr)
	if err != nil {
		// If we cannot connect, that is fine too
		return
	}
	defer conn.Close()

	// Try to send data
	_, err = conn.Write([]byte("Test data"))
	// We do not check the error here because UDP is connectionless
	// and the write might succeed even if the listener is stopped

	// Give a short time for any potential response
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	
	// We expect a timeout error since the listener is stopped
	if err == nil {
		t.Errorf("Expected read to fail after listener stopped, but it succeeded")
	}
}
