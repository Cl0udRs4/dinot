package listener

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/common"
)

// WSListener implements the Listener interface for WebSocket protocol
type WSListener struct {
	*BaseListener
	server   *http.Server
	connsMtx sync.RWMutex
	conns    map[string]bool
}

// NewWSListener creates a new WebSocket listener
func NewWSListener(config Config) *WSListener {
	return &WSListener{
		BaseListener: NewBaseListener("ws", config),
		conns: make(map[string]bool),
	}
}

// Start starts the WebSocket listener
func (w *WSListener) Start(ctx context.Context, handler ConnectionHandler) error {
	if w.GetStatus() == StatusRunning {
		return common.NewServerError(common.ErrListenerAlreadyRunning, "WebSocket listener is already running", nil)
	}

	if err := w.ValidateConfig(); err != nil {
		return err
	}

	// Create a new HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		w.handleWebSocket(writer, request, handler)
	})

	w.server = &http.Server{
		Addr:    w.Config.Address,
		Handler: mux,
	}

	ctx, w.cancel = context.WithCancel(ctx)
	w.setStatus(StatusRunning)

	// Start the HTTP server in a separate goroutine
	go func() {
		var err error
		if w.Config.EnableTLS && w.Config.TLSCertFile != "" && w.Config.TLSKeyFile != "" {
			err = w.server.ListenAndServeTLS(w.Config.TLSCertFile, w.Config.TLSKeyFile)
		} else {
			err = w.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			w.setStatus(StatusError)
		}
	}()

	// Create a goroutine to handle the context cancellation
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		w.server.Shutdown(shutdownCtx)
	}()

	return nil
}

// Stop stops the WebSocket listener
func (w *WSListener) Stop() error {
	if w.GetStatus() != StatusRunning {
		return common.NewServerError(common.ErrListenerNotRunning, "WebSocket listener is not running", nil)
	}

	// Call the base Stop method to cancel the context and update status
	return w.BaseListener.Stop()
}

// handleWebSocket handles incoming WebSocket connections
func (w *WSListener) handleWebSocket(writer http.ResponseWriter, request *http.Request, handler ConnectionHandler) {
	// This is a placeholder implementation
	// In a real implementation, we would use the gorilla/websocket package to upgrade the connection
	// For now, we'll just create a simple HTTP response
	
	// Check if we've reached the maximum number of connections
	w.connsMtx.RLock()
	if len(w.conns) >= w.Config.MaxConnections {
		w.connsMtx.RUnlock()
		http.Error(writer, "Too many connections", http.StatusServiceUnavailable)
		return
	}
	w.connsMtx.RUnlock()

	// For now, just write a response
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("WebSocket endpoint"))
}
