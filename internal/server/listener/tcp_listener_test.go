package listener

import (
	"testing"
)

func TestTCPListener_GetProtocol(t *testing.T) {
	// Create a test config
	config := Config{
		Address:        "127.0.0.1:8080",
		BufferSize:     1024,
		MaxConnections: 10,
		Timeout:        30,
	}

	// Create a TCP listener
	listener := NewTCPListener(config)

	// Test protocol
	if listener.GetProtocol() != "tcp" {
		t.Errorf("Expected protocol 'tcp', got '%s'", listener.GetProtocol())
	}
}
