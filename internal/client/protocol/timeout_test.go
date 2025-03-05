package protocol

import (
	"context"
	"net"
	"testing"
	"time"
)

// MockProtocol is a mock implementation of the Protocol interface for testing
type MockProtocol struct {
	name   string
	status Status
}

func NewMockProtocol(name string) *MockProtocol {
	return &MockProtocol{
		name:   name,
		status: StatusDisconnected,
	}
}

func (m *MockProtocol) Connect(ctx context.Context) error {
	m.status = StatusConnected
	return nil
}

func (m *MockProtocol) Disconnect() error {
	m.status = StatusDisconnected
	return nil
}

func (m *MockProtocol) Send(data []byte) (int, error) {
	return len(data), nil
}

func (m *MockProtocol) Receive() ([]byte, error) {
	return []byte("test"), nil
}

func (m *MockProtocol) GetName() string {
	return m.name
}

func (m *MockProtocol) GetStatus() Status {
	return m.status
}

func (m *MockProtocol) GetConfig() Config {
	return Config{}
}

func (m *MockProtocol) UpdateConfig(config Config) error {
	return nil
}

func (m *MockProtocol) IsConnected() bool {
	return m.status == StatusConnected
}

func (m *MockProtocol) GetLastError() error {
	return nil
}

func (m *MockProtocol) GetConnection() net.Conn {
	return nil
}

func TestTimeoutDetector_New(t *testing.T) {
	config := TimeoutDetectorConfig{
		TimeoutThreshold: 3,
		CheckInterval:    5,
		MaxInactivity:    60,
	}

	detector := NewTimeoutDetector(config)

	if detector.timeoutThreshold != config.TimeoutThreshold {
		t.Errorf("Expected timeout threshold to be %d, got %d", config.TimeoutThreshold, detector.timeoutThreshold)
	}

	if detector.checkInterval != time.Duration(config.CheckInterval)*time.Second {
		t.Errorf("Expected check interval to be %s, got %s", time.Duration(config.CheckInterval)*time.Second, detector.checkInterval)
	}

	if detector.maxInactivity != time.Duration(config.MaxInactivity)*time.Second {
		t.Errorf("Expected max inactivity to be %s, got %s", time.Duration(config.MaxInactivity)*time.Second, detector.maxInactivity)
	}

	if detector.IsRunning() {
		t.Error("Expected detector to be stopped initially")
	}

	if detector.GetTimeoutCount() != 0 {
		t.Errorf("Expected initial timeout count to be 0, got %d", detector.GetTimeoutCount())
	}
}

func TestTimeoutDetector_RegisterProtocol(t *testing.T) {
	detector := NewTimeoutDetector(TimeoutDetectorConfig{})

	// Register a protocol
	protocol1 := NewMockProtocol("tcp")
	detector.RegisterProtocol(protocol1)

	// Check that the protocol was registered and set as active
	if detector.activeProtocol != protocol1 {
		t.Error("Expected protocol to be set as active")
	}

	// Register another protocol
	protocol2 := NewMockProtocol("udp")
	detector.RegisterProtocol(protocol2)

	// Check that the first protocol is still active
	if detector.activeProtocol != protocol1 {
		t.Error("Expected first protocol to remain active")
	}
}

func TestTimeoutDetector_SetActiveProtocol(t *testing.T) {
	detector := NewTimeoutDetector(TimeoutDetectorConfig{})

	// Register protocols
	protocol1 := NewMockProtocol("tcp")
	protocol2 := NewMockProtocol("udp")
	detector.RegisterProtocol(protocol1)
	detector.RegisterProtocol(protocol2)

	// Set the second protocol as active
	detector.SetActiveProtocol(protocol2)

	// Check that the second protocol is now active
	if detector.activeProtocol != protocol2 {
		t.Error("Expected second protocol to be active")
	}

	// Check that the timeout count was reset
	if detector.GetTimeoutCount() != 0 {
		t.Errorf("Expected timeout count to be reset to 0, got %d", detector.GetTimeoutCount())
	}
}

func TestTimeoutDetector_RecordActivity(t *testing.T) {
	detector := NewTimeoutDetector(TimeoutDetectorConfig{})

	// Record some timeouts
	detector.RecordTimeout()
	detector.RecordTimeout()

	// Check that the timeout count was incremented
	if detector.GetTimeoutCount() != 2 {
		t.Errorf("Expected timeout count to be 2, got %d", detector.GetTimeoutCount())
	}

	// Record an activity
	detector.RecordActivity()

	// Check that the timeout count was reset
	if detector.GetTimeoutCount() != 0 {
		t.Errorf("Expected timeout count to be reset to 0, got %d", detector.GetTimeoutCount())
	}

	// Check that the last activity time was updated
	lastActivity := detector.GetLastActivity()
	if time.Since(lastActivity) > time.Second {
		t.Error("Expected last activity time to be recent")
	}
}

func TestTimeoutDetector_RecordTimeout(t *testing.T) {
	detector := NewTimeoutDetector(TimeoutDetectorConfig{
		TimeoutThreshold: 2,
	})

	// Register protocols
	protocol1 := NewMockProtocol("tcp")
	protocol2 := NewMockProtocol("udp")
	detector.RegisterProtocol(protocol1)
	detector.RegisterProtocol(protocol2)

	// Set the first protocol as active
	detector.SetActiveProtocol(protocol1)

	// Set up a switch handler to detect protocol switches
	switchCalled := false
	detector.SetOnSwitchHandler(func(oldProtocol, newProtocol Protocol) {
		if oldProtocol != protocol1 || newProtocol != protocol2 {
			t.Error("Expected switch from protocol1 to protocol2")
		}
		switchCalled = true
	})

	// Record a timeout
	detector.RecordTimeout()

	// Check that the timeout count was incremented
	if detector.GetTimeoutCount() != 1 {
		t.Errorf("Expected timeout count to be 1, got %d", detector.GetTimeoutCount())
	}

	// Record another timeout to trigger a switch
	detector.RecordTimeout()

	// Check that the switch handler was called
	if !switchCalled {
		t.Error("Expected switch handler to be called")
	}

	// Check that the active protocol was switched
	if detector.activeProtocol != protocol2 {
		t.Error("Expected active protocol to be switched to protocol2")
	}

	// Check that the timeout count was reset
	if detector.GetTimeoutCount() != 0 {
		t.Errorf("Expected timeout count to be reset to 0, got %d", detector.GetTimeoutCount())
	}
}

func TestTimeoutDetector_StartStop(t *testing.T) {
	detector := NewTimeoutDetector(TimeoutDetectorConfig{})

	// Check initial state
	if detector.IsRunning() {
		t.Error("Expected detector to be stopped initially")
	}

	// Start the detector
	detector.Start()

	// Check that the detector is running
	if !detector.IsRunning() {
		t.Error("Expected detector to be running after Start")
	}

	// Stop the detector
	detector.Stop()

	// Check that the detector is stopped
	if detector.IsRunning() {
		t.Error("Expected detector to be stopped after Stop")
	}
}

func TestTimeoutDetector_SetTimeoutThreshold(t *testing.T) {
	detector := NewTimeoutDetector(TimeoutDetectorConfig{
		TimeoutThreshold: 3,
	})

	// Check initial value
	if detector.timeoutThreshold != 3 {
		t.Errorf("Expected initial timeout threshold to be 3, got %d", detector.timeoutThreshold)
	}

	// Set a new value
	detector.SetTimeoutThreshold(5)

	// Check that the value was updated
	if detector.timeoutThreshold != 5 {
		t.Errorf("Expected timeout threshold to be updated to 5, got %d", detector.timeoutThreshold)
	}

	// Try to set an invalid value
	detector.SetTimeoutThreshold(0)

	// Check that the value was not updated
	if detector.timeoutThreshold != 5 {
		t.Errorf("Expected timeout threshold to remain 5, got %d", detector.timeoutThreshold)
	}
}

func TestTimeoutDetector_SetCheckInterval(t *testing.T) {
	detector := NewTimeoutDetector(TimeoutDetectorConfig{
		CheckInterval: 5,
	})

	// Check initial value
	if detector.checkInterval != 5*time.Second {
		t.Errorf("Expected initial check interval to be 5s, got %s", detector.checkInterval)
	}

	// Set a new value
	detector.SetCheckInterval(10)

	// Check that the value was updated
	if detector.checkInterval != 10*time.Second {
		t.Errorf("Expected check interval to be updated to 10s, got %s", detector.checkInterval)
	}

	// Try to set an invalid value
	detector.SetCheckInterval(0)

	// Check that the value was not updated
	if detector.checkInterval != 10*time.Second {
		t.Errorf("Expected check interval to remain 10s, got %s", detector.checkInterval)
	}
}

func TestTimeoutDetector_SetMaxInactivity(t *testing.T) {
	detector := NewTimeoutDetector(TimeoutDetectorConfig{
		MaxInactivity: 60,
	})

	// Check initial value
	if detector.maxInactivity != 60*time.Second {
		t.Errorf("Expected initial max inactivity to be 60s, got %s", detector.maxInactivity)
	}

	// Set a new value
	detector.SetMaxInactivity(120)

	// Check that the value was updated
	if detector.maxInactivity != 120*time.Second {
		t.Errorf("Expected max inactivity to be updated to 120s, got %s", detector.maxInactivity)
	}

	// Try to set an invalid value
	detector.SetMaxInactivity(0)

	// Check that the value was not updated
	if detector.maxInactivity != 120*time.Second {
		t.Errorf("Expected max inactivity to remain 120s, got %s", detector.maxInactivity)
	}
}
