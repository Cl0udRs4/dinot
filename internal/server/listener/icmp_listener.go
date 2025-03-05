package listener

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/common"
)

// ICMPListener implements the Listener interface for ICMP protocol
type ICMPListener struct {
	*BaseListener
	conn       net.PacketConn
	clientsMtx sync.RWMutex
	clients    map[string]net.Addr
}

// NewICMPListener creates a new ICMP listener
func NewICMPListener(config Config) *ICMPListener {
	return &ICMPListener{
		BaseListener: NewBaseListener("icmp", config),
		clients:      make(map[string]net.Addr),
	}
}

// Start starts the ICMP listener
func (i *ICMPListener) Start(ctx context.Context, handler ConnectionHandler) error {
	if i.GetStatus() == StatusRunning {
		return common.NewServerError(common.ErrListenerAlreadyRunning, "ICMP listener is already running", nil)
	}

	if err := i.ValidateConfig(); err != nil {
		return err
	}

	// Create ICMP connection
	var err error
	i.conn, err = net.ListenPacket("ip4:icmp", i.Config.Address)
	if err != nil {
		i.setStatus(StatusError)
		return common.NewServerError(common.ErrInvalidConfig, "failed to start ICMP listener", err)
	}

	ctx, i.cancel = context.WithCancel(ctx)
	i.setStatus(StatusRunning)

	// Start handling ICMP packets in a separate goroutine
	go i.handlePackets(ctx, handler)

	return nil
}

// Stop stops the ICMP listener
func (i *ICMPListener) Stop() error {
	if i.GetStatus() != StatusRunning {
		return common.NewServerError(common.ErrListenerNotRunning, "ICMP listener is not running", nil)
	}

	if i.conn != nil {
		err := i.conn.Close()
		if err != nil {
			return common.NewServerError(common.ErrListenerNotRunning, "failed to stop ICMP listener", err)
		}
	}

	// Clear clients map
	i.clientsMtx.Lock()
	i.clients = make(map[string]net.Addr)
	i.clientsMtx.Unlock()

	// Call the base Stop method to cancel the context and update status
	return i.BaseListener.Stop()
}
// handlePackets handles incoming ICMP packets
func (i *ICMPListener) handlePackets(ctx context.Context, handler ConnectionHandler) {
	// Create a buffer for reading ICMP packets
	buffer := make([]byte, i.Config.BufferSize)

	// Create a semaphore to limit the number of concurrent packet handlers
	semaphore := make(chan struct{}, i.Config.MaxConnections)

	// Create a goroutine to handle the context cancellation
	go func() {
		<-ctx.Done()
		if i.conn != nil {
			i.conn.Close()
		}
	}()

	for {
		// Check if the context is done
		select {
		case <-ctx.Done():
			i.setStatus(StatusStopped)
			return
		default:
			// Continue handling packets
		}

		// Set read deadline to allow for context cancellation checks
		if i.Config.Timeout > 0 {
			i.conn.SetReadDeadline(time.Now().Add(time.Duration(i.Config.Timeout) * time.Second))
		}

		// Read an ICMP packet
		n, addr, err := i.conn.ReadFrom(buffer)
		if err != nil {
			// Check if the error is due to timeout
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			// Check if the context is done
			select {
			case <-ctx.Done():
				return
			default:
				// Log the error and continue
				continue
			}
		}

		// Store client address
		clientKey := addr.String()
		i.clientsMtx.Lock()
		i.clients[clientKey] = addr
		i.clientsMtx.Unlock()

		// Acquire a semaphore slot
		semaphore <- struct{}{}

		// Handle the packet in a separate goroutine
		go func(data []byte, addr net.Addr) {
			defer func() {
				// Release the semaphore slot
				<-semaphore
			}()

			// Create an ICMP connection wrapper to make it compatible with the ConnectionHandler
			conn := &ICMPConnWrapper{
				packetConn: i.conn,
				remoteAddr: addr,
				localAddr:  i.conn.LocalAddr(),
				buffer:     data,
			}

			// Call the connection handler
			handler(conn)
		}(buffer[:n], addr)
	}
}

// ICMPConnWrapper wraps an ICMP connection to make it compatible with the net.Conn interface
type ICMPConnWrapper struct {
	packetConn net.PacketConn
	remoteAddr net.Addr
	localAddr  net.Addr
	buffer     []byte
	readPos    int
}

// Read reads data from the connection
func (i *ICMPConnWrapper) Read(b []byte) (n int, err error) {
	// If we have read all the data, return EOF
	if i.readPos >= len(i.buffer) {
		return 0, net.ErrClosed
	}

	// Copy data from the buffer to the provided slice
	n = copy(b, i.buffer[i.readPos:])
	i.readPos += n
	return n, nil
}

// Write writes data to the connection
func (i *ICMPConnWrapper) Write(b []byte) (n int, err error) {
	return i.packetConn.WriteTo(b, i.remoteAddr)
}

// Close closes the connection
func (i *ICMPConnWrapper) Close() error {
	// ICMP is connectionless, so there is nothing to close for a specific client
	return nil
}

// LocalAddr returns the local network address
func (i *ICMPConnWrapper) LocalAddr() net.Addr {
	return i.localAddr
}

// RemoteAddr returns the remote network address
func (i *ICMPConnWrapper) RemoteAddr() net.Addr {
	return i.remoteAddr
}

// SetDeadline sets the read and write deadlines
func (i *ICMPConnWrapper) SetDeadline(t time.Time) error {
	return i.packetConn.SetDeadline(t)
}

// SetReadDeadline sets the read deadline
func (i *ICMPConnWrapper) SetReadDeadline(t time.Time) error {
	return i.packetConn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline
func (i *ICMPConnWrapper) SetWriteDeadline(t time.Time) error {
	return i.packetConn.SetWriteDeadline(t)
}
