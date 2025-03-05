package protocol

import (
	"context"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// ICMPProtocol implements the Protocol interface for ICMP
type ICMPProtocol struct {
	*BaseProtocol
	packetConn *icmp.PacketConn
	remoteAddr net.Addr
	sequence   int
	identifier int
}

// NewICMPProtocol creates a new ICMP protocol instance
func NewICMPProtocol(config Config) *ICMPProtocol {
	return &ICMPProtocol{
		BaseProtocol: NewBaseProtocol("icmp", config),
		sequence:     0,
		identifier:   os.Getpid() & 0xffff, // Use process ID as identifier
	}
}

// Connect establishes an ICMP connection to the server
func (i *ICMPProtocol) Connect(ctx context.Context) error {
	if i.IsConnected() {
		return ErrAlreadyConnected
	}

	if err := i.ValidateConfig(); err != nil {
		return err
	}

	// Create a context with cancel function
	ctx, i.cancel = context.WithCancel(ctx)

	// Resolve the remote address
	var err error
	i.remoteAddr, err = net.ResolveIPAddr("ip4", i.Config.ServerAddress)
	if err != nil {
		i.setLastError(err)
		return NewClientError(ErrTypeConnection, "failed to resolve ICMP address", err)
	}

	// Try to connect with retry logic
	var conn *icmp.PacketConn
	var retryCount int

	for retryCount <= i.Config.RetryCount {
		// Check if the context is done
		select {
		case <-ctx.Done():
			i.setStatus(StatusDisconnected)
			i.setLastError(ctx.Err())
			return NewClientError(ErrTypeConnection, "connection cancelled", ctx.Err())
		default:
			// Continue connecting
		}

		// Create an ICMP packet connection
		conn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		if err == nil {
			break // Connection successful
		}

		// If this was the last retry, return the error
		if retryCount == i.Config.RetryCount {
			i.setStatus(StatusDisconnected)
			i.setLastError(err)
			return NewClientError(ErrTypeConnection, "failed to connect after retries", err)
		}

		// Wait before retrying
		retryCount++
		select {
		case <-ctx.Done():
			i.setStatus(StatusDisconnected)
			i.setLastError(ctx.Err())
			return NewClientError(ErrTypeConnection, "connection cancelled during retry", ctx.Err())
		case <-time.After(time.Duration(i.Config.RetryInterval) * time.Second):
			// Continue to next retry
		}
	}

	// Store the connection
	i.packetConn = conn
	i.conn = nil // ICMP doesn't use the standard net.Conn
	i.setStatus(StatusConnected)
	i.setLastError(nil)

	return nil
}

// Disconnect closes the ICMP connection
func (i *ICMPProtocol) Disconnect() error {
	if !i.IsConnected() {
		return nil // Already disconnected
	}

	// Close the ICMP connection
	if i.packetConn != nil {
		err := i.packetConn.Close()
		if err != nil {
			i.setLastError(err)
			return NewClientError(ErrTypeDisconnection, "failed to close ICMP connection", err)
		}
		i.packetConn = nil
	}

	// Call the base Disconnect method to cancel the context and update status
	i.BaseProtocol.Disconnect()
	return nil
}

// Send sends data over the ICMP connection
func (i *ICMPProtocol) Send(data []byte) (int, error) {
	if !i.IsConnected() {
		return 0, ErrNotConnected
	}

	// Increment sequence number
	i.sequence = (i.sequence + 1) % 65536

	// Create an ICMP echo message
	msg := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   i.identifier,
			Seq:  i.sequence,
			Data: data,
		},
	}

	// Marshal the message
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		i.setLastError(err)
		return 0, NewClientError(ErrTypeSend, "failed to marshal ICMP message", err)
	}

	// Set write deadline if configured
	if i.Config.WriteTimeout > 0 {
		i.packetConn.SetWriteDeadline(time.Now().Add(time.Duration(i.Config.WriteTimeout) * time.Second))
		defer i.packetConn.SetWriteDeadline(time.Time{}) // Clear the deadline
	}

	// Send the message
	n, err := i.packetConn.WriteTo(msgBytes, i.remoteAddr)
	if err != nil {
		// Check if it's a timeout error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			i.setLastError(err)
			return n, NewClientError(ErrTypeTimeout, "write timeout", err)
		}

		// Handle other errors
		i.setLastError(err)
		return n, NewClientError(ErrTypeSend, "failed to send data", err)
	}

	return len(data), nil // Return the length of the original data
}

// Receive receives data from the ICMP connection
func (i *ICMPProtocol) Receive() ([]byte, error) {
	if !i.IsConnected() {
		return nil, ErrNotConnected
	}

	// Set read deadline if configured
	if i.Config.ReadTimeout > 0 {
		i.packetConn.SetReadDeadline(time.Now().Add(time.Duration(i.Config.ReadTimeout) * time.Second))
		defer i.packetConn.SetReadDeadline(time.Time{}) // Clear the deadline
	}

	// Create a buffer to read data
	buffer := make([]byte, i.Config.BufferSize)

	// Read a packet
	n, addr, err := i.packetConn.ReadFrom(buffer)
	if err != nil {
		// Check if it's a timeout error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			i.setLastError(err)
			return nil, NewClientError(ErrTypeTimeout, "read timeout", err)
		}

		// Handle other errors
		i.setLastError(err)
		return nil, NewClientError(ErrTypeReceive, "failed to receive data", err)
	}

	// Parse the ICMP message
	msg, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), buffer[:n])
	if err != nil {
		i.setLastError(err)
		return nil, NewClientError(ErrTypeReceive, "failed to parse ICMP message", err)
	}

	// Check if the message is an echo reply
	if msg.Type != ipv4.ICMPTypeEchoReply {
		i.setLastError(ErrNotImplemented)
		return nil, NewClientError(ErrTypeReceive, "received non-echo reply message", nil)
	}

	// Extract the data from the echo reply
	echo, ok := msg.Body.(*icmp.Echo)
	if !ok {
		i.setLastError(ErrNotImplemented)
		return nil, NewClientError(ErrTypeReceive, "received message with invalid body", nil)
	}

	// Check if the echo reply is for our request
	if echo.ID != i.identifier {
		// This is not our echo reply, try again
		return i.Receive()
	}

	return echo.Data, nil
}

// GetConnection returns the underlying connection
// Note: This overrides the base method since ICMP doesn't use net.Conn
func (i *ICMPProtocol) GetConnection() net.Conn {
	return nil
}

// GetICMPConnection returns the underlying ICMP connection
func (i *ICMPProtocol) GetICMPConnection() *icmp.PacketConn {
	return i.packetConn
}

// ValidateConfig validates the ICMP protocol configuration
func (i *ICMPProtocol) ValidateConfig() error {
	// Call the base ValidateConfig method
	if err := i.BaseProtocol.ValidateConfig(); err != nil {
		return err
	}

	// Additional ICMP-specific validation
	// ICMP requires root privileges on most systems
	if os.Geteuid() != 0 {
		return NewClientError(ErrTypeConfiguration, "ICMP protocol requires root privileges", nil)
	}

	return nil
}

// IsConnected returns true if the ICMP connection is connected
// Note: This overrides the base method to check the ICMP connection
func (i *ICMPProtocol) IsConnected() bool {
	return i.packetConn != nil && i.GetStatus() == StatusConnected
}
