// Package listener provides interfaces and implementations for different protocol listeners
package listener

import (
	"context"
	"net"
)

// Config defines the common configuration for all listeners
type Config struct {
	// Address is the listening address in format "host:port"
	Address string
	
	// EnableTLS enables TLS for protocols that support it
	EnableTLS bool
	
	// TLSCertFile is the path to the TLS certificate file
	TLSCertFile string
	
	// TLSKeyFile is the path to the TLS key file
	TLSKeyFile string
	
	// BufferSize is the size of the buffer for reading data
	BufferSize int
	
	// MaxConnections is the maximum number of concurrent connections
	MaxConnections int
	
	// Timeout is the connection timeout in seconds
	Timeout int
}

// Status represents the current status of a listener
type Status string

const (
	// StatusStopped indicates the listener is stopped
	StatusStopped Status = "stopped"
	
	// StatusRunning indicates the listener is running
	StatusRunning Status = "running"
	
	// StatusError indicates the listener encountered an error
	StatusError Status = "error"
)

// ConnectionHandler defines the function signature for handling new connections
type ConnectionHandler func(conn net.Conn)

// Listener defines the interface that all protocol listeners must implement
type Listener interface {
	// Start starts the listener with the given context and connection handler
	Start(ctx context.Context, handler ConnectionHandler) error
	
	// Stop stops the listener
	Stop() error
	
	// GetProtocol returns the protocol name
	GetProtocol() string
	
	// GetStatus returns the current status of the listener
	GetStatus() Status
	
	// GetConfig returns the current configuration of the listener
	GetConfig() Config
	
	// UpdateConfig updates the listener configuration
	UpdateConfig(config Config) error
}
