package listener

import (
	"testing"
)

func TestWSListener_GetProtocol(t *testing.T) {
	// Create a test config
	config := Config{
		Address:        "127.0.0.1:8082",
		BufferSize:     1024,
		MaxConnections: 10,
		Timeout:        30,
	}

	// Create a WebSocket listener
	listener := NewWSListener(config)

	// Test protocol
	if listener.GetProtocol() != "ws" {
		t.Errorf("Expected protocol 'ws', got '%s'", listener.GetProtocol())
	}

	// Test config
	if listener.GetConfig().Address != config.Address {
		t.Errorf("Expected address '%s', got '%s'", config.Address, listener.GetConfig().Address)
	}
}
