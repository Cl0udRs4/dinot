package client

import (
	"testing"
	"time"
)

func TestNewHeartbeatMonitor(t *testing.T) {
	clientManager := NewClientManager()
	checkInterval := 100 * time.Millisecond
	timeout := 500 * time.Millisecond
	
	monitor := NewHeartbeatMonitor(clientManager, checkInterval, timeout)
	
	if monitor == nil {
		t.Fatal("Expected NewHeartbeatMonitor to return a non-nil monitor")
	}
	
	if monitor.clientManager != clientManager {
		t.Error("Expected monitor.clientManager to be set correctly")
	}
	
	if monitor.checkInterval != checkInterval {
		t.Errorf("Expected checkInterval to be %v, got %v", checkInterval, monitor.checkInterval)
	}
	
	if monitor.timeout != timeout {
		t.Errorf("Expected timeout to be %v, got %v", timeout, monitor.timeout)
	}
	
	if monitor.useRandomIntervals {
		t.Error("Expected useRandomIntervals to be false by default")
	}
}

func TestHeartbeatMonitorSetCheckInterval(t *testing.T) {
	clientManager := NewClientManager()
	monitor := NewHeartbeatMonitor(clientManager, 100*time.Millisecond, 500*time.Millisecond)
	
	// Set a new check interval
	newInterval := 200 * time.Millisecond
	monitor.SetCheckInterval(newInterval)
	
	// Verify the interval was updated
	if monitor.checkInterval != newInterval {
		t.Errorf("Expected checkInterval to be %v, got %v", newInterval, monitor.checkInterval)
	}
}

func TestHeartbeatMonitorSetTimeout(t *testing.T) {
	clientManager := NewClientManager()
	monitor := NewHeartbeatMonitor(clientManager, 100*time.Millisecond, 500*time.Millisecond)
	
	// Set a new timeout
	newTimeout := 1 * time.Second
	monitor.SetTimeout(newTimeout)
	
	// Verify the timeout was updated
	if monitor.timeout != newTimeout {
		t.Errorf("Expected timeout to be %v, got %v", newTimeout, monitor.timeout)
	}
}

func TestHeartbeatMonitorRandomIntervals(t *testing.T) {
	clientManager := NewClientManager()
	monitor := NewHeartbeatMonitor(clientManager, 100*time.Millisecond, 500*time.Millisecond)
	
	// Enable random intervals
	minInterval := 1 * time.Second
	maxInterval := 5 * time.Second
	monitor.EnableRandomIntervals(minInterval, maxInterval)
	
	// Verify settings were updated
	if !monitor.useRandomIntervals {
		t.Error("Expected useRandomIntervals to be true after enabling")
	}
	
	if monitor.minRandomInterval != minInterval {
		t.Errorf("Expected minRandomInterval to be %v, got %v", minInterval, monitor.minRandomInterval)
	}
	
	if monitor.maxRandomInterval != maxInterval {
		t.Errorf("Expected maxRandomInterval to be %v, got %v", maxInterval, monitor.maxRandomInterval)
	}
	
	// Disable random intervals
	monitor.DisableRandomIntervals()
	
	// Verify setting was updated
	if monitor.useRandomIntervals {
		t.Error("Expected useRandomIntervals to be false after disabling")
	}
}

func TestHeartbeatMonitorAssignRandomInterval(t *testing.T) {
	clientManager := NewClientManager()
	monitor := NewHeartbeatMonitor(clientManager, 100*time.Millisecond, 500*time.Millisecond)
	
	// Create and register a test client
	client := NewClient("test-client-1", "Test Client", "192.168.1.100", "linux", "amd64", []string{"shell"}, "tcp")
	clientManager.RegisterClient(client)
	
	// Set the initial heartbeat interval
	initialInterval := 60 * time.Second
	client.SetHeartbeatInterval(initialInterval)
	
	// Try to assign a random interval with random intervals disabled
	err := monitor.AssignRandomInterval(client.ID)
	if err != nil {
		t.Fatalf("Failed to assign random interval: %v", err)
	}
	
	// Verify the interval didn't change
	if client.HeartbeatInterval != initialInterval {
		t.Errorf("Expected heartbeat interval to remain %v, got %v", initialInterval, client.HeartbeatInterval)
	}
	
	// Enable random intervals
	minInterval := 1 * time.Second
	maxInterval := 5 * time.Second
	monitor.EnableRandomIntervals(minInterval, maxInterval)
	
	// Assign a random interval
	err = monitor.AssignRandomInterval(client.ID)
	if err != nil {
		t.Fatalf("Failed to assign random interval: %v", err)
	}
	
	// Verify the interval changed and is within the expected range
	if client.HeartbeatInterval == initialInterval {
		t.Error("Expected heartbeat interval to change")
	}
	
	if client.HeartbeatInterval < minInterval || client.HeartbeatInterval > maxInterval {
		t.Errorf("Expected heartbeat interval to be between %v and %v, got %v", 
			minInterval, maxInterval, client.HeartbeatInterval)
	}
	
	// Try to assign a random interval to a non-existent client
	err = monitor.AssignRandomInterval("non-existent")
	if err != ErrClientNotFound {
		t.Errorf("Expected ErrClientNotFound, got %v", err)
	}
}

func TestHeartbeatMonitorProcessHeartbeat(t *testing.T) {
	clientManager := NewClientManager()
	monitor := NewHeartbeatMonitor(clientManager, 100*time.Millisecond, 500*time.Millisecond)
	
	// Create and register a test client
	client := NewClient("test-client-2", "Test Client", "192.168.1.101", "linux", "amd64", []string{"shell"}, "tcp")
	clientManager.RegisterClient(client)
	
	// Set the client to offline
	client.UpdateStatus(StatusOffline, "Heartbeat timeout exceeded")
	
	// Record the initial last seen time
	initialLastSeen := client.LastSeen
	
	// Wait a short time
	time.Sleep(10 * time.Millisecond)
	
	// Process a heartbeat
	err := monitor.ProcessHeartbeat(client.ID)
	if err != nil {
		t.Fatalf("Failed to process heartbeat: %v", err)
	}
	
	// Verify the client is now online
	if client.Status != StatusOnline {
		t.Errorf("Expected client status to be %s, got %s", StatusOnline, client.Status)
	}
	
	// Verify the last seen time was updated
	if !client.LastSeen.After(initialLastSeen) {
		t.Error("Expected LastSeen to be updated to a later time")
	}
	
	// Enable random intervals
	minInterval := 1 * time.Second
	maxInterval := 5 * time.Second
	monitor.EnableRandomIntervals(minInterval, maxInterval)
	
	// Set a fixed heartbeat interval
	fixedInterval := 60 * time.Second
	client.SetHeartbeatInterval(fixedInterval)
	
	// Process another heartbeat
	err = monitor.ProcessHeartbeat(client.ID)
	if err != nil {
		t.Fatalf("Failed to process heartbeat: %v", err)
	}
	
	// Verify the heartbeat interval was changed
	if client.HeartbeatInterval == fixedInterval {
		t.Error("Expected heartbeat interval to change when random intervals are enabled")
	}
	
	// Try to process a heartbeat for a non-existent client
	err = monitor.ProcessHeartbeat("non-existent")
	if err != ErrClientNotFound {
		t.Errorf("Expected ErrClientNotFound, got %v", err)
	}
}

func TestHeartbeatMonitorStartStop(t *testing.T) {
	clientManager := NewClientManager()
	
	// Use a very short check interval for testing
	checkInterval := 10 * time.Millisecond
	timeout := 50 * time.Millisecond
	
	monitor := NewHeartbeatMonitor(clientManager, checkInterval, timeout)
	
	// Start the monitor
	monitor.Start()
	
	// Create and register a test client
	client := NewClient("test-client-3", "Test Client", "192.168.1.102", "linux", "amd64", []string{"shell"}, "tcp")
	client.SetHeartbeatInterval(20 * time.Millisecond)
	clientManager.RegisterClient(client)
	
	// Update the client's last seen to be in the past
	pastTime := time.Now().Add(-100 * time.Millisecond)
	client.mu.Lock()
	client.LastSeen = pastTime
	client.mu.Unlock()
	
	// Wait for the monitor to check clients
	time.Sleep(100 * time.Millisecond)
	
	// Verify the client was marked as offline
	if client.Status != StatusOffline {
		t.Errorf("Expected client status to be %s, got %s", StatusOffline, client.Status)
	}
	
	// Stop the monitor
	monitor.Stop()
	
	// Reset the client status
	client.UpdateStatus(StatusOnline, "")
	
	// Update the client's last seen to be in the past again
	client.mu.Lock()
	client.LastSeen = pastTime
	client.mu.Unlock()
	
	// Wait for what would have been another check
	time.Sleep(100 * time.Millisecond)
	
	// Verify the client status didn't change after stopping the monitor
	if client.Status != StatusOnline {
		t.Errorf("Expected client status to remain %s after stopping monitor, got %s", 
			StatusOnline, client.Status)
	}
}
