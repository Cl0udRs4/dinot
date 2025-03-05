package protocol

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// WSProtocol implements the Protocol interface for WebSocket
type WSProtocol struct {
	*BaseProtocol
	dialer    *websocket.Dialer
	wsConn    *websocket.Conn
	urlString string
}

// NewWSProtocol creates a new WebSocket protocol instance
func NewWSProtocol(config Config) *WSProtocol {
	return &WSProtocol{
		BaseProtocol: NewBaseProtocol("ws", config),
	}
}

// Connect establishes a WebSocket connection to the server
func (w *WSProtocol) Connect(ctx context.Context) error {
	if w.IsConnected() {
		return ErrAlreadyConnected
	}

	if err := w.ValidateConfig(); err != nil {
		return err
	}

	// Create a context with cancel function
	ctx, w.cancel = context.WithCancel(ctx)

	// Parse the server address into a URL
	scheme := "ws"
	if w.Config.EnableTLS {
		scheme = "wss"
	}
	
	// If the address doesn't have a scheme, add it
	if w.urlString == "" {
		w.urlString = scheme + "://" + w.Config.ServerAddress
	}

	u, err := url.Parse(w.urlString)
	if err != nil {
		w.setLastError(err)
		return NewClientError(ErrTypeConnection, "invalid WebSocket URL", err)
	}

	// Create a dialer with timeout
	w.dialer = &websocket.Dialer{
		HandshakeTimeout: time.Duration(w.Config.ConnectTimeout) * time.Second,
	}

	// Configure TLS if enabled
	if w.Config.EnableTLS {
		w.dialer.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: w.Config.TLSSkipVerify,
		}
	}

	// Try to connect with retry logic
	var conn *websocket.Conn
	var retryCount int
	var header http.Header

	for retryCount <= w.Config.RetryCount {
		// Check if the context is done
		select {
		case <-ctx.Done():
			w.setStatus(StatusDisconnected)
			w.setLastError(ctx.Err())
			return NewClientError(ErrTypeConnection, "connection cancelled", ctx.Err())
		default:
			// Continue connecting
		}

		// Try to connect
		conn, _, err = w.dialer.DialContext(ctx, u.String(), header)
		if err == nil {
			break // Connection successful
		}

		// If this was the last retry, return the error
		if retryCount == w.Config.RetryCount {
			w.setStatus(StatusDisconnected)
			w.setLastError(err)
			return NewClientError(ErrTypeConnection, "failed to connect after retries", err)
		}

		// Wait before retrying
		retryCount++
		select {
		case <-ctx.Done():
			w.setStatus(StatusDisconnected)
			w.setLastError(ctx.Err())
			return NewClientError(ErrTypeConnection, "connection cancelled during retry", ctx.Err())
		case <-time.After(time.Duration(w.Config.RetryInterval) * time.Second):
			// Continue to next retry
		}
	}

	// Store the connection
	w.wsConn = conn
	w.conn = nil // WebSocket doesn't use the standard net.Conn
	w.setStatus(StatusConnected)
	w.setLastError(nil)

	return nil
}

// Disconnect closes the WebSocket connection
func (w *WSProtocol) Disconnect() error {
	if !w.IsConnected() {
		return nil // Already disconnected
	}

	// Close the WebSocket connection
	if w.wsConn != nil {
		// Send close message
		err := w.wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			// Just log the error, still proceed with closing
			w.setLastError(err)
		}

		// Close the connection
		err = w.wsConn.Close()
		if err != nil {
			w.setLastError(err)
			return NewClientError(ErrTypeDisconnection, "failed to close WebSocket connection", err)
		}
		w.wsConn = nil
	}

	// Call the base Disconnect method to cancel the context and update status
	w.BaseProtocol.Disconnect()
	return nil
}

// Send sends data over the WebSocket connection
func (w *WSProtocol) Send(data []byte) (int, error) {
	if !w.IsConnected() {
		return 0, ErrNotConnected
	}

	// Set write deadline if configured
	if w.Config.WriteTimeout > 0 {
		w.wsConn.SetWriteDeadline(time.Now().Add(time.Duration(w.Config.WriteTimeout) * time.Second))
		defer w.wsConn.SetWriteDeadline(time.Time{}) // Clear the deadline
	}

	// Send the data as a binary message
	err := w.wsConn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		w.setLastError(err)
		return 0, NewClientError(ErrTypeSend, "failed to send data", err)
	}

	// WebSocket.WriteMessage doesn't return the number of bytes written,
	// but we can assume it's the length of the data if no error occurred
	return len(data), nil
}

// Receive receives data from the WebSocket connection
func (w *WSProtocol) Receive() ([]byte, error) {
	if !w.IsConnected() {
		return nil, ErrNotConnected
	}

	// Set read deadline if configured
	if w.Config.ReadTimeout > 0 {
		w.wsConn.SetReadDeadline(time.Now().Add(time.Duration(w.Config.ReadTimeout) * time.Second))
		defer w.wsConn.SetReadDeadline(time.Time{}) // Clear the deadline
	}

	// Read a message from the WebSocket
	messageType, data, err := w.wsConn.ReadMessage()
	if err != nil {
		w.setLastError(err)
		
		// Check if it's a close message
		if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			return nil, NewClientError(ErrTypeDisconnection, "connection closed", err)
		}
		
		// Check if it's a timeout error
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return nil, NewClientError(ErrTypeTimeout, "read timeout", err)
		}
		
		return nil, NewClientError(ErrTypeReceive, "failed to receive data", err)
	}

	// Handle different message types
	switch messageType {
	case websocket.TextMessage, websocket.BinaryMessage:
		return data, nil
	case websocket.CloseMessage:
		w.setLastError(ErrNotConnected)
		return nil, NewClientError(ErrTypeDisconnection, "received close message", nil)
	default:
		w.setLastError(ErrNotImplemented)
		return nil, NewClientError(ErrTypeReceive, "unsupported message type", nil)
	}
}

// GetConnection returns the underlying connection
// Note: This overrides the base method since WebSocket doesn't use net.Conn
func (w *WSProtocol) GetConnection() net.Conn {
	return nil
}

// GetWebSocketConnection returns the underlying WebSocket connection
func (w *WSProtocol) GetWebSocketConnection() *websocket.Conn {
	return w.wsConn
}

// ValidateConfig validates the WebSocket protocol configuration
func (w *WSProtocol) ValidateConfig() error {
	// Call the base ValidateConfig method
	if err := w.BaseProtocol.ValidateConfig(); err != nil {
		return err
	}

	// Additional WebSocket-specific validation
	if w.Config.EnableTLS && w.Config.TLSCertFile != "" {
		// Validate that the certificate file exists
		if _, err := os.Stat(w.Config.TLSCertFile); err != nil {
			return NewClientError(ErrTypeConfiguration, "TLS certificate file not found", err)
		}
	}

	return nil
}

// IsConnected returns true if the WebSocket is connected
// Note: This overrides the base method to check the WebSocket connection
func (w *WSProtocol) IsConnected() bool {
	return w.wsConn != nil && w.GetStatus() == StatusConnected
}
