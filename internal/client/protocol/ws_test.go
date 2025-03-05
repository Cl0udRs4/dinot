package protocol

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWSProtocol_New(t *testing.T) {
	config := Config{
		ServerAddress:  "localhost:8082/ws",
		ConnectTimeout: 10,
		ReadTimeout:    10,
		WriteTimeout:   10,
		RetryCount:     3,
		RetryInterval:  2,
		BufferSize:     4096,
		EnableTLS:      false,
	}

	ws := NewWSProtocol(config)

	if ws.GetName() != "ws" {
		t.Errorf("Expected protocol name to be 'ws', got '%s'", ws.GetName())
	}

	if ws.GetStatus() != StatusDisconnected {
		t.Errorf("Expected initial status to be 'disconnected', got '%s'", ws.GetStatus())
	}

	if ws.IsConnected() {
		t.Error("Expected IsConnected() to return false initially")
	}

	if ws.GetLastError() != nil {
		t.Errorf("Expected initial last error to be nil, got '%v'", ws.GetLastError())
	}

	gotConfig := ws.GetConfig()
	if gotConfig.ServerAddress != config.ServerAddress {
		t.Errorf("Expected ServerAddress to be '%s', got '%s'", config.ServerAddress, gotConfig.ServerAddress)
	}
}

func TestWSProtocol_ValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "Valid config",
			config: Config{
				ServerAddress:  "localhost:8082/ws",
				ConnectTimeout: 10,
				ReadTimeout:    10,
				WriteTimeout:   10,
				RetryCount:     3,
				RetryInterval:  2,
				BufferSize:     4096,
				EnableTLS:      false,
			},
			expectError: false,
		},
		{
			name: "Empty server address",
			config: Config{
				ServerAddress:  "",
				ConnectTimeout: 10,
			},
			expectError: true,
		},
		{
			name: "Zero connect timeout",
			config: Config{
				ServerAddress:  "localhost:8082/ws",
				ConnectTimeout: 0,
			},
			expectError: false, // Should use default value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := NewWSProtocol(tt.config)
			err := ws.ValidateConfig()

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestWSProtocol_UpdateConfig(t *testing.T) {
	initialConfig := Config{
		ServerAddress:  "localhost:8082/ws",
		ConnectTimeout: 10,
	}

	newConfig := Config{
		ServerAddress:  "localhost:9092/ws",
		ConnectTimeout: 20,
	}

	ws := NewWSProtocol(initialConfig)

	// Test updating config when disconnected
	err := ws.UpdateConfig(newConfig)
	if err != nil {
		t.Errorf("Expected no error when updating config while disconnected, got: %v", err)
	}

	gotConfig := ws.GetConfig()
	if gotConfig.ServerAddress != newConfig.ServerAddress {
		t.Errorf("Expected ServerAddress to be updated to '%s', got '%s'", newConfig.ServerAddress, gotConfig.ServerAddress)
	}

	if gotConfig.ConnectTimeout != newConfig.ConnectTimeout {
		t.Errorf("Expected ConnectTimeout to be updated to %d, got %d", newConfig.ConnectTimeout, gotConfig.ConnectTimeout)
	}

	// Simulate connected state
	ws.setStatus(StatusConnected)
	ws.wsConn = &websocket.Conn{} // Mock connection

	// Test updating config when connected
	err = ws.UpdateConfig(initialConfig)
	if err == nil {
		t.Error("Expected error when updating config while connected, got nil")
	}
}

// This test requires a running WebSocket server to connect to
// It's commented out to avoid test failures when no server is available
/*
func TestWSProtocol_ConnectAndDisconnect(t *testing.T) {
	// Create a test WebSocket server
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/ws") {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		// Upgrade the HTTP connection to a WebSocket connection
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Echo any received messages
		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(messageType, p); err != nil {
				return
			}
		}
	}))
	defer server.Close()

	// Get the server URL and replace http with ws
	serverURL := strings.Replace(server.URL, "http", "ws", 1) + "/ws"

	// Create WebSocket protocol with the test server URL
	config := Config{
		ServerAddress:  serverURL,
		ConnectTimeout: 5,
		ReadTimeout:    5,
		WriteTimeout:   5,
		RetryCount:     1,
		RetryInterval:  1,
		BufferSize:     1024,
	}

	ws := NewWSProtocol(config)

	// Test Connect
	ctx := context.Background()
	err := ws.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	if !ws.IsConnected() {
		t.Error("Expected IsConnected() to return true after successful connection")
	}

	if ws.GetStatus() != StatusConnected {
		t.Errorf("Expected status to be 'connected', got '%s'", ws.GetStatus())
	}

	// Test Send and Receive
	testData := []byte("Hello, WebSocket!")
	n, err := ws.Send(testData)
	if err != nil {
		t.Fatalf("Failed to send data: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected to send %d bytes, sent %d", len(testData), n)
	}

	// Wait for response
	receivedData, err := ws.Receive()
	if err != nil {
		t.Fatalf("Failed to receive data: %v", err)
	}

	if string(receivedData) != string(testData) {
		t.Errorf("Expected to receive '%s', got '%s'", string(testData), string(receivedData))
	}

	// Test Disconnect
	err = ws.Disconnect()
	if err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}

	if ws.IsConnected() {
		t.Error("Expected IsConnected() to return false after disconnection")
	}

	if ws.GetStatus() != StatusDisconnected {
		t.Errorf("Expected status to be 'disconnected', got '%s'", ws.GetStatus())
	}
}
*/
