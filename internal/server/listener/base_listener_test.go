package listener

import (
	"testing"
)

func TestBaseListener(t *testing.T) {
	// Create a test config
	config := Config{
		Address:        "127.0.0.1:0",
		BufferSize:     1024,
		MaxConnections: 10,
		Timeout:        30,
	}

	// Create a base listener
	baseListener := NewBaseListener("test", config)

	// Test GetProtocol
	if baseListener.GetProtocol() != "test" {
		t.Errorf("Expected protocol 'test', got '%s'", baseListener.GetProtocol())
	}
}
