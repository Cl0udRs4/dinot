package protocol

import (
	"context"
	"net"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/common"
)

// UDPProtocol implements the Protocol interface for UDP
type UDPProtocol struct {
	*BaseProtocol
	remoteAddr *net.UDPAddr
}

// NewUDPProtocol creates a new UDP protocol instance
func NewUDPProtocol(config Config) *UDPProtocol {
	return &UDPProtocol{
		BaseProtocol: NewBaseProtocol("udp", config),
	}
}

// Connect establishes a UDP connection to the server
func (u *UDPProtocol) Connect(ctx context.Context) error {
	if u.IsConnected() {
		return ErrAlreadyConnected
	}

	if err := u.ValidateConfig(); err != nil {
		return err
	}

	// Resolve the remote address
	var err error
	u.remoteAddr, err = net.ResolveUDPAddr("udp", u.Config.ServerAddress)
	if err != nil {
		u.setLastError(err)
		return NewClientError(ErrTypeConnection, "failed to resolve UDP address", err)
	}

	// Create a context with cancel function
	ctx, u.cancel = context.WithCancel(ctx)

	// Try to connect with retry logic
	var conn *net.UDPConn
	var retryCount int

	for retryCount <= u.Config.RetryCount {
		// Check if the context is done
		select {
		case <-ctx.Done():
			u.setStatus(StatusDisconnected)
			u.setLastError(ctx.Err())
			return NewClientError(ErrTypeConnection, "connection cancelled", ctx.Err())
		default:
			// Continue connecting
		}

		// Create a local address to bind to
		localAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
		if err != nil {
			u.setLastError(err)
			return NewClientError(ErrTypeConnection, "failed to resolve local UDP address", err)
		}

		// Create a UDP connection
		conn, err = net.DialUDP("udp", localAddr, u.remoteAddr)
		if err == nil {
			break // Connection successful
		}

		// If this was the last retry, return the error
		if retryCount == u.Config.RetryCount {
			u.setStatus(StatusDisconnected)
			u.setLastError(err)
			return NewClientError(ErrTypeConnection, "failed to connect after retries", err)
		}

		// Wait before retrying
		retryCount++
		select {
		case <-ctx.Done():
			u.setStatus(StatusDisconnected)
			u.setLastError(ctx.Err())
			return NewClientError(ErrTypeConnection, "connection cancelled during retry", ctx.Err())
		case <-time.After(time.Duration(u.Config.RetryInterval) * time.Second):
			// Continue to next retry
		}
	}

	// Store the connection
	u.conn = conn
	u.setStatus(StatusConnected)
	u.setLastError(nil)

	return nil
}

// Disconnect closes the UDP connection
func (u *UDPProtocol) Disconnect() error {
	if !u.IsConnected() {
		return nil // Already disconnected
	}

	// Call the base Disconnect method
	err := u.BaseProtocol.Disconnect()
	if err != nil {
		return NewClientError(ErrTypeDisconnection, "failed to disconnect", err)
	}

	return nil
}

// Send sends data over the UDP connection
func (u *UDPProtocol) Send(data []byte) (int, error) {
	if !u.IsConnected() {
		return 0, ErrNotConnected
	}

	// Set write timeout if configured
	if u.Config.WriteTimeout > 0 {
		u.conn.SetWriteDeadline(time.Now().Add(time.Duration(u.Config.WriteTimeout) * time.Second))
		defer u.conn.SetWriteDeadline(time.Time{}) // Clear the deadline
	}

	// Send the data
	n, err := u.conn.Write(data)
	if err != nil {
		// Check if it's a timeout error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			u.setLastError(err)
			return n, NewClientError(ErrTypeTimeout, "write timeout", err)
		}

		// Handle other errors
		u.setLastError(err)
		return n, NewClientError(ErrTypeSend, "failed to send data", err)
	}

	return n, nil
}

// Receive receives data from the UDP connection
func (u *UDPProtocol) Receive() ([]byte, error) {
	if !u.IsConnected() {
		return nil, ErrNotConnected
	}

	// Set read timeout if configured
	if u.Config.ReadTimeout > 0 {
		u.conn.SetReadDeadline(time.Now().Add(time.Duration(u.Config.ReadTimeout) * time.Second))
		defer u.conn.SetReadDeadline(time.Time{}) // Clear the deadline
	}

	// Create a buffer to read data
	buffer := make([]byte, u.Config.BufferSize)

	// Read data from the connection
	n, _, err := u.conn.(*net.UDPConn).ReadFromUDP(buffer)
	if err != nil {
		// Check if it's a timeout error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			u.setLastError(err)
			return nil, NewClientError(ErrTypeTimeout, "read timeout", err)
		}

		// Handle other errors
		u.setLastError(err)
		return nil, NewClientError(ErrTypeReceive, "failed to receive data", err)
	}

	// Return only the data that was read
	return buffer[:n], nil
}

// ValidateConfig validates the UDP protocol configuration
func (u *UDPProtocol) ValidateConfig() error {
	// Call the base ValidateConfig method
	if err := u.BaseProtocol.ValidateConfig(); err != nil {
		return err
	}

	// Additional UDP-specific validation could be added here
	return nil
}
