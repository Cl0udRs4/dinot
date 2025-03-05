package listener

import (
	"context"
	"net"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/common"
)

// TCPListener implements the Listener interface for TCP protocol
type TCPListener struct {
	*BaseListener
	listener net.Listener
}

// NewTCPListener creates a new TCP listener
func NewTCPListener(config Config) *TCPListener {
	return &TCPListener{
		BaseListener: NewBaseListener("tcp", config),
	}
}

// Start starts the TCP listener
func (t *TCPListener) Start(ctx context.Context, handler ConnectionHandler) error {
	if t.GetStatus() == StatusRunning {
		return common.NewServerError(common.ErrListenerAlreadyRunning, "TCP listener is already running", nil)
	}

	if err := t.ValidateConfig(); err != nil {
		return err
	}

	var err error
	t.listener, err = net.Listen("tcp", t.Config.Address)
	if err != nil {
		t.setStatus(StatusError)
		return common.NewServerError(common.ErrInvalidConfig, "failed to start TCP listener", err)
	}

	ctx, t.cancel = context.WithCancel(ctx)
	t.setStatus(StatusRunning)

	// Start accepting connections in a separate goroutine
	go t.acceptConnections(ctx, handler)

	return nil
}

// Stop stops the TCP listener
func (t *TCPListener) Stop() error {
	if t.GetStatus() != StatusRunning {
		return common.NewServerError(common.ErrListenerNotRunning, "TCP listener is not running", nil)
	}

	if t.listener != nil {
		err := t.listener.Close()
		if err != nil {
			return common.NewServerError(common.ErrListenerNotRunning, "failed to stop TCP listener", err)
		}
	}

	// Call the base Stop method to cancel the context and update status
	return t.BaseListener.Stop()
}

// acceptConnections accepts incoming TCP connections
func (t *TCPListener) acceptConnections(ctx context.Context, handler ConnectionHandler) {
	// Create a semaphore to limit the number of concurrent connections
	semaphore := make(chan struct{}, t.Config.MaxConnections)

	// Create a goroutine to handle the context cancellation
	go func() {
		<-ctx.Done()
		if t.listener != nil {
			t.listener.Close()
		}
	}()

	for {
		// Check if the context is done
		select {
		case <-ctx.Done():
			t.setStatus(StatusStopped)
			return
		default:
			// Continue accepting connections
		}

		// Set accept deadline to allow for context cancellation checks
		if t.Config.Timeout > 0 {
			t.listener.(*net.TCPListener).SetDeadline(time.Now().Add(time.Duration(t.Config.Timeout) * time.Second))
		}

		// Accept a new connection
		conn, err := t.listener.Accept()
		if err != nil {
			// Check if the error is due to the listener being closed
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

		// Acquire a semaphore slot
		semaphore <- struct{}{}

		// Handle the connection in a separate goroutine
		go func(conn net.Conn) {
			defer func() {
				conn.Close()
				// Release the semaphore slot
				<-semaphore
			}()

			// Set connection timeout
			if t.Config.Timeout > 0 {
				conn.SetDeadline(time.Now().Add(time.Duration(t.Config.Timeout) * time.Second))
			}

			// Call the connection handler
			handler(conn)
		}(conn)
	}
}
