package protocol

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestProtocolManager_New tests the creation of a new protocol manager
func TestProtocolManager_New(t *testing.T) {
	config := ManagerConfig{
		PrimaryProtocol:   "tcp",
		FallbackOrder:     []string{"tcp", "udp", "ws"},
		SwitchThreshold:   3,
		MinSwitchInterval: 60,
	}

	manager := NewProtocolManager(config)

	if manager.primaryProtocol != config.PrimaryProtocol {
		t.Errorf("Expected primary protocol to be '%s', got '%s'", config.PrimaryProtocol, manager.primaryProtocol)
	}

	if len(manager.fallbackOrder) != len(config.FallbackOrder) {
		t.Errorf("Expected fallback order length to be %d, got %d", len(config.FallbackOrder), len(manager.fallbackOrder))
	}

	for i, protocol := range config.FallbackOrder {
		if manager.fallbackOrder[i] != protocol {
			t.Errorf("Expected fallback order at index %d to be '%s', got '%s'", i, protocol, manager.fallbackOrder[i])
		}
	}

	if manager.switchThreshold != config.SwitchThreshold {
		t.Errorf("Expected switch threshold to be %d, got %d", config.SwitchThreshold, manager.switchThreshold)
	}

	if manager.minSwitchInterval != time.Duration(config.MinSwitchInterval)*time.Second {
		t.Errorf("Expected min switch interval to be %s, got %s", time.Duration(config.MinSwitchInterval)*time.Second, manager.minSwitchInterval)
	}

	if manager.GetActiveProtocol() != nil {
		t.Error("Expected active protocol to be nil initially")
	}
}

// TestProtocolManager_RegisterProtocol tests registering protocols with the manager
func TestProtocolManager_RegisterProtocol(t *testing.T) {
	manager := NewProtocolManager(ManagerConfig{
		PrimaryProtocol: "tcp",
		FallbackOrder:   []string{"tcp", "udp"},
	})

	// Register a protocol
	protocol1 := NewMockProtocol("tcp")
	err := manager.RegisterProtocol(protocol1)
	if err != nil {
		t.Errorf("Failed to register protocol: %v", err)
	}

	// Check that the protocol was registered
	p, err := manager.GetProtocol("tcp")
	if err != nil {
		t.Errorf("Failed to get protocol: %v", err)
	}
	if p != protocol1 {
		t.Error("Expected to get the registered protocol")
	}

	// Check that the active protocol was set to the primary protocol
	active := manager.GetActiveProtocol()
	if active != protocol1 {
		t.Error("Expected active protocol to be set to the primary protocol")
	}

	// Register another protocol
	protocol2 := NewMockProtocol("udp")
	err = manager.RegisterProtocol(protocol2)
	if err != nil {
		t.Errorf("Failed to register protocol: %v", err)
	}

	// Check that both protocols are registered
	p, err = manager.GetProtocol("udp")
	if err != nil {
		t.Errorf("Failed to get protocol: %v", err)
	}
	if p != protocol2 {
		t.Error("Expected to get the registered protocol")
	}

	// Try to register a protocol with the same name
	protocol3 := NewMockProtocol("tcp")
	err = manager.RegisterProtocol(protocol3)
	if err == nil {
		t.Error("Expected error when registering a protocol with the same name")
	}
}

// TestProtocolManager_SetActiveProtocol tests setting the active protocol
func TestProtocolManager_SetActiveProtocol(t *testing.T) {
	manager := NewProtocolManager(ManagerConfig{
		PrimaryProtocol: "tcp",
		FallbackOrder:   []string{"tcp", "udp"},
	})

	// Register protocols
	protocol1 := NewMockProtocol("tcp")
	protocol2 := NewMockProtocol("udp")
	manager.RegisterProtocol(protocol1)
	manager.RegisterProtocol(protocol2)

	// Check that the primary protocol is active
	active := manager.GetActiveProtocol()
	if active != protocol1 {
		t.Error("Expected primary protocol to be active")
	}

	// Set the second protocol as active
	err := manager.SetActiveProtocol("udp")
	if err != nil {
		t.Errorf("Failed to set active protocol: %v", err)
	}

	// Check that the second protocol is now active
	active = manager.GetActiveProtocol()
	if active != protocol2 {
		t.Error("Expected second protocol to be active")
	}

	// Try to set a non-existent protocol as active
	err = manager.SetActiveProtocol("nonexistent")
	if err == nil {
		t.Error("Expected error when setting a non-existent protocol as active")
	}
}

// TestProtocolManager_Connect tests connecting with the active protocol
func TestProtocolManager_Connect(t *testing.T) {
	manager := NewProtocolManager(ManagerConfig{
		PrimaryProtocol: "tcp",
	})

	// Register a protocol
	protocol := NewMockProtocol("tcp")
	manager.RegisterProtocol(protocol)

	// Connect
	ctx := context.Background()
	err := manager.Connect(ctx)
	if err != nil {
		t.Errorf("Failed to connect: %v", err)
	}

	// Check that the protocol is connected
	if !protocol.IsConnected() {
		t.Error("Expected protocol to be connected")
	}

	// Try to connect when already connected
	err = manager.Connect(ctx)
	if err != nil {
		t.Errorf("Failed to connect when already connected: %v", err)
	}

	// Disconnect
	err = manager.Disconnect()
	if err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}

	// Check that the protocol is disconnected
	if protocol.IsConnected() {
		t.Error("Expected protocol to be disconnected")
	}
}

// TestProtocolManager_Disconnect tests disconnecting the active protocol
func TestProtocolManager_Disconnect(t *testing.T) {
	manager := NewProtocolManager(ManagerConfig{})

	// Try to disconnect when no active protocol
	err := manager.Disconnect()
	if err != nil {
		t.Errorf("Failed to disconnect when no active protocol: %v", err)
	}

	// Register a protocol
	protocol := NewMockProtocol("tcp")
	manager.RegisterProtocol(protocol)

	// Connect
	ctx := context.Background()
	err = manager.Connect(ctx)
	if err != nil {
		t.Errorf("Failed to connect: %v", err)
	}

	// Disconnect
	err = manager.Disconnect()
	if err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}

	// Check that the protocol is disconnected
	if protocol.IsConnected() {
		t.Error("Expected protocol to be disconnected")
	}

	// Try to disconnect when already disconnected
	err = manager.Disconnect()
	if err != nil {
		t.Errorf("Failed to disconnect when already disconnected: %v", err)
	}
}

// TestProtocolManager_SendReceive tests sending and receiving data with the active protocol
func TestProtocolManager_SendReceive(t *testing.T) {
	manager := NewProtocolManager(ManagerConfig{})

	// Try to send when no active protocol
	_, err := manager.Send([]byte("test"))
	if err == nil {
		t.Error("Expected error when sending with no active protocol")
	}

	// Try to receive when no active protocol
	_, err = manager.Receive()
	if err == nil {
		t.Error("Expected error when receiving with no active protocol")
	}

	// Register a protocol
	protocol := NewMockProtocol("tcp")
	manager.RegisterProtocol(protocol)

	// Connect
	ctx := context.Background()
	err = manager.Connect(ctx)
	if err != nil {
		t.Errorf("Failed to connect: %v", err)
	}

	// Send data
	testData := []byte("test")
	n, err := manager.Send(testData)
	if err != nil {
		t.Errorf("Failed to send data: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected to send %d bytes, sent %d", len(testData), n)
	}

	// Receive data
	receivedData, err := manager.Receive()
	if err != nil {
		t.Errorf("Failed to receive data: %v", err)
	}
	if string(receivedData) != "test" {
		t.Errorf("Expected to receive 'test', got '%s'", string(receivedData))
	}

	// Disconnect
	err = manager.Disconnect()
	if err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}

	// Try to send when disconnected
	_, err = manager.Send(testData)
	if err == nil {
		t.Error("Expected error when sending while disconnected")
	}

	// Try to receive when disconnected
	_, err = manager.Receive()
	if err == nil {
		t.Error("Expected error when receiving while disconnected")
	}
}

// TestProtocolManager_SwitchProtocol tests switching protocols
func TestProtocolManager_SwitchProtocol(t *testing.T) {
	manager := NewProtocolManager(ManagerConfig{
		PrimaryProtocol:   "tcp",
		FallbackOrder:     []string{"tcp", "udp", "ws"},
		SwitchThreshold:   2,
		MinSwitchInterval: 0, // No delay for testing
	})

	// Register protocols
	protocol1 := NewMockProtocol("tcp")
	protocol2 := NewMockProtocol("udp")
	protocol3 := NewMockProtocol("ws")
	manager.RegisterProtocol(protocol1)
	manager.RegisterProtocol(protocol2)
	manager.RegisterProtocol(protocol3)

	// Connect
	ctx := context.Background()
	err := manager.Connect(ctx)
	if err != nil {
		t.Errorf("Failed to connect: %v", err)
	}

	// Check that the primary protocol is active
	active := manager.GetActiveProtocol()
	if active != protocol1 {
		t.Error("Expected primary protocol to be active")
	}

	// Record failures to trigger a switch
	manager.RecordFailure()
	manager.RecordFailure()

	// Check that the active protocol was switched
	active = manager.GetActiveProtocol()
	if active != protocol2 {
		t.Errorf("Expected active protocol to be switched to %s, got %s", protocol2.GetName(), active.GetName())
	}

	// Record more failures to trigger another switch
	manager.RecordFailure()
	manager.RecordFailure()

	// Check that the active protocol was switched again
	active = manager.GetActiveProtocol()
	if active != protocol3 {
		t.Errorf("Expected active protocol to be switched to %s, got %s", protocol3.GetName(), active.GetName())
	}

	// Record more failures to trigger a wrap-around switch
	manager.RecordFailure()
	manager.RecordFailure()

	// Check that the active protocol was switched back to the first one
	active = manager.GetActiveProtocol()
	if active != protocol1 {
		t.Errorf("Expected active protocol to be switched back to %s, got %s", protocol1.GetName(), active.GetName())
	}

	// Disconnect
	err = manager.Disconnect()
	if err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}
}

// TestProtocolManager_RecordSuccess tests recording successful operations
func TestProtocolManager_RecordSuccess(t *testing.T) {
	manager := NewProtocolManager(ManagerConfig{
		SwitchThreshold: 2,
	})

	// Register a protocol
	protocol := NewMockProtocol("tcp")
	manager.RegisterProtocol(protocol)

	// Record a failure
	manager.RecordFailure()

	// Check that the failure count was incremented
	if manager.failureCount != 1 {
		t.Errorf("Expected failure count to be 1, got %d", manager.failureCount)
	}

	// Record a success
	manager.RecordSuccess()

	// Check that the failure count was reset
	if manager.failureCount != 0 {
		t.Errorf("Expected failure count to be reset to 0, got %d", manager.failureCount)
	}
}

// TestProtocolManager_GetProtocol tests retrieving protocols by name
func TestProtocolManager_GetProtocol(t *testing.T) {
	manager := NewProtocolManager(ManagerConfig{})

	// Try to get a non-existent protocol
	_, err := manager.GetProtocol("nonexistent")
	if err == nil {
		t.Error("Expected error when getting a non-existent protocol")
	}

	// Register a protocol
	protocol := NewMockProtocol("tcp")
	manager.RegisterProtocol(protocol)

	// Get the protocol
	p, err := manager.GetProtocol("tcp")
	if err != nil {
		t.Errorf("Failed to get protocol: %v", err)
	}
	if p != protocol {
		t.Error("Expected to get the registered protocol")
	}
}

// TestProtocolManager_GetActiveProtocol tests retrieving the active protocol
func TestProtocolManager_GetActiveProtocol(t *testing.T) {
	manager := NewProtocolManager(ManagerConfig{})

	// Check that there's no active protocol initially
	active := manager.GetActiveProtocol()
	if active != nil {
		t.Error("Expected no active protocol initially")
	}

	// Register a protocol
	protocol := NewMockProtocol("tcp")
	manager.RegisterProtocol(protocol)

	// Check that the protocol is now active
	active = manager.GetActiveProtocol()
	if active != protocol {
		t.Error("Expected registered protocol to be active")
	}
}

// TestProtocolManager_HandleError tests error handling
func TestProtocolManager_HandleError(t *testing.T) {
	manager := NewProtocolManager(ManagerConfig{
		SwitchThreshold: 2,
	})

	// Register a protocol
	protocol := NewMockProtocol("tcp")
	manager.RegisterProtocol(protocol)

	// Handle a non-timeout error
	err := errors.New("test error")
	manager.HandleError(err)

	// Check that the failure count was incremented
	if manager.failureCount != 1 {
		t.Errorf("Expected failure count to be 1, got %d", manager.failureCount)
	}

	// Handle a timeout error
	timeoutErr := NewClientError(ErrTypeTimeout, "timeout", nil)
	manager.HandleError(timeoutErr)

	// Check that the failure count was incremented again
	if manager.failureCount != 2 {
		t.Errorf("Expected failure count to be 2, got %d", manager.failureCount)
	}
}
