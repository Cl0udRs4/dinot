package client

import (
	"testing"
	"time"
)

func TestNewClientManager(t *testing.T) {
	manager := NewClientManager()
	
	if manager == nil {
		t.Fatal("Expected NewClientManager to return a non-nil manager")
	}
	
	if manager.clients == nil {
		t.Fatal("Expected manager.clients to be initialized")
	}
	
	if count := manager.Count(); count != 0 {
		t.Errorf("Expected new manager to have 0 clients, got %d", count)
	}
}

func TestClientManagerRegisterClient(t *testing.T) {
	manager := NewClientManager()
	
	// Create a test client
	client := NewClient("test-client-1", "Test Client", "192.168.1.100", "linux", "amd64", []string{"shell"}, "tcp")
	
	// Register the client
	err := manager.RegisterClient(client)
	if err != nil {
		t.Fatalf("Failed to register client: %v", err)
	}
	
	// Verify the client was registered
	if count := manager.Count(); count != 1 {
		t.Errorf("Expected manager to have 1 client, got %d", count)
	}
	
	// Try to register the same client again
	err = manager.RegisterClient(client)
	if err != ErrClientAlreadyExists {
		t.Errorf("Expected ErrClientAlreadyExists, got %v", err)
	}
	
	// Verify the count didn't change
	if count := manager.Count(); count != 1 {
		t.Errorf("Expected manager to still have 1 client, got %d", count)
	}
}

func TestClientManagerUnregisterClient(t *testing.T) {
	manager := NewClientManager()
	
	// Create and register a test client
	client := NewClient("test-client-2", "Test Client", "192.168.1.101", "windows", "x86", []string{"shell"}, "udp")
	err := manager.RegisterClient(client)
	if err != nil {
		t.Fatalf("Failed to register client: %v", err)
	}
	
	// Unregister the client
	err = manager.UnregisterClient(client.ID)
	if err != nil {
		t.Fatalf("Failed to unregister client: %v", err)
	}
	
	// Verify the client was unregistered
	if count := manager.Count(); count != 0 {
		t.Errorf("Expected manager to have 0 clients, got %d", count)
	}
	
	// Try to unregister a non-existent client
	err = manager.UnregisterClient("non-existent")
	if err != ErrClientNotFound {
		t.Errorf("Expected ErrClientNotFound, got %v", err)
	}
}

func TestClientManagerGetClient(t *testing.T) {
	manager := NewClientManager()
	
	// Create and register a test client
	client := NewClient("test-client-3", "Test Client", "192.168.1.102", "macos", "arm64", []string{"shell"}, "ws")
	err := manager.RegisterClient(client)
	if err != nil {
		t.Fatalf("Failed to register client: %v", err)
	}
	
	// Get the client
	retrievedClient, err := manager.GetClient(client.ID)
	if err != nil {
		t.Fatalf("Failed to get client: %v", err)
	}
	
	// Verify it's the same client
	if retrievedClient != client {
		t.Errorf("Retrieved client is not the same as the original client")
	}
	
	// Try to get a non-existent client
	_, err = manager.GetClient("non-existent")
	if err != ErrClientNotFound {
		t.Errorf("Expected ErrClientNotFound, got %v", err)
	}
}

func TestClientManagerGetAllClients(t *testing.T) {
	manager := NewClientManager()
	
	// Create and register multiple clients
	client1 := NewClient("test-client-4", "Test Client 1", "192.168.1.103", "linux", "amd64", []string{"shell"}, "tcp")
	client2 := NewClient("test-client-5", "Test Client 2", "192.168.1.104", "windows", "x86", []string{"shell"}, "udp")
	client3 := NewClient("test-client-6", "Test Client 3", "192.168.1.105", "macos", "arm64", []string{"shell"}, "ws")
	
	manager.RegisterClient(client1)
	manager.RegisterClient(client2)
	manager.RegisterClient(client3)
	
	// Get all clients
	clients := manager.GetAllClients()
	
	// Verify the count
	if len(clients) != 3 {
		t.Errorf("Expected 3 clients, got %d", len(clients))
	}
	
	// Verify all clients are in the list
	foundClient1 := false
	foundClient2 := false
	foundClient3 := false
	
	for _, client := range clients {
		switch client.ID {
		case client1.ID:
			foundClient1 = true
		case client2.ID:
			foundClient2 = true
		case client3.ID:
			foundClient3 = true
		}
	}
	
	if !foundClient1 || !foundClient2 || !foundClient3 {
		t.Errorf("Not all clients were returned: found1=%v, found2=%v, found3=%v", 
			foundClient1, foundClient2, foundClient3)
	}
}

func TestClientManagerGetClientsByStatus(t *testing.T) {
	manager := NewClientManager()
	
	// Create and register multiple clients with different statuses
	client1 := NewClient("test-client-7", "Test Client 1", "192.168.1.106", "linux", "amd64", []string{"shell"}, "tcp")
	client2 := NewClient("test-client-8", "Test Client 2", "192.168.1.107", "windows", "x86", []string{"shell"}, "udp")
	client3 := NewClient("test-client-9", "Test Client 3", "192.168.1.108", "macos", "arm64", []string{"shell"}, "ws")
	
	manager.RegisterClient(client1)
	manager.RegisterClient(client2)
	manager.RegisterClient(client3)
	
	// Set different statuses
	client1.UpdateStatus(StatusOnline, "")
	client2.UpdateStatus(StatusBusy, "")
	client3.UpdateStatus(StatusOffline, "")
	
	// Get clients by status
	onlineClients := manager.GetClientsByStatus(StatusOnline)
	busyClients := manager.GetClientsByStatus(StatusBusy)
	offlineClients := manager.GetClientsByStatus(StatusOffline)
	errorClients := manager.GetClientsByStatus(StatusError)
	
	// Verify counts
	if len(onlineClients) != 1 {
		t.Errorf("Expected 1 online client, got %d", len(onlineClients))
	}
	
	if len(busyClients) != 1 {
		t.Errorf("Expected 1 busy client, got %d", len(busyClients))
	}
	
	if len(offlineClients) != 1 {
		t.Errorf("Expected 1 offline client, got %d", len(offlineClients))
	}
	
	if len(errorClients) != 0 {
		t.Errorf("Expected 0 error clients, got %d", len(errorClients))
	}
	
	// Verify the correct clients are returned
	if len(onlineClients) > 0 && onlineClients[0].ID != client1.ID {
		t.Errorf("Expected online client to be %s, got %s", client1.ID, onlineClients[0].ID)
	}
	
	if len(busyClients) > 0 && busyClients[0].ID != client2.ID {
		t.Errorf("Expected busy client to be %s, got %s", client2.ID, busyClients[0].ID)
	}
	
	if len(offlineClients) > 0 && offlineClients[0].ID != client3.ID {
		t.Errorf("Expected offline client to be %s, got %s", client3.ID, offlineClients[0].ID)
	}
}

func TestClientManagerUpdateClientStatus(t *testing.T) {
	manager := NewClientManager()
	
	// Create and register a test client
	client := NewClient("test-client-10", "Test Client", "192.168.1.109", "linux", "amd64", []string{"shell"}, "tcp")
	err := manager.RegisterClient(client)
	if err != nil {
		t.Fatalf("Failed to register client: %v", err)
	}
	
	// Update the client's status
	err = manager.UpdateClientStatus(client.ID, StatusBusy, "")
	if err != nil {
		t.Fatalf("Failed to update client status: %v", err)
	}
	
	// Verify the status was updated
	if client.Status != StatusBusy {
		t.Errorf("Expected client status to be %s, got %s", StatusBusy, client.Status)
	}
	
	// Update to error status with message
	errorMsg := "Connection lost"
	err = manager.UpdateClientStatus(client.ID, StatusError, errorMsg)
	if err != nil {
		t.Fatalf("Failed to update client status: %v", err)
	}
	
	// Verify the status and error message were updated
	if client.Status != StatusError {
		t.Errorf("Expected client status to be %s, got %s", StatusError, client.Status)
	}
	
	if client.ErrorMessage != errorMsg {
		t.Errorf("Expected error message to be %s, got %s", errorMsg, client.ErrorMessage)
	}
	
	// Try to update a non-existent client
	err = manager.UpdateClientStatus("non-existent", StatusOnline, "")
	if err != ErrClientNotFound {
		t.Errorf("Expected ErrClientNotFound, got %v", err)
	}
}

func TestClientManagerUpdateClientLastSeen(t *testing.T) {
	manager := NewClientManager()
	
	// Create and register a test client
	client := NewClient("test-client-11", "Test Client", "192.168.1.110", "linux", "amd64", []string{"shell"}, "tcp")
	err := manager.RegisterClient(client)
	if err != nil {
		t.Fatalf("Failed to register client: %v", err)
	}
	
	// Record the initial last seen time
	initialLastSeen := client.LastSeen
	
	// Wait a short time
	time.Sleep(10 * time.Millisecond)
	
	// Update last seen
	err = manager.UpdateClientLastSeen(client.ID)
	if err != nil {
		t.Fatalf("Failed to update client last seen: %v", err)
	}
	
	// Verify last seen was updated
	if !client.LastSeen.After(initialLastSeen) {
		t.Errorf("Expected LastSeen to be updated to a later time")
	}
	
	// Try to update a non-existent client
	err = manager.UpdateClientLastSeen("non-existent")
	if err != ErrClientNotFound {
		t.Errorf("Expected ErrClientNotFound, got %v", err)
	}
}

func TestClientManagerCheckOfflineClients(t *testing.T) {
	manager := NewClientManager()
	
	// Create and register multiple clients
	client1 := NewClient("test-client-12", "Test Client 1", "192.168.1.111", "linux", "amd64", []string{"shell"}, "tcp")
	client2 := NewClient("test-client-13", "Test Client 2", "192.168.1.112", "windows", "x86", []string{"shell"}, "udp")
	
	manager.RegisterClient(client1)
	manager.RegisterClient(client2)
	
	// Set different heartbeat intervals
	client1.SetHeartbeatInterval(100 * time.Millisecond)
	client2.SetHeartbeatInterval(5 * time.Second)
	
	// Update client1's last seen to be in the past
	pastTime := time.Now().Add(-1 * time.Second)
	client1.mu.Lock()
	client1.LastSeen = pastTime
	client1.mu.Unlock()
	
	// Check for offline clients with a small timeout
	timeout := 500 * time.Millisecond
	offlineClients := manager.CheckOfflineClients(timeout)
	
	// Verify client1 is marked as offline
	if len(offlineClients) != 1 {
		t.Errorf("Expected 1 offline client, got %d", len(offlineClients))
	} else if offlineClients[0].ID != client1.ID {
		t.Errorf("Expected offline client to be %s, got %s", client1.ID, offlineClients[0].ID)
	}
	
	// Verify client1's status is now offline
	if client1.Status != StatusOffline {
		t.Errorf("Expected client1 status to be %s, got %s", StatusOffline, client1.Status)
	}
	
	// Verify client2's status is still online
	if client2.Status != StatusOnline {
		t.Errorf("Expected client2 status to be %s, got %s", StatusOnline, client2.Status)
	}
}

func TestClientManagerCountByStatus(t *testing.T) {
	manager := NewClientManager()
	
	// Create and register multiple clients with different statuses
	client1 := NewClient("test-client-14", "Test Client 1", "192.168.1.113", "linux", "amd64", []string{"shell"}, "tcp")
	client2 := NewClient("test-client-15", "Test Client 2", "192.168.1.114", "windows", "x86", []string{"shell"}, "udp")
	client3 := NewClient("test-client-16", "Test Client 3", "192.168.1.115", "macos", "arm64", []string{"shell"}, "ws")
	client4 := NewClient("test-client-17", "Test Client 4", "192.168.1.116", "linux", "arm", []string{"shell"}, "dns")
	
	manager.RegisterClient(client1)
	manager.RegisterClient(client2)
	manager.RegisterClient(client3)
	manager.RegisterClient(client4)
	
	// Set different statuses
	client1.UpdateStatus(StatusOnline, "")
	client2.UpdateStatus(StatusBusy, "")
	client3.UpdateStatus(StatusOffline, "")
	client4.UpdateStatus(StatusOnline, "")
	
	// Count by status
	onlineCount := manager.CountByStatus(StatusOnline)
	busyCount := manager.CountByStatus(StatusBusy)
	offlineCount := manager.CountByStatus(StatusOffline)
	errorCount := manager.CountByStatus(StatusError)
	
	// Verify counts
	if onlineCount != 2 {
		t.Errorf("Expected 2 online clients, got %d", onlineCount)
	}
	
	if busyCount != 1 {
		t.Errorf("Expected 1 busy client, got %d", busyCount)
	}
	
	if offlineCount != 1 {
		t.Errorf("Expected 1 offline client, got %d", offlineCount)
	}
	
	if errorCount != 0 {
		t.Errorf("Expected 0 error clients, got %d", errorCount)
	}
}
