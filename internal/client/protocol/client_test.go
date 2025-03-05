package protocol

import (
	"testing"
	"time"
)

func TestClient_New(t *testing.T) {
	config := ClientConfig{
		Protocols: map[string]Config{
			"tcp": {
				ServerAddress:  "localhost:8080",
				ConnectTimeout: 10,
				ReadTimeout:    10,
				WriteTimeout:   10,
			},
			"udp": {
				ServerAddress:  "localhost:8081",
				ConnectTimeout: 10,
				ReadTimeout:    10,
				WriteTimeout:   10,
			},
		},
		PrimaryProtocol:    "tcp",
		FallbackOrder:      []string{"tcp", "udp"},
		SwitchStrategy:     StrategySequential,
		SwitchThreshold:    3,
		MinSwitchInterval:  60,
		TimeoutThreshold:   3,
		CheckInterval:      5,
		MaxInactivity:      60,
		JitterMin:          1,
		JitterMax:          5,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.IsConnected() {
		t.Error("Expected client to be disconnected initially")
	}

	// Check that the active protocol is the primary protocol
	active := client.GetActiveProtocol()
	if active == nil {
		t.Fatal("Expected active protocol to be set")
	}
	if active.GetName() != "tcp" {
		t.Errorf("Expected active protocol to be tcp, got %s", active.GetName())
	}
}

func TestClient_ConnectDisconnect(t *testing.T) {
	// Create a client with mock protocols
	client := createClientWithMockProtocols(t)

	// Set up connect and disconnect handlers
	connectCalled := false
	client.SetOnConnectHandler(func() {
		connectCalled = true
	})

	disconnectCalled := false
	client.SetOnDisconnectHandler(func(err error) {
		disconnectCalled = true
	})

	// Connect
	err := client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Wait for the connect handler to be called
	time.Sleep(10 * time.Millisecond)

	if !client.IsConnected() {
		t.Error("Expected client to be connected after Connect")
	}

	if !connectCalled {
		t.Error("Expected connect handler to be called")
	}

	// Disconnect
	err = client.Disconnect()
	if err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}

	// Wait for the disconnect handler to be called
	time.Sleep(10 * time.Millisecond)

	if client.IsConnected() {
		t.Error("Expected client to be disconnected after Disconnect")
	}

	if !disconnectCalled {
		t.Error("Expected disconnect handler to be called")
	}
}

func TestClient_SendReceive(t *testing.T) {
	// Create a client with mock protocols
	client := createClientWithMockProtocols(t)

	// Set up error handler
	errorCalled := false
	client.SetOnErrorHandler(func(err error) {
		errorCalled = true
	})

	// Try to send before connecting
	_, err := client.Send([]byte("test"))
	if err == nil {
		t.Error("Expected error when sending before connecting")
	}

	// Try to receive before connecting
	_, err = client.Receive()
	if err == nil {
		t.Error("Expected error when receiving before connecting")
	}

	// Connect
	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Send data
	n, err := client.Send([]byte("test"))
	if err != nil {
		t.Fatalf("Failed to send data: %v", err)
	}
	if n != 4 {
		t.Errorf("Expected to send 4 bytes, sent %d", n)
	}

	// Receive data
	data, err := client.Receive()
	if err != nil {
		t.Fatalf("Failed to receive data: %v", err)
	}
	if string(data) != "test" {
		t.Errorf("Expected to receive 'test', got '%s'", string(data))
	}

	// Disconnect
	err = client.Disconnect()
	if err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}

	// Check that the error handler was not called
	if errorCalled {
		t.Error("Expected error handler not to be called")
	}
}

func TestClient_SwitchProtocol(t *testing.T) {
	// Create a client with mock protocols
	client := createClientWithMockProtocols(t)

	// Set up protocol switch handler
	switchCalled := false
	var oldProto, newProto Protocol
	client.SetOnProtocolSwitchHandler(func(old, new Protocol) {
		switchCalled = true
		oldProto = old
		newProto = new
	})

	// Try to switch before connecting
	err := client.SwitchProtocol()
	if err == nil {
		t.Error("Expected error when switching before connecting")
	}

	// Connect
	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Get the initial active protocol
	initialProto := client.GetActiveProtocol()

	// Switch protocol
	err = client.SwitchProtocol()
	if err != nil {
		t.Fatalf("Failed to switch protocol: %v", err)
	}

	// Wait for the switch handler to be called
	time.Sleep(10 * time.Millisecond)

	// Check that the switch handler was called
	if !switchCalled {
		t.Error("Expected switch handler to be called")
	}

	// Check that the old and new protocols are correct
	if oldProto != initialProto {
		t.Errorf("Expected old protocol to be %s, got %s", initialProto.GetName(), oldProto.GetName())
	}

	// Check that the active protocol was switched
	active := client.GetActiveProtocol()
	if active == initialProto {
		t.Error("Expected active protocol to be different from the initial one")
	}
	if active != newProto {
		t.Errorf("Expected active protocol to be %s, got %s", newProto.GetName(), active.GetName())
	}

	// Disconnect
	err = client.Disconnect()
	if err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}
}

func TestClient_Close(t *testing.T) {
	// Create a client with mock protocols
	client := createClientWithMockProtocols(t)

	// Connect
	err := client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Close
	err = client.Close()
	if err != nil {
		t.Fatalf("Failed to close: %v", err)
	}

	// Check that the client is disconnected
	if client.IsConnected() {
		t.Error("Expected client to be disconnected after Close")
	}

	// Try to connect again after closing
	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect after closing: %v", err)
	}

	// Disconnect
	err = client.Disconnect()
	if err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}
}

// Helper function to create a client with mock protocols
func createClientWithMockProtocols(t *testing.T) *Client {
	config := ClientConfig{
		Protocols: map[string]Config{
			"mock1": {},
			"mock2": {},
		},
		PrimaryProtocol:    "mock1",
		FallbackOrder:      []string{"mock1", "mock2"},
		SwitchStrategy:     StrategySequential,
		SwitchThreshold:    3,
		MinSwitchInterval:  0, // No delay for testing
		TimeoutThreshold:   3,
		CheckInterval:      1,
		MaxInactivity:      5,
		JitterMin:          0, // No jitter for testing
		JitterMax:          0,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Replace the switcher with one that uses mock protocols
	switcher := NewProtocolSwitcher(ProtocolSwitcherConfig{
		SwitchStrategy: StrategySequential,
		JitterMin:      0,
		JitterMax:      0,
		TimeoutDetectorConfig: TimeoutDetectorConfig{
			TimeoutThreshold: 3,
			CheckInterval:    1,
			MaxInactivity:    5,
		},
		ManagerConfig: ManagerConfig{
			PrimaryProtocol:    "mock1",
			FallbackOrder:      []string{"mock1", "mock2"},
			SwitchThreshold:    3,
			MinSwitchInterval:  0,
		},
	})

	// Register mock protocols
	protocol1 := NewMockProtocol("mock1")
	protocol2 := NewMockProtocol("mock2")
	switcher.RegisterProtocol(protocol1)
	switcher.RegisterProtocol(protocol2)

	// Set the switcher
	client.switcher = switcher

	// Set up the protocol switch handler
	switcher.SetOnSwitchHandler(client.handleProtocolSwitch)

	return client
}
