package listener

import (
	"testing"
)

func TestICMPListener_GetProtocol(t *testing.T) {
	// Create a test config
	config := Config{
		Address:        "0.0.0.0",
		BufferSize:     1024,
		MaxConnections: 10,
		Timeout:        30,
	}

	// Create an ICMP listener
	listener := NewICMPListener(config)

	// Test protocol
	if listener.GetProtocol() != "icmp" {
		t.Errorf("Expected protocol 'icmp', got '%s'", listener.GetProtocol())
	}

	// Test config
	if listener.GetConfig().Address != config.Address {
		t.Errorf("Expected address '%s', got '%s'", config.Address, listener.GetConfig().Address)
	}
}
