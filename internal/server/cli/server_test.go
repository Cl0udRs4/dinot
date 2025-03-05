package cli

import (
	"testing"
	"time"
)

// TestServerCreation tests the creation of a new server
func TestServerCreation(t *testing.T) {
	server := NewServer()
	
	if server == nil {
		t.Fatalf("NewServer() returned nil")
	}
	
	if server.clientManager == nil {
		t.Errorf("Server.clientManager is nil")
	}
	
	if server.heartbeatMonitor == nil {
		t.Errorf("Server.heartbeatMonitor is nil")
	}
	
	if server.listenerManager == nil {
		t.Errorf("Server.listenerManager is nil")
	}
	
	if server.console == nil {
		t.Errorf("Server.console is nil")
	}
}

// TestServerGetters tests the getter methods of the server
func TestServerGetters(t *testing.T) {
	server := NewServer()
	
	clientManager := server.GetClientManager()
	if clientManager == nil {
		t.Errorf("GetClientManager() returned nil")
	}
	
	heartbeatMonitor := server.GetHeartbeatMonitor()
	if heartbeatMonitor == nil {
		t.Errorf("GetHeartbeatMonitor() returned nil")
	}
	
	listenerManager := server.GetListenerManager()
	if listenerManager == nil {
		t.Errorf("GetListenerManager() returned nil")
	}
}

// TestServerStartStop tests the start and stop methods of the server
func TestServerStartStop(t *testing.T) {
	server := NewServer()
	
	// Replace the console with a mock that doesn't block
	mockConsole := NewConsole(server.clientManager, server.heartbeatMonitor)
	mockConsole.running = true
	server.console = mockConsole
	
	// Start the server in a goroutine
	go func() {
		err := server.Start()
		if err != nil {
			t.Errorf("Server.Start() failed: %v", err)
		}
	}()
	
	// Give the server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Stop the server
	server.Stop()
	
	// Check that the console is not running
	if mockConsole.running {
		t.Errorf("Console should not be running after server stop")
	}
}
