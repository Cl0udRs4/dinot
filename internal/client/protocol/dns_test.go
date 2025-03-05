package protocol

import (
	"context"
	"testing"
)

func TestDNSProtocol_New(t *testing.T) {
	config := Config{
		ServerAddress:  "example.com@8.8.8.8",
		ConnectTimeout: 10,
		ReadTimeout:    10,
		WriteTimeout:   10,
		RetryCount:     3,
		RetryInterval:  2,
		BufferSize:     4096,
	}

	dns := NewDNSProtocol(config)

	if dns.GetName() != "dns" {
		t.Errorf("Expected protocol name to be 'dns', got '%s'", dns.GetName())
	}

	if dns.GetStatus() != StatusDisconnected {
		t.Errorf("Expected initial status to be 'disconnected', got '%s'", dns.GetStatus())
	}

	if dns.IsConnected() {
		t.Error("Expected IsConnected() to return false initially")
	}

	if dns.GetLastError() != nil {
		t.Errorf("Expected initial last error to be nil, got '%v'", dns.GetLastError())
	}

	gotConfig := dns.GetConfig()
	if gotConfig.ServerAddress != config.ServerAddress {
		t.Errorf("Expected ServerAddress to be '%s', got '%s'", config.ServerAddress, gotConfig.ServerAddress)
	}

	// Check default query type
	if dns.queryType != 16 { // 16 is the value for dns.TypeTXT
		t.Errorf("Expected default query type to be TXT (16), got %d", dns.queryType)
	}

	// Check default max data size
	if dns.maxDataSize != 250 {
		t.Errorf("Expected default max data size to be 250, got %d", dns.maxDataSize)
	}
}

func TestDNSProtocol_ValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "Valid config",
			config: Config{
				ServerAddress:  "example.com@8.8.8.8",
				ConnectTimeout: 10,
				ReadTimeout:    10,
				WriteTimeout:   10,
				RetryCount:     3,
				RetryInterval:  2,
				BufferSize:     4096,
			},
			expectError: false,
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
			name: "Invalid server address format",
			config: Config{
				ServerAddress:  "example.com",
				ConnectTimeout: 10,
			},
			expectError: true,
		},
		{
			name: "Empty domain",
			config: Config{
				ServerAddress:  "@8.8.8.8",
				ConnectTimeout: 10,
			},
			expectError: true,
		},
		{
			name: "Empty DNS server",
			config: Config{
				ServerAddress:  "example.com@",
				ConnectTimeout: 10,
			},
			expectError: true,
		},
		{
			name: "Invalid DNS server",
			config: Config{
				ServerAddress:  "example.com@invalid-server",
				ConnectTimeout: 10,
			},
			expectError: true,
		},
		{
			name: "Zero connect timeout",
			config: Config{
				ServerAddress:  "example.com@8.8.8.8",
				ConnectTimeout: 0,
			},
			expectError: false, // Should use default value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dns := NewDNSProtocol(tt.config)
			err := dns.ValidateConfig()

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestDNSProtocol_UpdateConfig(t *testing.T) {
	initialConfig := Config{
		ServerAddress:  "example.com@8.8.8.8",
		ConnectTimeout: 10,
	}

	newConfig := Config{
		ServerAddress:  "example.org@1.1.1.1",
		ConnectTimeout: 20,
	}

	dns := NewDNSProtocol(initialConfig)

	// Test updating config when disconnected
	err := dns.UpdateConfig(newConfig)
	if err != nil {
		t.Errorf("Expected no error when updating config while disconnected, got: %v", err)
	}

	gotConfig := dns.GetConfig()
	if gotConfig.ServerAddress != newConfig.ServerAddress {
		t.Errorf("Expected ServerAddress to be updated to '%s', got '%s'", newConfig.ServerAddress, gotConfig.ServerAddress)
	}

	if gotConfig.ConnectTimeout != newConfig.ConnectTimeout {
		t.Errorf("Expected ConnectTimeout to be updated to %d, got %d", newConfig.ConnectTimeout, gotConfig.ConnectTimeout)
	}

	// Simulate connected state
	dns.setStatus(StatusConnected)

	// Test updating config when connected
	err = dns.UpdateConfig(initialConfig)
	if err == nil {
		t.Error("Expected error when updating config while connected, got nil")
	}
}

func TestDNSProtocol_SetQueryType(t *testing.T) {
	dns := NewDNSProtocol(Config{ServerAddress: "example.com@8.8.8.8"})

	// Test default query type
	if dns.queryType != 16 { // 16 is the value for dns.TypeTXT
		t.Errorf("Expected default query type to be TXT (16), got %d", dns.queryType)
	}

	// Test setting query type
	dns.SetQueryType(1) // 1 is the value for dns.TypeA
	if dns.queryType != 1 {
		t.Errorf("Expected query type to be A (1), got %d", dns.queryType)
	}
}

func TestDNSProtocol_SetMaxDataSize(t *testing.T) {
	dns := NewDNSProtocol(Config{ServerAddress: "example.com@8.8.8.8"})

	// Test default max data size
	if dns.maxDataSize != 250 {
		t.Errorf("Expected default max data size to be 250, got %d", dns.maxDataSize)
	}

	// Test setting max data size
	dns.SetMaxDataSize(100)
	if dns.maxDataSize != 100 {
		t.Errorf("Expected max data size to be 100, got %d", dns.maxDataSize)
	}

	// Test setting invalid max data size
	dns.SetMaxDataSize(0)
	if dns.maxDataSize != 100 {
		t.Errorf("Expected max data size to remain 100, got %d", dns.maxDataSize)
	}
}

func TestSplitString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		chunkSize int
		expected  []string
	}{
		{
			name:      "Empty string",
			input:     "",
			chunkSize: 5,
			expected:  []string{""},
		},
		{
			name:      "String smaller than chunk size",
			input:     "hello",
			chunkSize: 10,
			expected:  []string{"hello"},
		},
		{
			name:      "String equal to chunk size",
			input:     "hello",
			chunkSize: 5,
			expected:  []string{"hello"},
		},
		{
			name:      "String larger than chunk size",
			input:     "hello world",
			chunkSize: 5,
			expected:  []string{"hello", " worl", "d"},
		},
		{
			name:      "Zero chunk size",
			input:     "hello",
			chunkSize: 0,
			expected:  []string{"hello"},
		},
		{
			name:      "Negative chunk size",
			input:     "hello",
			chunkSize: -1,
			expected:  []string{"hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitString(tt.input, tt.chunkSize)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d chunks, got %d", len(tt.expected), len(result))
			}
			for i, chunk := range result {
				if i < len(tt.expected) && chunk != tt.expected[i] {
					t.Errorf("Chunk %d: expected '%s', got '%s'", i, tt.expected[i], chunk)
				}
			}
		})
	}
}

// This test requires a running DNS server to connect to
// It's commented out to avoid test failures when no server is available
/*
func TestDNSProtocol_ConnectAndDisconnect(t *testing.T) {
	// Create DNS protocol with a test server address
	config := Config{
		ServerAddress:  "example.com@8.8.8.8", // Use Google's public DNS for testing
		ConnectTimeout: 5,
		ReadTimeout:    5,
		WriteTimeout:   5,
		RetryCount:     1,
		RetryInterval:  1,
		BufferSize:     1024,
	}

	dns := NewDNSProtocol(config)

	// Test Connect
	ctx := context.Background()
	err := dns.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	if !dns.IsConnected() {
		t.Error("Expected IsConnected() to return true after successful connection")
	}

	if dns.GetStatus() != StatusConnected {
		t.Errorf("Expected status to be 'connected', got '%s'", dns.GetStatus())
	}

	// Test Disconnect
	err = dns.Disconnect()
	if err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}

	if dns.IsConnected() {
		t.Error("Expected IsConnected() to return false after disconnection")
	}

	if dns.GetStatus() != StatusDisconnected {
		t.Errorf("Expected status to be 'disconnected', got '%s'", dns.GetStatus())
	}
}
*/
