package listener

import (
	"testing"
)

func TestListenerManager(t *testing.T) {
	// Create a default config
	config := Config{
		Address:        "127.0.0.1:8080",
		BufferSize:     1024,
		MaxConnections: 10,
		Timeout:        30,
	}

	// Create a listener manager
	manager := NewListenerManager(config)

	// Test creating a TCP listener
	_, err := manager.CreateListener("tcp", config)
	if err != nil {
		t.Errorf("Failed to create TCP listener: %v", err)
	}

	// Test getting a listener
	listener, err := manager.GetListener("tcp")
	if err != nil {
		t.Errorf("Failed to get TCP listener: %v", err)
	}

	if listener.GetProtocol() != "tcp" {
		t.Errorf("Expected protocol 'tcp', got '%s'", listener.GetProtocol())
	}

	// Test unregistering a listener
	err = manager.UnregisterListener("tcp")
	if err != nil {
		t.Errorf("Failed to unregister TCP listener: %v", err)
	}

	// Test getting a listener after unregistering
	_, err = manager.GetListener("tcp")
	if err == nil {
		t.Errorf("Expected error when getting unregistered listener")
	}
}
