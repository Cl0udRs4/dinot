package client

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	// Test creating a new client
	id := "test-client-1"
	name := "Test Client"
	ip := "192.168.1.100"
	os := "linux"
	arch := "amd64"
	modules := []string{"shell", "file", "process"}
	protocol := "tcp"
	
	client := NewClient(id, name, ip, os, arch, modules, protocol)
	
	// Verify client properties
	if client.ID != id {
		t.Errorf("Expected client ID to be %s, got %s", id, client.ID)
	}
	
	if client.Name != name {
		t.Errorf("Expected client name to be %s, got %s", name, client.Name)
	}
	
	if client.IPAddress != ip {
		t.Errorf("Expected client IP to be %s, got %s", ip, client.IPAddress)
	}
	
	if client.OS != os {
		t.Errorf("Expected client OS to be %s, got %s", os, client.OS)
	}
	
	if client.Architecture != arch {
		t.Errorf("Expected client architecture to be %s, got %s", arch, client.Architecture)
	}
	
	if client.Protocol != protocol {
		t.Errorf("Expected client protocol to be %s, got %s", protocol, client.Protocol)
	}
	
	if client.Status != StatusOnline {
		t.Errorf("Expected client status to be %s, got %s", StatusOnline, client.Status)
	}
	
	if len(client.SupportedModules) != len(modules) {
		t.Errorf("Expected %d supported modules, got %d", len(modules), len(client.SupportedModules))
	}
	
	if len(client.ActiveModules) != 0 {
		t.Errorf("Expected 0 active modules, got %d", len(client.ActiveModules))
	}
}

func TestClientUpdateStatus(t *testing.T) {
	client := NewClient("test-client-2", "Test Client", "192.168.1.101", "windows", "x86", []string{"shell"}, "udp")
	
	// Test updating status to busy
	client.UpdateStatus(StatusBusy, "")
	if client.Status != StatusBusy {
		t.Errorf("Expected client status to be %s, got %s", StatusBusy, client.Status)
	}
	
	// Test updating status to error with message
	errorMsg := "Connection lost"
	client.UpdateStatus(StatusError, errorMsg)
	if client.Status != StatusError {
		t.Errorf("Expected client status to be %s, got %s", StatusError, client.Status)
	}
	if client.ErrorMessage != errorMsg {
		t.Errorf("Expected error message to be %s, got %s", errorMsg, client.ErrorMessage)
	}
	
	// Test updating status to online clears error message
	client.UpdateStatus(StatusOnline, "")
	if client.Status != StatusOnline {
		t.Errorf("Expected client status to be %s, got %s", StatusOnline, client.Status)
	}
	if client.ErrorMessage != "" {
		t.Errorf("Expected error message to be empty, got %s", client.ErrorMessage)
	}
}

func TestClientUpdateLastSeen(t *testing.T) {
	client := NewClient("test-client-3", "Test Client", "192.168.1.102", "macos", "arm64", []string{"shell"}, "ws")
	
	// Record the initial last seen time
	initialLastSeen := client.LastSeen
	
	// Wait a short time
	time.Sleep(10 * time.Millisecond)
	
	// Update last seen
	client.UpdateLastSeen()
	
	// Verify last seen was updated
	if !client.LastSeen.After(initialLastSeen) {
		t.Errorf("Expected LastSeen to be updated to a later time")
	}
}

func TestClientSetHeartbeatInterval(t *testing.T) {
	client := NewClient("test-client-4", "Test Client", "192.168.1.103", "linux", "arm", []string{"shell"}, "tcp")
	
	// Default interval should be 60 seconds
	if client.HeartbeatInterval != 60*time.Second {
		t.Errorf("Expected default heartbeat interval to be 60s, got %v", client.HeartbeatInterval)
	}
	
	// Set a new interval
	newInterval := 30 * time.Second
	client.SetHeartbeatInterval(newInterval)
	
	// Verify the interval was updated
	if client.HeartbeatInterval != newInterval {
		t.Errorf("Expected heartbeat interval to be %v, got %v", newInterval, client.HeartbeatInterval)
	}
}

func TestClientModuleManagement(t *testing.T) {
	supportedModules := []string{"shell", "file", "process"}
	client := NewClient("test-client-5", "Test Client", "192.168.1.104", "linux", "amd64", supportedModules, "tcp")
	
	// Test adding a supported module
	success := client.AddActiveModule("shell")
	if !success {
		t.Errorf("Expected AddActiveModule to return true for supported module")
	}
	if len(client.ActiveModules) != 1 || client.ActiveModules[0] != "shell" {
		t.Errorf("Expected ActiveModules to contain ['shell'], got %v", client.ActiveModules)
	}
	
	// Test adding an unsupported module
	success = client.AddActiveModule("unsupported")
	if success {
		t.Errorf("Expected AddActiveModule to return false for unsupported module")
	}
	if len(client.ActiveModules) != 1 {
		t.Errorf("Expected ActiveModules length to remain 1, got %d", len(client.ActiveModules))
	}
	
	// Test adding a duplicate module
	success = client.AddActiveModule("shell")
	if !success {
		t.Errorf("Expected AddActiveModule to return true for duplicate module")
	}
	if len(client.ActiveModules) != 1 {
		t.Errorf("Expected ActiveModules length to remain 1, got %d", len(client.ActiveModules))
	}
	
	// Test adding another supported module
	success = client.AddActiveModule("file")
	if !success {
		t.Errorf("Expected AddActiveModule to return true for supported module")
	}
	if len(client.ActiveModules) != 2 {
		t.Errorf("Expected ActiveModules length to be 2, got %d", len(client.ActiveModules))
	}
	
	// Test checking if a module is active
	if !client.IsModuleActive("shell") {
		t.Errorf("Expected IsModuleActive to return true for active module")
	}
	if client.IsModuleActive("process") {
		t.Errorf("Expected IsModuleActive to return false for inactive module")
	}
	
	// Test removing an active module
	success = client.RemoveActiveModule("shell")
	if !success {
		t.Errorf("Expected RemoveActiveModule to return true for active module")
	}
	if len(client.ActiveModules) != 1 || client.ActiveModules[0] != "file" {
		t.Errorf("Expected ActiveModules to contain ['file'], got %v", client.ActiveModules)
	}
	
	// Test removing an inactive module
	success = client.RemoveActiveModule("shell")
	if success {
		t.Errorf("Expected RemoveActiveModule to return false for inactive module")
	}
	if len(client.ActiveModules) != 1 {
		t.Errorf("Expected ActiveModules length to remain 1, got %d", len(client.ActiveModules))
	}
}

func TestClientJSON(t *testing.T) {
	// Create a client
	client := NewClient("test-client-6", "Test Client", "192.168.1.105", "linux", "amd64", []string{"shell", "file"}, "tcp")
	client.AddActiveModule("shell")
	
	// Convert to JSON
	jsonData, err := client.ToJSON()
	if err != nil {
		t.Fatalf("Failed to convert client to JSON: %v", err)
	}
	
	// Create a new client from the JSON
	newClient := &Client{}
	err = newClient.FromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse client from JSON: %v", err)
	}
	
	// Verify the new client has the same properties
	if newClient.ID != client.ID {
		t.Errorf("Expected client ID to be %s, got %s", client.ID, newClient.ID)
	}
	
	if newClient.Name != client.Name {
		t.Errorf("Expected client name to be %s, got %s", client.Name, newClient.Name)
	}
	
	if newClient.Status != client.Status {
		t.Errorf("Expected client status to be %s, got %s", client.Status, newClient.Status)
	}
	
	if len(newClient.SupportedModules) != len(client.SupportedModules) {
		t.Errorf("Expected %d supported modules, got %d", len(client.SupportedModules), len(newClient.SupportedModules))
	}
	
	if len(newClient.ActiveModules) != len(client.ActiveModules) {
		t.Errorf("Expected %d active modules, got %d", len(client.ActiveModules), len(newClient.ActiveModules))
	}
}
