package protocol

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ProtocolManager manages multiple protocols and handles protocol switching
type ProtocolManager struct {
	// protocols is a map of protocol name to Protocol
	protocols map[string]Protocol
	
	// activeProtocol is the currently active protocol
	activeProtocol Protocol
	
	// primaryProtocol is the preferred protocol to use
	primaryProtocol string
	
	// fallbackOrder is the order to try protocols when the active one fails
	fallbackOrder []string
	
	// switchThreshold is the number of consecutive failures before switching protocols
	switchThreshold int
	
	// failureCount tracks consecutive failures for the active protocol
	failureCount int
	
	// lastSwitchTime is the time of the last protocol switch
	lastSwitchTime time.Time
	
	// minSwitchInterval is the minimum time between protocol switches
	minSwitchInterval time.Duration
	
	// mu protects concurrent access to the manager
	mu sync.RWMutex
	
	// ctx is the context for the manager
	ctx context.Context
	
	// cancel is the function to cancel the manager context
	cancel context.CancelFunc
}

// ManagerConfig defines the configuration for the protocol manager
type ManagerConfig struct {
	// PrimaryProtocol is the preferred protocol to use
	PrimaryProtocol string
	
	// FallbackOrder is the order to try protocols when the active one fails
	FallbackOrder []string
	
	// SwitchThreshold is the number of consecutive failures before switching protocols
	SwitchThreshold int
	
	// MinSwitchInterval is the minimum time between protocol switches in seconds
	MinSwitchInterval int
}

// NewProtocolManager creates a new protocol manager
func NewProtocolManager(config ManagerConfig) *ProtocolManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Set default values if not provided
	if config.SwitchThreshold <= 0 {
		config.SwitchThreshold = 3
	}
	
	if config.MinSwitchInterval <= 0 {
		config.MinSwitchInterval = 60
	}
	
	return &ProtocolManager{
		protocols:        make(map[string]Protocol),
		primaryProtocol:  config.PrimaryProtocol,
		fallbackOrder:    config.FallbackOrder,
		switchThreshold:  config.SwitchThreshold,
		failureCount:     0,
		minSwitchInterval: time.Duration(config.MinSwitchInterval) * time.Second,
		ctx:              ctx,
		cancel:           cancel,
	}
}

// RegisterProtocol registers a protocol with the manager
func (m *ProtocolManager) RegisterProtocol(protocol Protocol) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	name := protocol.GetName()
	
	// Check if the protocol is already registered
	if _, exists := m.protocols[name]; exists {
		return errors.New("protocol already registered")
	}
	
	// Add the protocol to the map
	m.protocols[name] = protocol
	
	// If this is the primary protocol and no active protocol is set, make it active
	if name == m.primaryProtocol && m.activeProtocol == nil {
		m.activeProtocol = protocol
	}
	
	// If no active protocol is set, make this the active protocol
	if m.activeProtocol == nil {
		m.activeProtocol = protocol
	}
	
	return nil
}

// UnregisterProtocol removes a protocol from the manager
func (m *ProtocolManager) UnregisterProtocol(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if the protocol exists
	protocol, exists := m.protocols[name]
	if !exists {
		return errors.New("protocol not registered")
	}
	
	// If this is the active protocol, disconnect it
	if m.activeProtocol == protocol {
		m.activeProtocol.Disconnect()
		m.activeProtocol = nil
		
		// Try to activate another protocol
		m.activateNextProtocol()
	}
	
	// Remove the protocol from the map
	delete(m.protocols, name)
	
	return nil
}

// GetProtocol retrieves a protocol by name
func (m *ProtocolManager) GetProtocol(name string) (Protocol, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	protocol, exists := m.protocols[name]
	if !exists {
		return nil, errors.New("protocol not registered")
	}
	
	return protocol, nil
}

// GetActiveProtocol returns the currently active protocol
func (m *ProtocolManager) GetActiveProtocol() Protocol {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.activeProtocol
}

// SetActiveProtocol sets the active protocol by name
func (m *ProtocolManager) SetActiveProtocol(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if the protocol exists
	protocol, exists := m.protocols[name]
	if !exists {
		return errors.New("protocol not registered")
	}
	
	// If this is already the active protocol, do nothing
	if m.activeProtocol == protocol {
		return nil
	}
	
	// Disconnect the current active protocol if any
	if m.activeProtocol != nil {
		m.activeProtocol.Disconnect()
	}
	
	// Set the new active protocol
	m.activeProtocol = protocol
	m.failureCount = 0
	m.lastSwitchTime = time.Now()
	
	return nil
}

// Connect connects the active protocol
func (m *ProtocolManager) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.activeProtocol == nil {
		return errors.New("no active protocol")
	}
	
	return m.activeProtocol.Connect(ctx)
}

// Disconnect disconnects the active protocol
func (m *ProtocolManager) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.activeProtocol == nil {
		return nil
	}
	
	return m.activeProtocol.Disconnect()
}

// Send sends data using the active protocol
func (m *ProtocolManager) Send(data []byte) (int, error) {
	m.mu.RLock()
	protocol := m.activeProtocol
	m.mu.RUnlock()
	
	if protocol == nil {
		return 0, errors.New("no active protocol")
	}
	
	n, err := protocol.Send(data)
	
	// Handle protocol switching if needed
	if err != nil {
		m.handleError(err)
		
		// If the error is ErrProtocolSwitch, try to switch protocols and retry
		if errors.Is(err, ErrProtocolSwitch) {
			m.mu.Lock()
			switched := m.switchProtocol()
			newProtocol := m.activeProtocol
			m.mu.Unlock()
			
			if switched && newProtocol != nil {
				// Try to connect with the new protocol
				if connectErr := newProtocol.Connect(m.ctx); connectErr == nil {
					// Retry the send with the new protocol
					return newProtocol.Send(data)
				}
			}
		}
	} else {
		// Reset failure count on success
		m.mu.Lock()
		m.failureCount = 0
		m.mu.Unlock()
	}
	
	return n, err
}

// Receive receives data using the active protocol
func (m *ProtocolManager) Receive() ([]byte, error) {
	m.mu.RLock()
	protocol := m.activeProtocol
	m.mu.RUnlock()
	
	if protocol == nil {
		return nil, errors.New("no active protocol")
	}
	
	data, err := protocol.Receive()
	
	// Handle protocol switching if needed
	if err != nil {
		m.handleError(err)
		
		// If the error is ErrProtocolSwitch, try to switch protocols and retry
		if errors.Is(err, ErrProtocolSwitch) {
			m.mu.Lock()
			switched := m.switchProtocol()
			newProtocol := m.activeProtocol
			m.mu.Unlock()
			
			if switched && newProtocol != nil {
				// Try to connect with the new protocol
				if connectErr := newProtocol.Connect(m.ctx); connectErr == nil {
					// Retry the receive with the new protocol
					return newProtocol.Receive()
				}
			}
		}
	} else {
		// Reset failure count on success
		m.mu.Lock()
		m.failureCount = 0
		m.mu.Unlock()
	}
	
	return data, err
}

// handleError increments the failure count and checks if a protocol switch is needed
func (m *ProtocolManager) handleError(err error) {
	if err == nil {
		return
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Increment failure count
	m.failureCount++
	
	// Check if we need to switch protocols
	if m.failureCount >= m.switchThreshold {
		m.switchProtocol()
	}
}

// switchProtocol switches to the next protocol in the fallback order
func (m *ProtocolManager) switchProtocol() bool {
	// Check if enough time has passed since the last switch
	if time.Since(m.lastSwitchTime) < m.minSwitchInterval {
		return false
	}
	
	// Disconnect the current active protocol
	if m.activeProtocol != nil {
		m.activeProtocol.Disconnect()
	}
	
	// Try to activate the next protocol
	if m.activateNextProtocol() {
		m.failureCount = 0
		m.lastSwitchTime = time.Now()
		return true
	}
	
	return false
}

// activateNextProtocol activates the next protocol in the fallback order
func (m *ProtocolManager) activateNextProtocol() bool {
	// If no protocols are registered, return false
	if len(m.protocols) == 0 {
		m.activeProtocol = nil
		return false
	}
	
	// If no active protocol is set, try to activate the primary protocol
	if m.activeProtocol == nil {
		if protocol, exists := m.protocols[m.primaryProtocol]; exists {
			m.activeProtocol = protocol
			return true
		}
	}
	
	// Get the current active protocol name
	var currentName string
	if m.activeProtocol != nil {
		currentName = m.activeProtocol.GetName()
	}
	
	// Find the current protocol in the fallback order
	currentIndex := -1
	for i, name := range m.fallbackOrder {
		if name == currentName {
			currentIndex = i
			break
		}
	}
	
	// Try the next protocol in the fallback order
	if currentIndex >= 0 && currentIndex < len(m.fallbackOrder)-1 {
		nextName := m.fallbackOrder[currentIndex+1]
		if protocol, exists := m.protocols[nextName]; exists {
			m.activeProtocol = protocol
			return true
		}
	}
	
	// If we couldn't find the next protocol, try the first one in the fallback order
	if len(m.fallbackOrder) > 0 {
		for _, name := range m.fallbackOrder {
			if protocol, exists := m.protocols[name]; exists {
				m.activeProtocol = protocol
				return true
			}
		}
	}
	
	// If all else fails, use the first registered protocol
	for _, protocol := range m.protocols {
		m.activeProtocol = protocol
		return true
	}
	
	return false
}

// Close closes the protocol manager and all registered protocols
func (m *ProtocolManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Cancel the context
	if m.cancel != nil {
		m.cancel()
	}
	
	// Disconnect all protocols
	for _, protocol := range m.protocols {
		protocol.Disconnect()
	}
	
	// Clear the protocols map
	m.protocols = make(map[string]Protocol)
	m.activeProtocol = nil
}
