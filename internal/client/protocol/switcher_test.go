package protocol

import (
	"context"
	"testing"
	"time"
)

func TestProtocolSwitcher_New(t *testing.T) {
	config := ProtocolSwitcherConfig{
		SwitchStrategy: StrategySequential,
		JitterMin:      1,
		JitterMax:      5,
		TimeoutDetectorConfig: TimeoutDetectorConfig{
			TimeoutThreshold: 3,
			CheckInterval:    5,
			MaxInactivity:    60,
		},
		ManagerConfig: ManagerConfig{
			PrimaryProtocol: "tcp",
			FallbackOrder:   []string{"tcp", "udp", "ws", "icmp", "dns"},
			SwitchThreshold: 3,
			MinSwitchInterval: 60,
		},
	}

	switcher := NewProtocolSwitcher(config)

	if switcher.switchStrategy != StrategySequential {
		t.Errorf("Expected switch strategy to be %s, got %s", StrategySequential, switcher.switchStrategy)
	}

	if switcher.jitterRange[0] != 1 || switcher.jitterRange[1] != 5 {
		t.Errorf("Expected jitter range to be [1, 5], got %v", switcher.jitterRange)
	}

	if switcher.IsRunning() {
		t.Error("Expected switcher to be stopped initially")
	}
}

func TestProtocolSwitcher_RegisterProtocol(t *testing.T) {
	switcher := NewProtocolSwitcher(ProtocolSwitcherConfig{
		ManagerConfig: ManagerConfig{
			PrimaryProtocol: "tcp",
			FallbackOrder:   []string{"tcp", "udp"},
		},
	})

	// Register a protocol
	protocol1 := NewMockProtocol("tcp")
	err := switcher.RegisterProtocol(protocol1)
	if err != nil {
		t.Errorf("Failed to register protocol: %v", err)
	}

	// Check that the protocol was registered
	p, err := switcher.GetProtocol("tcp")
	if err != nil {
		t.Errorf("Failed to get protocol: %v", err)
	}
	if p != protocol1 {
		t.Error("Expected to get the registered protocol")
	}

	// Register another protocol
	protocol2 := NewMockProtocol("udp")
	err = switcher.RegisterProtocol(protocol2)
	if err != nil {
		t.Errorf("Failed to register protocol: %v", err)
	}

	// Check that both protocols are registered
	p, err = switcher.GetProtocol("udp")
	if err != nil {
		t.Errorf("Failed to get protocol: %v", err)
	}
	if p != protocol2 {
		t.Error("Expected to get the registered protocol")
	}
}

func TestProtocolSwitcher_SetActiveProtocol(t *testing.T) {
	switcher := NewProtocolSwitcher(ProtocolSwitcherConfig{
		ManagerConfig: ManagerConfig{
			PrimaryProtocol: "tcp",
			FallbackOrder:   []string{"tcp", "udp"},
		},
	})

	// Register protocols
	protocol1 := NewMockProtocol("tcp")
	protocol2 := NewMockProtocol("udp")
	switcher.RegisterProtocol(protocol1)
	switcher.RegisterProtocol(protocol2)

	// Check that the primary protocol is active
	active := switcher.GetActiveProtocol()
	if active != protocol1 {
		t.Error("Expected primary protocol to be active")
	}

	// Set the second protocol as active
	err := switcher.SetActiveProtocol("udp")
	if err != nil {
		t.Errorf("Failed to set active protocol: %v", err)
	}

	// Check that the second protocol is now active
	active = switcher.GetActiveProtocol()
	if active != protocol2 {
		t.Error("Expected second protocol to be active")
	}
}

func TestProtocolSwitcher_StartStop(t *testing.T) {
	switcher := NewProtocolSwitcher(ProtocolSwitcherConfig{})

	// Check initial state
	if switcher.IsRunning() {
		t.Error("Expected switcher to be stopped initially")
	}

	// Start the switcher
	err := switcher.Start()
	if err != nil {
		t.Errorf("Failed to start switcher: %v", err)
	}

	// Check that the switcher is running
	if !switcher.IsRunning() {
		t.Error("Expected switcher to be running after Start")
	}

	// Stop the switcher
	switcher.Stop()

	// Check that the switcher is stopped
	if switcher.IsRunning() {
		t.Error("Expected switcher to be stopped after Stop")
	}
}

func TestProtocolSwitcher_SetSwitchStrategy(t *testing.T) {
	switcher := NewProtocolSwitcher(ProtocolSwitcherConfig{
		SwitchStrategy: StrategySequential,
	})

	// Check initial strategy
	if switcher.switchStrategy != StrategySequential {
		t.Errorf("Expected initial strategy to be %s, got %s", StrategySequential, switcher.switchStrategy)
	}

	// Set a new strategy
	switcher.SetSwitchStrategy(StrategyRandom)

	// Check that the strategy was updated
	if switcher.switchStrategy != StrategyRandom {
		t.Errorf("Expected strategy to be updated to %s, got %s", StrategyRandom, switcher.switchStrategy)
	}
}

func TestProtocolSwitcher_SetJitterRange(t *testing.T) {
	switcher := NewProtocolSwitcher(ProtocolSwitcherConfig{
		JitterMin: 1,
		JitterMax: 5,
	})

	// Check initial range
	if switcher.jitterRange[0] != 1 || switcher.jitterRange[1] != 5 {
		t.Errorf("Expected initial jitter range to be [1, 5], got %v", switcher.jitterRange)
	}

	// Set a new range
	switcher.SetJitterRange(2, 10)

	// Check that the range was updated
	if switcher.jitterRange[0] != 2 || switcher.jitterRange[1] != 10 {
		t.Errorf("Expected jitter range to be updated to [2, 10], got %v", switcher.jitterRange)
	}

	// Try to set an invalid range
	switcher.SetJitterRange(-1, 0)

	// Check that the range was corrected
	if switcher.jitterRange[0] != 0 || switcher.jitterRange[1] != 30 {
		t.Errorf("Expected jitter range to be corrected to [0, 30], got %v", switcher.jitterRange)
	}
}

func TestProtocolSwitcher_SwitchToNextProtocol(t *testing.T) {
	switcher := NewProtocolSwitcher(ProtocolSwitcherConfig{
		SwitchStrategy: StrategySequential,
		JitterMin:      0, // No jitter for testing
		JitterMax:      0,
		ManagerConfig: ManagerConfig{
			PrimaryProtocol: "tcp",
			FallbackOrder:   []string{"tcp", "udp", "ws"},
		},
	})

	// Register protocols
	protocol1 := NewMockProtocol("tcp")
	protocol2 := NewMockProtocol("udp")
	protocol3 := NewMockProtocol("ws")
	switcher.RegisterProtocol(protocol1)
	switcher.RegisterProtocol(protocol2)
	switcher.RegisterProtocol(protocol3)

	// Set up a switch handler to detect protocol switches
	switchCalled := false
	var oldProto, newProto Protocol
	switcher.SetOnSwitchHandler(func(old, new Protocol) {
		switchCalled = true
		oldProto = old
		newProto = new
	})

	// Check that the primary protocol is active
	active := switcher.GetActiveProtocol()
	if active != protocol1 {
		t.Error("Expected primary protocol to be active")
	}

	// Switch to the next protocol
	err := switcher.SwitchToNextProtocol()
	if err != nil {
		t.Errorf("Failed to switch protocol: %v", err)
	}

	// Check that the switch handler was called
	if !switchCalled {
		t.Error("Expected switch handler to be called")
	}

	// Check that the old and new protocols are correct
	if oldProto != protocol1 {
		t.Errorf("Expected old protocol to be %s, got %s", protocol1.GetName(), oldProto.GetName())
	}
	if newProto != protocol2 {
		t.Errorf("Expected new protocol to be %s, got %s", protocol2.GetName(), newProto.GetName())
	}

	// Check that the active protocol was switched
	active = switcher.GetActiveProtocol()
	if active != protocol2 {
		t.Errorf("Expected active protocol to be %s, got %s", protocol2.GetName(), active.GetName())
	}

	// Reset the switch handler flag
	switchCalled = false

	// Switch to the next protocol again
	err = switcher.SwitchToNextProtocol()
	if err != nil {
		t.Errorf("Failed to switch protocol: %v", err)
	}

	// Check that the switch handler was called
	if !switchCalled {
		t.Error("Expected switch handler to be called")
	}

	// Check that the active protocol was switched
	active = switcher.GetActiveProtocol()
	if active != protocol3 {
		t.Errorf("Expected active protocol to be %s, got %s", protocol3.GetName(), active.GetName())
	}

	// Reset the switch handler flag
	switchCalled = false

	// Switch to the next protocol again (should wrap around to the first one)
	err = switcher.SwitchToNextProtocol()
	if err != nil {
		t.Errorf("Failed to switch protocol: %v", err)
	}

	// Check that the switch handler was called
	if !switchCalled {
		t.Error("Expected switch handler to be called")
	}

	// Check that the active protocol was switched back to the first one
	active = switcher.GetActiveProtocol()
	if active != protocol1 {
		t.Errorf("Expected active protocol to be %s, got %s", protocol1.GetName(), active.GetName())
	}
}

func TestProtocolSwitcher_RandomStrategy(t *testing.T) {
	switcher := NewProtocolSwitcher(ProtocolSwitcherConfig{
		SwitchStrategy: StrategyRandom,
		JitterMin:      0, // No jitter for testing
		JitterMax:      0,
	})

	// Register protocols
	protocol1 := NewMockProtocol("tcp")
	protocol2 := NewMockProtocol("udp")
	protocol3 := NewMockProtocol("ws")
	switcher.RegisterProtocol(protocol1)
	switcher.RegisterProtocol(protocol2)
	switcher.RegisterProtocol(protocol3)

	// Set the first protocol as active
	switcher.SetActiveProtocol("tcp")

	// Switch to a random protocol
	err := switcher.SwitchToNextProtocol()
	if err != nil {
		t.Errorf("Failed to switch protocol: %v", err)
	}

	// Check that the active protocol was switched to something else
	active := switcher.GetActiveProtocol()
	if active == protocol1 {
		t.Error("Expected active protocol to be different from the initial one")
	}
}

// This test simulates a timeout and protocol switch scenario
func TestProtocolSwitcher_TimeoutAndSwitch(t *testing.T) {
	// Create a switcher with a low timeout threshold
	switcher := NewProtocolSwitcher(ProtocolSwitcherConfig{
		SwitchStrategy: StrategySequential,
		JitterMin:      0, // No jitter for testing
		JitterMax:      0,
		TimeoutDetectorConfig: TimeoutDetectorConfig{
			TimeoutThreshold: 2, // Switch after 2 timeouts
			CheckInterval:    1,
			MaxInactivity:    5,
		},
		ManagerConfig: ManagerConfig{
			PrimaryProtocol: "tcp",
			FallbackOrder:   []string{"tcp", "udp"},
			SwitchThreshold: 2,
			MinSwitchInterval: 0,
		},
	})

	// Register protocols
	protocol1 := NewMockProtocol("tcp")
	protocol2 := NewMockProtocol("udp")
	switcher.RegisterProtocol(protocol1)
	switcher.RegisterProtocol(protocol2)

	// Set up a switch handler to detect protocol switches
	switchCalled := false
	switcher.SetOnSwitchHandler(func(old, new Protocol) {
		switchCalled = true
	})

	// Start the switcher
	switcher.Start()

	// Simulate timeouts
	switcher.detector.RecordTimeout()
	switcher.detector.RecordTimeout() // This should trigger a switch

	// Give the switcher time to process the timeouts
	time.Sleep(100 * time.Millisecond)

	// Check that the switch handler was called
	if !switchCalled {
		t.Error("Expected switch handler to be called")
	}

	// Check that the active protocol was switched
	active := switcher.GetActiveProtocol()
	if active != protocol2 {
		t.Errorf("Expected active protocol to be %s, got %s", protocol2.GetName(), active.GetName())
	}

	// Stop the switcher
	switcher.Stop()
}

// This test checks that the Send and Receive methods record activity
func TestProtocolSwitcher_SendReceiveActivity(t *testing.T) {
	switcher := NewProtocolSwitcher(ProtocolSwitcherConfig{})

	// Register a protocol
	protocol := NewMockProtocol("tcp")
	switcher.RegisterProtocol(protocol)

	// Record the initial last activity time
	initialActivity := switcher.detector.GetLastActivity()

	// Wait a bit to ensure the timestamps will be different
	time.Sleep(10 * time.Millisecond)

	// Send some data
	_, err := switcher.Send([]byte("test"))
	if err != nil {
		t.Errorf("Failed to send data: %v", err)
	}

	// Check that the last activity time was updated
	newActivity := switcher.detector.GetLastActivity()
	if !newActivity.After(initialActivity) {
		t.Error("Expected last activity time to be updated after Send")
	}

	// Wait a bit to ensure the timestamps will be different
	time.Sleep(10 * time.Millisecond)
	initialActivity = newActivity

	// Receive some data
	_, err = switcher.Receive()
	if err != nil {
		t.Errorf("Failed to receive data: %v", err)
	}

	// Check that the last activity time was updated
	newActivity = switcher.detector.GetLastActivity()
	if !newActivity.After(initialActivity) {
		t.Error("Expected last activity time to be updated after Receive")
	}
}
