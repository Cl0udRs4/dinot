// Package protocol provides interfaces and implementations for different communication protocols
package protocol

import (
	"context"
	"errors"
	"net"
	"time"
)

// Common errors
var (
	ErrNotConnected      = errors.New("not connected")
	ErrAlreadyConnected  = errors.New("already connected")
	ErrConnectionTimeout = errors.New("connection timeout")
	ErrInvalidConfig     = errors.New("invalid configuration")
	ErrNotImplemented    = errors.New("not implemented")
	ErrProtocolSwitch    = errors.New("protocol switch required")
)

// Status represents the current status of a protocol connection
type Status string

const (
	// StatusDisconnected indicates the protocol is not connected
	StatusDisconnected Status = "disconnected"
	
	// StatusConnected indicates the protocol is connected
	StatusConnected Status = "connected"
	
	// StatusError indicates the protocol encountered an error
	StatusError Status = "error"
	
	// StatusSwitching indicates the protocol is switching to another protocol
	StatusSwitching Status = "switching"
)

// Config defines the common configuration for all protocols
type Config struct {
	// ServerAddress is the address of the server in format "host:port"
	ServerAddress string
	
	// EnableTLS enables TLS for protocols that support it
	EnableTLS bool
	
	// TLSCertFile is the path to the TLS certificate file
	TLSCertFile string
	
	// TLSSkipVerify skips TLS certificate verification
	TLSSkipVerify bool
	
	// BufferSize is the size of the buffer for reading data
	BufferSize int
	
	// ConnectTimeout is the connection timeout in seconds
	ConnectTimeout int
	
	// ReadTimeout is the read timeout in seconds
	ReadTimeout int
	
	// WriteTimeout is the write timeout in seconds
	WriteTimeout int
	
	// RetryCount is the number of connection retry attempts
	RetryCount int
	
	// RetryInterval is the interval between retry attempts in seconds
	RetryInterval int
	
	// KeepAlive enables TCP keepalive for protocols that support it
	KeepAlive bool
	
	// KeepAliveInterval is the keepalive interval in seconds
	KeepAliveInterval int
}

// Protocol defines the interface that all protocol implementations must implement
type Protocol interface {
	// Connect establishes a connection to the server
	Connect(ctx context.Context) error
	
	// Disconnect closes the connection to the server
	Disconnect() error
	
	// Send sends data to the server
	Send(data []byte) (int, error)
	
	// Receive receives data from the server
	Receive() ([]byte, error)
	
	// GetName returns the protocol name
	GetName() string
	
	// GetStatus returns the current status of the protocol
	GetStatus() Status
	
	// GetConfig returns the current configuration of the protocol
	GetConfig() Config
	
	// UpdateConfig updates the protocol configuration
	UpdateConfig(config Config) error
	
	// IsConnected returns true if the protocol is connected
	IsConnected() bool
	
	// GetLastError returns the last error encountered by the protocol
	GetLastError() error
	
	// GetConnection returns the underlying net.Conn if available
	GetConnection() net.Conn
}

// BaseProtocol provides common functionality for all protocols
type BaseProtocol struct {
	// Name is the name of the protocol
	Name string
	
	// Config is the protocol configuration
	Config Config
	
	// status is the current status of the protocol
	status Status
	
	// lastError is the last error encountered by the protocol
	lastError error
	
	// conn is the underlying connection if available
	conn net.Conn
	
	// cancel is the function to cancel the protocol context
	cancel context.CancelFunc
}

// NewBaseProtocol creates a new base protocol
func NewBaseProtocol(name string, config Config) *BaseProtocol {
	return &BaseProtocol{
		Name:   name,
		Config: config,
		status: StatusDisconnected,
	}
}

// GetName returns the protocol name
func (b *BaseProtocol) GetName() string {
	return b.Name
}

// GetStatus returns the current status of the protocol
func (b *BaseProtocol) GetStatus() Status {
	return b.status
}

// setStatus sets the status of the protocol
func (b *BaseProtocol) setStatus(status Status) {
	b.status = status
}

// GetConfig returns the current configuration of the protocol
func (b *BaseProtocol) GetConfig() Config {
	return b.Config
}

// UpdateConfig updates the protocol configuration
func (b *BaseProtocol) UpdateConfig(config Config) error {
	if b.IsConnected() {
		return ErrAlreadyConnected
	}
	b.Config = config
	return nil
}

// IsConnected returns true if the protocol is connected
func (b *BaseProtocol) IsConnected() bool {
	return b.status == StatusConnected
}

// GetLastError returns the last error encountered by the protocol
func (b *BaseProtocol) GetLastError() error {
	return b.lastError
}

// setLastError sets the last error encountered by the protocol
func (b *BaseProtocol) setLastError(err error) {
	b.lastError = err
	if err != nil {
		b.setStatus(StatusError)
	}
}

// GetConnection returns the underlying net.Conn if available
func (b *BaseProtocol) GetConnection() net.Conn {
	return b.conn
}

// Connect is a placeholder that should be overridden by specific protocols
func (b *BaseProtocol) Connect(ctx context.Context) error {
	return ErrNotImplemented
}

// Disconnect is a placeholder that should be overridden by specific protocols
func (b *BaseProtocol) Disconnect() error {
	if b.cancel != nil {
		b.cancel()
		b.cancel = nil
	}
	
	if b.conn != nil {
		err := b.conn.Close()
		b.conn = nil
		b.setStatus(StatusDisconnected)
		return err
	}
	
	b.setStatus(StatusDisconnected)
	return nil
}

// Send is a placeholder that should be overridden by specific protocols
func (b *BaseProtocol) Send(data []byte) (int, error) {
	return 0, ErrNotImplemented
}

// Receive is a placeholder that should be overridden by specific protocols
func (b *BaseProtocol) Receive() ([]byte, error) {
	return nil, ErrNotImplemented
}

// ValidateConfig validates the protocol configuration
func (b *BaseProtocol) ValidateConfig() error {
	if b.Config.ServerAddress == "" {
		return ErrInvalidConfig
	}
	
	if b.Config.BufferSize <= 0 {
		b.Config.BufferSize = 4096 // Default buffer size
	}
	
	if b.Config.ConnectTimeout <= 0 {
		b.Config.ConnectTimeout = 30 // Default connect timeout in seconds
	}
	
	if b.Config.ReadTimeout <= 0 {
		b.Config.ReadTimeout = 30 // Default read timeout in seconds
	}
	
	if b.Config.WriteTimeout <= 0 {
		b.Config.WriteTimeout = 30 // Default write timeout in seconds
	}
	
	if b.Config.RetryCount < 0 {
		b.Config.RetryCount = 3 // Default retry count
	}
	
	if b.Config.RetryInterval <= 0 {
		b.Config.RetryInterval = 5 // Default retry interval in seconds
	}
	
	if b.Config.KeepAlive && b.Config.KeepAliveInterval <= 0 {
		b.Config.KeepAliveInterval = 60 // Default keepalive interval in seconds
	}
	
	return nil
}

// SetReadTimeout sets the read timeout on the connection if available
func (b *BaseProtocol) SetReadTimeout(timeout time.Duration) error {
	if b.conn == nil {
		return ErrNotConnected
	}
	
	return b.conn.SetReadDeadline(time.Now().Add(timeout))
}

// SetWriteTimeout sets the write timeout on the connection if available
func (b *BaseProtocol) SetWriteTimeout(timeout time.Duration) error {
	if b.conn == nil {
		return ErrNotConnected
	}
	
	return b.conn.SetWriteDeadline(time.Now().Add(timeout))
}

// ClearTimeout clears any timeout on the connection if available
func (b *BaseProtocol) ClearTimeout() error {
	if b.conn == nil {
		return ErrNotConnected
	}
	
	return b.conn.SetDeadline(time.Time{})
}
