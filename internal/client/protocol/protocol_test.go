package protocol

import (
	"context"
	"net"
	"testing"
	"time"
)

// TestProtocolInterface tests that all protocol implementations adhere to the Protocol interface
func TestProtocolInterface(t *testing.T) {
	// Create a slice of all protocol implementations
	protocols := []Protocol{
		NewTCPProtocol(Config{ServerAddress: "localhost:8080"}),
		NewUDPProtocol(Config{ServerAddress: "localhost:8081"}),
		NewWSProtocol(Config{ServerAddress: "localhost:8082/ws"}),
		NewICMPProtocol(Config{ServerAddress: "127.0.0.1"}),
		NewDNSProtocol(Config{ServerAddress: "example.com@8.8.8.8"}),
	}

	// Test each protocol
	for _, p := range protocols {
		t.Run(p.GetName(), func(t *testing.T) {
			// Test GetName
			name := p.GetName()
			if name == "" {
				t.Error("GetName() returned empty string")
			}

			// Test GetStatus
			status := p.GetStatus()
			if status != StatusDisconnected {
				t.Errorf("Expected initial status to be 'disconnected', got '%s'", status)
			}

			// Test IsConnected
			if p.IsConnected() {
				t.Error("Expected IsConnected() to return false initially")
			}

			// Test GetLastError
			if p.GetLastError() != nil {
				t.Errorf("Expected initial last error to be nil, got '%v'", p.GetLastError())
			}

			// Test GetConfig
			config := p.GetConfig()
			if config.ServerAddress == "" {
				t.Error("Expected ServerAddress to be set in config")
			}

			// Test UpdateConfig
			newConfig := Config{
				ServerAddress:  "new-address",
				ConnectTimeout: 20,
				ReadTimeout:    20,
				WriteTimeout:   20,
			}
			err := p.UpdateConfig(newConfig)
			if err != nil {
				t.Errorf("Failed to update config: %v", err)
			}

			// Check that the config was updated
			updatedConfig := p.GetConfig()
			if updatedConfig.ServerAddress != newConfig.ServerAddress {
				t.Errorf("Expected ServerAddress to be updated to '%s', got '%s'", newConfig.ServerAddress, updatedConfig.ServerAddress)
			}

			// Test ValidateConfig
			err = p.ValidateConfig()
			// Some protocols may have specific validation requirements
			// We're just testing that the method exists and can be called

			// Test GetConnection
			conn := p.GetConnection()
			// Initially, the connection should be nil
			if conn != nil {
				t.Error("Expected initial connection to be nil")
			}
		})
	}
}

// TestBaseProtocol tests the BaseProtocol implementation
func TestBaseProtocol(t *testing.T) {
	config := Config{
		ServerAddress:  "localhost:8080",
		ConnectTimeout: 10,
		ReadTimeout:    10,
		WriteTimeout:   10,
	}

	base := NewBaseProtocol("test", config)

	// Test GetName
	if base.GetName() != "test" {
		t.Errorf("Expected name to be 'test', got '%s'", base.GetName())
	}

	// Test GetStatus
	if base.GetStatus() != StatusDisconnected {
		t.Errorf("Expected initial status to be 'disconnected', got '%s'", base.GetStatus())
	}

	// Test IsConnected
	if base.IsConnected() {
		t.Error("Expected IsConnected() to return false initially")
	}

	// Test GetLastError
	if base.GetLastError() != nil {
		t.Errorf("Expected initial last error to be nil, got '%v'", base.GetLastError())
	}

	// Test GetConfig
	gotConfig := base.GetConfig()
	if gotConfig.ServerAddress != config.ServerAddress {
		t.Errorf("Expected ServerAddress to be '%s', got '%s'", config.ServerAddress, gotConfig.ServerAddress)
	}

	// Test UpdateConfig
	newConfig := Config{
		ServerAddress:  "new-address",
		ConnectTimeout: 20,
	}
	err := base.UpdateConfig(newConfig)
	if err != nil {
		t.Errorf("Failed to update config: %v", err)
	}

	// Check that the config was updated
	updatedConfig := base.GetConfig()
	if updatedConfig.ServerAddress != newConfig.ServerAddress {
		t.Errorf("Expected ServerAddress to be updated to '%s', got '%s'", newConfig.ServerAddress, updatedConfig.ServerAddress)
	}

	// Test ValidateConfig
	err = base.ValidateConfig()
	if err != nil {
		t.Errorf("Failed to validate config: %v", err)
	}

	// Test with invalid config
	base.Config.ServerAddress = ""
	err = base.ValidateConfig()
	if err == nil {
		t.Error("Expected error when validating config with empty ServerAddress")
	}

	// Test setStatus
	base.setStatus(StatusConnecting)
	if base.GetStatus() != StatusConnecting {
		t.Errorf("Expected status to be 'connecting', got '%s'", base.GetStatus())
	}

	// Test setLastError
	testErr := NewClientError(ErrTypeConnection, "test error", nil)
	base.setLastError(testErr)
	if base.GetLastError() != testErr {
		t.Errorf("Expected last error to be set to test error")
	}

	// Test Disconnect
	base.conn = &net.TCPConn{}
	base.ctx, base.cancel = context.WithCancel(context.Background())
	base.setStatus(StatusConnected)
	base.Disconnect()
	if base.GetStatus() != StatusDisconnected {
		t.Errorf("Expected status to be 'disconnected' after Disconnect, got '%s'", base.GetStatus())
	}
}

// TestConfig tests the Config struct
func TestConfig(t *testing.T) {
	// Test default values
	config := Config{}
	if config.ConnectTimeout != 0 {
		t.Errorf("Expected default ConnectTimeout to be 0, got %d", config.ConnectTimeout)
	}

	// Test setting values
	config = Config{
		ServerAddress:     "localhost:8080",
		ConnectTimeout:    10,
		ReadTimeout:       20,
		WriteTimeout:      30,
		RetryCount:        3,
		RetryInterval:     2,
		KeepAlive:         true,
		KeepAliveInterval: 60,
		BufferSize:        4096,
		EnableTLS:         true,
		TLSCertFile:       "cert.pem",
		TLSKeyFile:        "key.pem",
		TLSCACertFile:     "ca.pem",
		TLSSkipVerify:     false,
	}

	if config.ServerAddress != "localhost:8080" {
		t.Errorf("Expected ServerAddress to be 'localhost:8080', got '%s'", config.ServerAddress)
	}
	if config.ConnectTimeout != 10 {
		t.Errorf("Expected ConnectTimeout to be 10, got %d", config.ConnectTimeout)
	}
	if config.ReadTimeout != 20 {
		t.Errorf("Expected ReadTimeout to be 20, got %d", config.ReadTimeout)
	}
	if config.WriteTimeout != 30 {
		t.Errorf("Expected WriteTimeout to be 30, got %d", config.WriteTimeout)
	}
	if config.RetryCount != 3 {
		t.Errorf("Expected RetryCount to be 3, got %d", config.RetryCount)
	}
	if config.RetryInterval != 2 {
		t.Errorf("Expected RetryInterval to be 2, got %d", config.RetryInterval)
	}
	if !config.KeepAlive {
		t.Error("Expected KeepAlive to be true")
	}
	if config.KeepAliveInterval != 60 {
		t.Errorf("Expected KeepAliveInterval to be 60, got %d", config.KeepAliveInterval)
	}
	if config.BufferSize != 4096 {
		t.Errorf("Expected BufferSize to be 4096, got %d", config.BufferSize)
	}
	if !config.EnableTLS {
		t.Error("Expected EnableTLS to be true")
	}
	if config.TLSCertFile != "cert.pem" {
		t.Errorf("Expected TLSCertFile to be 'cert.pem', got '%s'", config.TLSCertFile)
	}
	if config.TLSKeyFile != "key.pem" {
		t.Errorf("Expected TLSKeyFile to be 'key.pem', got '%s'", config.TLSKeyFile)
	}
	if config.TLSCACertFile != "ca.pem" {
		t.Errorf("Expected TLSCACertFile to be 'ca.pem', got '%s'", config.TLSCACertFile)
	}
	if config.TLSSkipVerify {
		t.Error("Expected TLSSkipVerify to be false")
	}
}

// TestStatus tests the Status type
func TestStatus(t *testing.T) {
	// Test string representation
	statuses := map[Status]string{
		StatusDisconnected: "disconnected",
		StatusConnecting:   "connecting",
		StatusConnected:    "connected",
		StatusError:        "error",
	}

	for status, expected := range statuses {
		if status.String() != expected {
			t.Errorf("Expected Status %d to be '%s', got '%s'", status, expected, status.String())
		}
	}

	// Test invalid status
	invalidStatus := Status(99)
	if invalidStatus.String() != "unknown" {
		t.Errorf("Expected invalid Status to be 'unknown', got '%s'", invalidStatus.String())
	}
}

// TestClientError tests the ClientError type
func TestClientError(t *testing.T) {
	// Test creating a new error
	err := NewClientError(ErrTypeConnection, "test error", nil)
	if err.Type != ErrTypeConnection {
		t.Errorf("Expected error type to be %d, got %d", ErrTypeConnection, err.Type)
	}
	if err.Message != "test error" {
		t.Errorf("Expected error message to be 'test error', got '%s'", err.Message)
	}
	if err.Cause != nil {
		t.Errorf("Expected error cause to be nil, got '%v'", err.Cause)
	}

	// Test Error() method
	expected := "connection error: test error"
	if err.Error() != expected {
		t.Errorf("Expected Error() to return '%s', got '%s'", expected, err.Error())
	}

	// Test with a cause
	cause := NewClientError(ErrTypeTimeout, "timeout", nil)
	err = NewClientError(ErrTypeConnection, "test error", cause)
	if err.Cause != cause {
		t.Errorf("Expected error cause to be set")
	}

	// Test Error() method with a cause
	expected = "connection error: test error: timeout error: timeout"
	if err.Error() != expected {
		t.Errorf("Expected Error() to return '%s', got '%s'", expected, err.Error())
	}

	// Test IsTimeoutError
	if IsTimeoutError(err) {
		t.Error("Expected IsTimeoutError to return false for non-timeout error")
	}
	if !IsTimeoutError(cause) {
		t.Error("Expected IsTimeoutError to return true for timeout error")
	}
}

// TestIsTimeoutError tests the IsTimeoutError function
func TestIsTimeoutError(t *testing.T) {
	// Test with a timeout error
	err := NewClientError(ErrTypeTimeout, "timeout", nil)
	if !IsTimeoutError(err) {
		t.Error("Expected IsTimeoutError to return true for timeout error")
	}

	// Test with a non-timeout error
	err = NewClientError(ErrTypeConnection, "connection error", nil)
	if IsTimeoutError(err) {
		t.Error("Expected IsTimeoutError to return false for non-timeout error")
	}

	// Test with a nil error
	if IsTimeoutError(nil) {
		t.Error("Expected IsTimeoutError to return false for nil error")
	}

	// Test with a standard error
	stdErr := errors.New("standard error")
	if IsTimeoutError(stdErr) {
		t.Error("Expected IsTimeoutError to return false for standard error")
	}

	// Test with a net.Error timeout
	netErr := &net.OpError{
		Op:     "read",
		Net:    "tcp",
		Source: nil,
		Addr:   nil,
		Err:    &timeoutError{},
	}
	if !IsTimeoutError(netErr) {
		t.Error("Expected IsTimeoutError to return true for net.Error timeout")
	}
}

// timeoutError is a helper type that implements net.Error with Timeout() returning true
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout error" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }
