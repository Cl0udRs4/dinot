package listener

import (
	"testing"
)

func TestUDPListener_GetProtocol(t *testing.T) {
	// Create a test config
	config := Config{
		Address:        "127.0.0.1:8081",
		BufferSize:     1024,
		MaxConnections: 10,
		Timeout:        30,
	}

	// Create a UDP listener
	listener := NewUDPListener(config)

	// Test protocol
	if listener.GetProtocol() != "udp" {
		t.Errorf("Expected protocol 'udp', got '%s'", listener.GetProtocol())
	}

	// Test config
	if listener.GetConfig().Address != config.Address {
		t.Errorf("Expected address '%s', got '%s'", config.Address, listener.GetConfig().Address)
	}
}
