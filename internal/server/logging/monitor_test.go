package logging

import (
	"testing"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/client"
)

// TestMonitorManager tests the MonitorManager implementation
func TestMonitorManager(t *testing.T) {
	// Create a client manager
	clientManager := client.NewClientManager()

	// Create a test client
	testClient := client.NewClient(
		"test-client-id",
		"Test Client",
		"192.168.1.100",
		"Linux",
		"x86_64",
		[]string{"shell", "file", "process"},
		"tcp",
	)

	// Register the client
	err := clientManager.RegisterClient(testClient)
	if err != nil {
		t.Fatalf("Failed to register client: %v", err)
	}

	// Create a logger
	logger := NewLogrusLogger()

	// Create a monitor config
	config := MonitorConfig{
		CheckInterval:        100 * time.Millisecond,
		ReconnectInterval:    50 * time.Millisecond,
		MaxReconnectAttempts: 3,
	}

	// Create a monitor manager
	monitor := NewMonitorManager(logger, clientManager, config)

	// Start the monitor
	monitor.Start()

	// Set the client status to error
	err = clientManager.UpdateClientStatus("test-client-id", client.StatusError, "Test error")
	if err != nil {
		t.Fatalf("Failed to update client status: %v", err)
	}

	// Report an exception
	_, err = clientManager.ReportException(
		"test-client-id",
		"Test exception",
		client.SeverityError,
		"test-module",
		"test stack trace",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to report exception: %v", err)
	}

	// Wait for the monitor to detect the exception and attempt reconnection
	time.Sleep(500 * time.Millisecond)

	// Check if the client status was updated
	updatedClient, err := clientManager.GetClient("test-client-id")
	if err != nil {
		t.Fatalf("Failed to get client: %v", err)
	}

	if updatedClient.Status != client.StatusOnline {
		t.Errorf("Expected client status to be online, got %s", updatedClient.Status)
	}

	// Stop the monitor
	monitor.Stop()
}
