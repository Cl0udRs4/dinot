package protocol

import (
	"context"
	"os"
	"testing"
)

func TestICMPProtocol_New(t *testing.T) {
	config := Config{
		ServerAddress:  "127.0.0.1",
		ConnectTimeout: 10,
		ReadTimeout:    10,
		WriteTimeout:   10,
		RetryCount:     3,
		RetryInterval:  2,
		BufferSize:     4096,
	}

	icmp := NewICMPProtocol(config)

	if icmp.GetName() != "icmp" {
		t.Errorf("Expected protocol name to be 'icmp', got '%s'", icmp.GetName())
	}

	if icmp.GetStatus() != StatusDisconnected {
		t.Errorf("Expected initial status to be 'disconnected', got '%s'", icmp.GetStatus())
	}

	if icmp.IsConnected() {
		t.Error("Expected IsConnected() to return false initially")
	}

	if icmp.GetLastError() != nil {
		t.Errorf("Expected initial last error to be nil, got '%v'", icmp.GetLastError())
	}

	gotConfig := icmp.GetConfig()
	if gotConfig.ServerAddress != config.ServerAddress {
		t.Errorf("Expected ServerAddress to be '%s', got '%s'", config.ServerAddress, gotConfig.ServerAddress)
	}

	// Check that the identifier is set to the process ID
	expectedID := os.Getpid() & 0xffff
	if icmp.identifier != expectedID {
		t.Errorf("Expected identifier to be %d, got %d", expectedID, icmp.identifier)
	}

	// Check that the sequence starts at 0
	if icmp.sequence != 0 {
		t.Errorf("Expected initial sequence to be 0, got %d", icmp.sequence)
	}
}

func TestICMPProtocol_ValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "Valid config",
			config: Config{
				ServerAddress:  "127.0.0.1",
				ConnectTimeout: 10,
				ReadTimeout:    10,
				WriteTimeout:   10,
				RetryCount:     3,
				RetryInterval:  2,
				BufferSize:     4096,
			},
			expectError: os.Geteuid() != 0, // Expect error if not root
		},
		{
			name: "Empty server address",
			config: Config{
				ServerAddress:  "",
				ConnectTimeout: 10,
			},
			expectError: true,
		},
		{
			name: "Zero connect timeout",
			config: Config{
				ServerAddress:  "127.0.0.1",
				ConnectTimeout: 0,
			},
			expectError: os.Geteuid() != 0, // Expect error if not root
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icmp := NewICMPProtocol(tt.config)
			err := icmp.ValidateConfig()

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestICMPProtocol_UpdateConfig(t *testing.T) {
	initialConfig := Config{
		ServerAddress:  "127.0.0.1",
		ConnectTimeout: 10,
	}

	newConfig := Config{
		ServerAddress:  "8.8.8.8",
		ConnectTimeout: 20,
	}

	icmp := NewICMPProtocol(initialConfig)

	// Test updating config when disconnected
	err := icmp.UpdateConfig(newConfig)
	if err != nil {
		t.Errorf("Expected no error when updating config while disconnected, got: %v", err)
	}

	gotConfig := icmp.GetConfig()
	if gotConfig.ServerAddress != newConfig.ServerAddress {
		t.Errorf("Expected ServerAddress to be updated to '%s', got '%s'", newConfig.ServerAddress, gotConfig.ServerAddress)
	}

	if gotConfig.ConnectTimeout != newConfig.ConnectTimeout {
		t.Errorf("Expected ConnectTimeout to be updated to %d, got %d", newConfig.ConnectTimeout, gotConfig.ConnectTimeout)
	}

	// Simulate connected state
	icmp.setStatus(StatusConnected)
	icmp.packetConn = &icmp.PacketConn{} // Mock connection

	// Test updating config when connected
	err = icmp.UpdateConfig(initialConfig)
	if err == nil {
		t.Error("Expected error when updating config while connected, got nil")
	}
}

// This test requires root privileges and a running ICMP echo server to connect to
// It's commented out to avoid test failures when no server is available or not running as root
/*
func TestICMPProtocol_ConnectAndDisconnect(t *testing.T) {
	// Skip if not running as root
	if os.Geteuid() != 0 {
		t.Skip("Skipping test: ICMP protocol requires root privileges")
	}

	// Create ICMP protocol with a test server address
	config := Config{
		ServerAddress:  "127.0.0.1", // Use localhost for testing
		ConnectTimeout: 5,
		ReadTimeout:    5,
		WriteTimeout:   5,
		RetryCount:     1,
		RetryInterval:  1,
		BufferSize:     1024,
	}

	icmp := NewICMPProtocol(config)

	// Test Connect
	ctx := context.Background()
	err := icmp.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	if !icmp.IsConnected() {
		t.Error("Expected IsConnected() to return true after successful connection")
	}

	if icmp.GetStatus() != StatusConnected {
		t.Errorf("Expected status to be 'connected', got '%s'", icmp.GetStatus())
	}

	// Test Send and Receive
	testData := []byte("Hello, ICMP!")
	n, err := icmp.Send(testData)
	if err != nil {
		t.Fatalf("Failed to send data: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected to send %d bytes, sent %d", len(testData), n)
	}

	// Wait for response
	receivedData, err := icmp.Receive()
	if err != nil {
		t.Fatalf("Failed to receive data: %v", err)
	}

	if string(receivedData) != string(testData) {
		t.Errorf("Expected to receive '%s', got '%s'", string(testData), string(receivedData))
	}

	// Test Disconnect
	err = icmp.Disconnect()
	if err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}

	if icmp.IsConnected() {
		t.Error("Expected IsConnected() to return false after disconnection")
	}

	if icmp.GetStatus() != StatusDisconnected {
		t.Errorf("Expected status to be 'disconnected', got '%s'", icmp.GetStatus())
	}
}
*/
