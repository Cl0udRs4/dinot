package listener

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/common"
)

// UDPListener implements the Listener interface for UDP protocol
type UDPListener struct {
	*BaseListener
	conn       *net.UDPConn
	clientsMtx sync.RWMutex
	clients    map[string]*net.UDPAddr
}

// NewUDPListener creates a new UDP listener
func NewUDPListener(config Config) *UDPListener {
	return &UDPListener{
		BaseListener: NewBaseListener("udp", config),
		clients:      make(map[string]*net.UDPAddr),
	}
}

// Start starts the UDP listener
func (u *UDPListener) Start(ctx context.Context, handler ConnectionHandler) error {
	if u.GetStatus() == StatusRunning {
		return common.NewServerError(common.ErrListenerAlreadyRunning, "UDP listener is already running", nil)
	}

	if err := u.ValidateConfig(); err != nil {
		return err
	}

	// Resolve UDP address
	addr, err := net.ResolveUDPAddr("udp", u.Config.Address)
	if err != nil {
		u.setStatus(StatusError)
		return common.NewServerError(common.ErrInvalidConfig, "failed to resolve UDP address", err)
	}

	// Create UDP connection
	u.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		u.setStatus(StatusError)
		return common.NewServerError(common.ErrInvalidConfig, "failed to start UDP listener", err)
	}

	ctx, u.cancel = context.WithCancel(ctx)
	u.setStatus(StatusRunning)

	// Start handling UDP packets in a separate goroutine
	go u.handlePackets(ctx, handler)

	return nil
}

// Stop stops the UDP listener
func (u *UDPListener) Stop() error {
	if u.GetStatus() != StatusRunning {
		return common.NewServerError(common.ErrListenerNotRunning, "UDP listener is not running", nil)
	}

	if u.conn != nil {
		err := u.conn.Close()
		if err != nil {
			return common.NewServerError(common.ErrListenerNotRunning, "failed to stop UDP listener", err)
		}
	}

	// Clear clients map
	u.clientsMtx.Lock()
	u.clients = make(map[string]*net.UDPAddr)
	u.clientsMtx.Unlock()

	// Call the base Stop method to cancel the context and update status
	return u.BaseListener.Stop()
}

// handlePackets handles incoming UDP packets
func (u *UDPListener) handlePackets(ctx context.Context, handler ConnectionHandler) {
	// Create a buffer for reading UDP packets
	buffer := make([]byte, u.Config.BufferSize)

	// Create a semaphore to limit the number of concurrent packet handlers
	semaphore := make(chan struct{}, u.Config.MaxConnections)

	// Create a goroutine to handle the context cancellation
	go func() {
		<-ctx.Done()
		if u.conn != nil {
			u.conn.Close()
		}
	}()

	for {
		// Check if the context is done
		select {
		case <-ctx.Done():
			u.setStatus(StatusStopped)
			return
		default:
			// Continue handling packets
		}

		// Set read deadline to allow for context cancellation checks
		if u.Config.Timeout > 0 {
			u.conn.SetReadDeadline(time.Now().Add(time.Duration(u.Config.Timeout) * time.Second))
		}

		// Read a UDP packet
		n, addr, err := u.conn.ReadFromUDP(buffer)
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
		u.clientsMtx.Lock()
		u.clients[clientKey] = addr
		u.clientsMtx.Unlock()

		// Acquire a semaphore slot
		semaphore <- struct{}{}

		// Handle the packet in a separate goroutine
		go func(data []byte, addr *net.UDPAddr) {
			defer func() {
				// Release the semaphore slot
				<-semaphore
			}()

			// Create a UDP connection wrapper to make it compatible with the ConnectionHandler
			conn := &UDPConnWrapper{
				udpConn:  u.conn,
				remoteAddr: addr,
				localAddr:  u.conn.LocalAddr(),
				buffer:     data,
			}

			// Call the connection handler
			handler(conn)
		}(buffer[:n], addr)
	}
}

// UDPConnWrapper wraps a UDP connection to make it compatible with the net.Conn interface
type UDPConnWrapper struct {
	udpConn    *net.UDPConn
	remoteAddr *net.UDPAddr
	localAddr  net.Addr
	buffer     []byte
	readPos    int
}

// Read reads data from the connection
func (u *UDPConnWrapper) Read(b []byte) (n int, err error) {
	// If we've read all the data, return EOF
	if u.readPos >= len(u.buffer) {
		return 0, net.ErrClosed
	}

	// Copy data from the buffer to the provided slice
	n = copy(b, u.buffer[u.readPos:])
	u.readPos += n
	return n, nil
}

// Write writes data to the connection
func (u *UDPConnWrapper) Write(b []byte) (n int, err error) {
	return u.udpConn.WriteToUDP(b, u.remoteAddr)
}

// Close closes the connection
func (u *UDPConnWrapper) Close() error {
	// UDP is connectionless, so there's nothing to close for a specific client
	return nil
}

// LocalAddr returns the local network address
func (u *UDPConnWrapper) LocalAddr() net.Addr {
	return u.localAddr
}

// RemoteAddr returns the remote network address
func (u *UDPConnWrapper) RemoteAddr() net.Addr {
	return u.remoteAddr
}

// SetDeadline sets the read and write deadlines
func (u *UDPConnWrapper) SetDeadline(t time.Time) error {
	return u.udpConn.SetDeadline(t)
}

// SetReadDeadline sets the read deadline
func (u *UDPConnWrapper) SetReadDeadline(t time.Time) error {
	return u.udpConn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline
func (u *UDPConnWrapper) SetWriteDeadline(t time.Time) error {
	return u.udpConn.SetWriteDeadline(t)
}
