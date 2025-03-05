package integration

import (
	"fmt"
	"time"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	// ServerHost is the hostname or IP address of the server
	ServerHost string
	
	// ServerPort is the port number the server is listening on
	ServerPort int
	
	// Protocol is the communication protocol to use (tcp, udp, ws, icmp, dns)
	Protocol string
	
	// NumClients is the number of clients to simulate
	NumClients int
	
	// TestDuration is the duration of the test in seconds
	TestDuration time.Duration
	
	// HeartbeatDelay is the delay between heartbeats in seconds
	HeartbeatDelay time.Duration
	
	// ModulesToTest is a list of modules to test loading and unloading
	ModulesToTest []string
	
	// EncryptionType is the type of encryption to use (aes, chacha20)
	EncryptionType string
}

// DefaultTestConfig returns the default test configuration
func DefaultTestConfig() TestConfig {
	return TestConfig{
		ServerHost:     "localhost",
		ServerPort:     8080,
		Protocol:       "tcp",
		NumClients:     5,
		TestDuration:   60 * time.Second,
		HeartbeatDelay: 5 * time.Second,
		ModulesToTest:  []string{"shell", "file", "screenshot"},
		EncryptionType: "aes",
	}
}

// GetServerAddress returns the full server address with protocol, host and port
func (c *TestConfig) GetServerAddress() string {
	return c.Protocol + "://" + c.ServerHost + ":" + fmt.Sprintf("%d", c.ServerPort)
}
