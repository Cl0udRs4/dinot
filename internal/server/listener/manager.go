package listener

import (
	"context"
	"fmt"
	"sync"

	"github.com/Cl0udRs4/dinot/internal/server/common"
)

// ListenerManager manages multiple protocol listeners
type ListenerManager struct {
	listeners     map[string]Listener
	listenersMtx  sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	handler       ConnectionHandler
	defaultConfig Config
}

// NewListenerManager creates a new listener manager
func NewListenerManager(defaultConfig Config) *ListenerManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ListenerManager{
		listeners:     make(map[string]Listener),
		ctx:           ctx,
		cancel:        cancel,
		defaultConfig: defaultConfig,
	}
}

// RegisterListener registers a listener with the manager
func (m *ListenerManager) RegisterListener(listener Listener) error {
	protocol := listener.GetProtocol()
	
	m.listenersMtx.Lock()
	defer m.listenersMtx.Unlock()
	
	if _, exists := m.listeners[protocol]; exists {
		return common.NewServerError(common.ErrListenerAlreadyRegistered, 
			fmt.Sprintf("listener for protocol %s is already registered", protocol), nil)
	}
	
	m.listeners[protocol] = listener
	return nil
}

// UnregisterListener unregisters a listener from the manager
func (m *ListenerManager) UnregisterListener(protocol string) error {
	m.listenersMtx.Lock()
	defer m.listenersMtx.Unlock()
	
	listener, exists := m.listeners[protocol]
	if !exists {
		return common.NewServerError(common.ErrListenerNotRegistered, 
			fmt.Sprintf("listener for protocol %s is not registered", protocol), nil)
	}
	
	// Stop the listener if it is running
	if listener.GetStatus() == StatusRunning {
		if err := listener.Stop(); err != nil {
			return err
		}
	}
	
	delete(m.listeners, protocol)
	return nil
}

// GetListener gets a listener by protocol
func (m *ListenerManager) GetListener(protocol string) (Listener, error) {
	m.listenersMtx.RLock()
	defer m.listenersMtx.RUnlock()
	
	listener, exists := m.listeners[protocol]
	if !exists {
		return nil, common.NewServerError(common.ErrListenerNotRegistered, 
			fmt.Sprintf("listener for protocol %s is not registered", protocol), nil)
	}
	
	return listener, nil
}
// StartAll starts all registered listeners
func (m *ListenerManager) StartAll(handler ConnectionHandler) error {
	m.listenersMtx.Lock()
	defer m.listenersMtx.Unlock()
	
	// Store the connection handler
	m.handler = handler
	
	// Start all listeners
	for protocol, listener := range m.listeners {
		if listener.GetStatus() == StatusRunning {
			continue
		}
		
		if err := listener.Start(m.ctx, handler); err != nil {
			return common.NewServerError(common.ErrListenerStartFailed, 
				fmt.Sprintf("failed to start listener for protocol %s", protocol), err)
		}
	}
	
	return nil
}

// GetListeners returns a map of all registered listeners
func (m *ListenerManager) GetListeners() map[string]Listener {
	m.listenersMtx.RLock()
	defer m.listenersMtx.RUnlock()
	
	// Create a copy of the listeners map to avoid concurrent access issues
	listeners := make(map[string]Listener, len(m.listeners))
	for protocol, listener := range m.listeners {
		listeners[protocol] = listener
	}
	
	return listeners
}

// HaltAll halts all registered listeners
func (m *ListenerManager) HaltAll() error {
	m.listenersMtx.Lock()
	defer m.listenersMtx.Unlock()
	
	var lastErr error
	
	// Halt all listeners
	for protocol, listener := range m.listeners {
		if listener.GetStatus() != StatusRunning {
			continue
		}
		
		if err := listener.Stop(); err != nil {
			lastErr = common.NewServerError(common.ErrListenerStopFailed, 
				fmt.Sprintf("failed to halt listener for protocol %s", protocol), err)
		}
	}
	
	// Cancel the context to signal all listeners to halt
	m.cancel()
	
	return lastErr
}

// CreateListener creates a new listener for the specified protocol
func (m *ListenerManager) CreateListener(protocol string, config Config) (Listener, error) {
	var listener Listener
	
	// Create a new listener based on the protocol
	switch protocol {
	case "tcp":
		listener = NewTCPListener(config)
	case "udp":
		listener = NewUDPListener(config)
	case "ws":
		listener = NewWSListener(config)
	case "icmp":
		listener = NewICMPListener(config)
	case "dns":
		// For DNS, we need to convert the config to DNSConfig
		dnsConfig := DNSConfig{
			Config:      config,
			Domain:      "example.com", // Default domain
			TTL:         60,            // Default TTL
			RecordTypes: []string{"A", "TXT"},
		}
		listener = NewDNSListener(dnsConfig)
	default:
		return nil, common.NewServerError(common.ErrInvalidConfig, 
			fmt.Sprintf("unsupported protocol: %s", protocol), nil)
	}
	
	// Register the listener
	if err := m.RegisterListener(listener); err != nil {
		return nil, err
	}
	
	return listener, nil
}

// CleanupManager closes the listener manager and all registered listeners
func (m *ListenerManager) CleanupManager() error {
	// Halt all listeners
	if err := m.HaltAll(); err != nil {
		return err
	}
	
	// Clear the listeners map
	m.listenersMtx.Lock()
	m.listeners = make(map[string]Listener)
	m.listenersMtx.Unlock()
	
	return nil
}
