package protocol

import (
	"context"
	"net"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/common"
)

// TCPProtocol implements the Protocol interface for TCP
type TCPProtocol struct {
	*BaseProtocol
}

// NewTCPProtocol creates a new TCP protocol instance
func NewTCPProtocol(config Config) *TCPProtocol {
	return &TCPProtocol{
		BaseProtocol: NewBaseProtocol("tcp", config),
	}
}

// Connect establishes a TCP connection to the server
func (t *TCPProtocol) Connect(ctx context.Context) error {
	if t.IsConnected() {
		return ErrAlreadyConnected
	}

	if err := t.ValidateConfig(); err != nil {
		return err
	}

	// Create a dialer with timeout
	dialer := &net.Dialer{
		Timeout: time.Duration(t.Config.ConnectTimeout) * time.Second,
	}

	// Enable TCP keepalive if configured
	if t.Config.KeepAlive {
		dialer.KeepAlive = time.Duration(t.Config.KeepAliveInterval) * time.Second
	}

	// Create a context with cancel function
	ctx, t.cancel = context.WithCancel(ctx)

	// Try to connect with retry logic
	var conn net.Conn
	var err error
	var retryCount int

	for retryCount <= t.Config.RetryCount {
		// Check if the context is done
		select {
		case <-ctx.Done():
			t.setStatus(StatusDisconnected)
			t.setLastError(ctx.Err())
			return NewClientError(ErrTypeConnection, "connection cancelled", ctx.Err())
		default:
			// Continue connecting
		}

		// Try to connect
		conn, err = dialer.DialContext(ctx, "tcp", t.Config.ServerAddress)
		if err == nil {
			break // Connection successful
		}

		// If this was the last retry, return the error
		if retryCount == t.Config.RetryCount {
			t.setStatus(StatusDisconnected)
			t.setLastError(err)
			return NewClientError(ErrTypeConnection, "failed to connect after retries", err)
		}

		// Wait before retrying
		retryCount++
		select {
		case <-ctx.Done():
			t.setStatus(StatusDisconnected)
			t.setLastError(ctx.Err())
			return NewClientError(ErrTypeConnection, "connection cancelled during retry", ctx.Err())
		case <-time.After(time.Duration(t.Config.RetryInterval) * time.Second):
			// Continue to next retry
		}
	}

	// Store the connection
	t.conn = conn
	t.setStatus(StatusConnected)
	t.setLastError(nil)

	return nil
}

// Disconnect closes the TCP connection
func (t *TCPProtocol) Disconnect() error {
	if !t.IsConnected() {
		return nil // Already disconnected
	}

	// Call the base Disconnect method
	err := t.BaseProtocol.Disconnect()
	if err != nil {
		return NewClientError(ErrTypeDisconnection, "failed to disconnect", err)
	}

	return nil
}

// Send sends data over the TCP connection
func (t *TCPProtocol) Send(data []byte) (int, error) {
	if !t.IsConnected() {
		return 0, ErrNotConnected
	}

	// Set write timeout if configured
	if t.Config.WriteTimeout > 0 {
		t.conn.SetWriteDeadline(time.Now().Add(time.Duration(t.Config.WriteTimeout) * time.Second))
		defer t.conn.SetWriteDeadline(time.Time{}) // Clear the deadline
	}

	// Send the data
	n, err := t.conn.Write(data)
	if err != nil {
		// Check if it's a timeout error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			t.setLastError(err)
			return n, NewClientError(ErrTypeTimeout, "write timeout", err)
		}

		// Handle other errors
		t.setLastError(err)
		return n, NewClientError(ErrTypeSend, "failed to send data", err)
	}

	return n, nil
}

// Receive receives data from the TCP connection
func (t *TCPProtocol) Receive() ([]byte, error) {
	if !t.IsConnected() {
		return nil, ErrNotConnected
	}

	// Set read timeout if configured
	if t.Config.ReadTimeout > 0 {
		t.conn.SetReadDeadline(time.Now().Add(time.Duration(t.Config.ReadTimeout) * time.Second))
		defer t.conn.SetReadDeadline(time.Time{}) // Clear the deadline
	}

	// Create a buffer to read data
	buffer := make([]byte, t.Config.BufferSize)

	// Read data from the connection
	n, err := t.conn.Read(buffer)
	if err != nil {
		// Check if it's a timeout error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			t.setLastError(err)
			return nil, NewClientError(ErrTypeTimeout, "read timeout", err)
		}

		// Handle other errors
		t.setLastError(err)
		return nil, NewClientError(ErrTypeReceive, "failed to receive data", err)
	}

	// Return only the data that was read
	return buffer[:n], nil
}

// ValidateConfig validates the TCP protocol configuration
func (t *TCPProtocol) ValidateConfig() error {
	// Call the base ValidateConfig method
	if err := t.BaseProtocol.ValidateConfig(); err != nil {
		return err
	}

	// Additional TCP-specific validation
	if t.Config.KeepAlive && t.Config.KeepAliveInterval <= 0 {
		return common.NewServerError(common.ErrInvalidConfig, "keepalive interval must be greater than 0", nil)
	}

	return nil
}
