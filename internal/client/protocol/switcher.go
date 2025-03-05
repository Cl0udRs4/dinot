package protocol

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"
)

// ProtocolSwitcher is responsible for managing protocol switching
// based on timeout detection and other factors
type ProtocolSwitcher struct {
	// manager is the protocol manager
	manager *ProtocolManager
	
	// detector is the timeout detector
	detector *TimeoutDetector
	
	// switchStrategy determines how to select the next protocol
	switchStrategy SwitchStrategy
	
	// jitterRange is the range for random jitter in seconds
	jitterRange [2]int
	
	// mu protects concurrent access to the switcher
	mu sync.RWMutex
	
	// ctx is the context for the switcher
	ctx context.Context
	
	// cancel is the function to cancel the switcher context
	cancel context.CancelFunc
	
	// running indicates if the switcher is running
	running bool
	
	// onSwitch is called when a protocol switch occurs
	onSwitch func(oldProtocol, newProtocol Protocol)
}

// SwitchStrategy defines the strategy for selecting the next protocol
type SwitchStrategy string

const (
	// StrategySequential tries protocols in the order defined in the fallback list
	StrategySequential SwitchStrategy = "sequential"
	
	// StrategyRandom selects a random protocol from the available ones
	StrategyRandom SwitchStrategy = "random"
	
	// StrategyWeighted selects a protocol based on weighted success rates
	StrategyWeighted SwitchStrategy = "weighted"
)

// ProtocolSwitcherConfig defines the configuration for the protocol switcher
type ProtocolSwitcherConfig struct {
	// SwitchStrategy determines how to select the next protocol
	SwitchStrategy SwitchStrategy
	
	// JitterMin is the minimum jitter in seconds
	JitterMin int
	
	// JitterMax is the maximum jitter in seconds
	JitterMax int
	
	// TimeoutDetectorConfig is the configuration for the timeout detector
	TimeoutDetectorConfig TimeoutDetectorConfig
	
	// ManagerConfig is the configuration for the protocol manager
	ManagerConfig ManagerConfig
}

// NewProtocolSwitcher creates a new protocol switcher
func NewProtocolSwitcher(config ProtocolSwitcherConfig) *ProtocolSwitcher {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Set default values if not provided
	if config.SwitchStrategy == "" {
		config.SwitchStrategy = StrategySequential
	}
	
	if config.JitterMin < 0 {
		config.JitterMin = 0
	}
	
	if config.JitterMax <= config.JitterMin {
		config.JitterMax = config.JitterMin + 30
	}
	
	// Create the protocol manager
	manager := NewProtocolManager(config.ManagerConfig)
	
	// Create the timeout detector
	detector := NewTimeoutDetector(config.TimeoutDetectorConfig)
	
	switcher := &ProtocolSwitcher{
		manager:        manager,
		detector:       detector,
		switchStrategy: config.SwitchStrategy,
		jitterRange:    [2]int{config.JitterMin, config.JitterMax},
		ctx:            ctx,
		cancel:         cancel,
		running:        false,
	}
	
	// Set up the timeout detector handlers
	detector.SetOnTimeoutHandler(switcher.handleTimeout)
	detector.SetOnSwitchHandler(switcher.handleSwitch)
	
	return switcher
}

// RegisterProtocol registers a protocol with the switcher
func (s *ProtocolSwitcher) RegisterProtocol(protocol Protocol) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Register with the manager
	err := s.manager.RegisterProtocol(protocol)
	if err != nil {
		return err
	}
	
	// Register with the detector
	s.detector.RegisterProtocol(protocol)
	
	return nil
}

// SetActiveProtocol sets the active protocol
func (s *ProtocolSwitcher) SetActiveProtocol(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Set the active protocol in the manager
	err := s.manager.SetActiveProtocol(name)
	if err != nil {
		return err
	}
	
	// Get the active protocol
	protocol := s.manager.GetActiveProtocol()
	
	// Set the active protocol in the detector
	s.detector.SetActiveProtocol(protocol)
	
	return nil
}

// Start starts the protocol switcher
func (s *ProtocolSwitcher) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.running {
		return nil
	}
	
	// Start the timeout detector
	s.detector.Start()
	
	s.running = true
	return nil
}

// Stop stops the protocol switcher
func (s *ProtocolSwitcher) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.running {
		return
	}
	
	// Stop the timeout detector
	s.detector.Stop()
	
	// Cancel the context
	s.cancel()
	
	s.running = false
}

// Connect connects the active protocol
func (s *ProtocolSwitcher) Connect(ctx context.Context) error {
	// Connect using the manager
	err := s.manager.Connect(ctx)
	if err != nil {
		return err
	}
	
	// Record activity in the detector
	s.detector.RecordActivity()
	
	return nil
}

// Disconnect disconnects the active protocol
func (s *ProtocolSwitcher) Disconnect() error {
	// Disconnect using the manager
	return s.manager.Disconnect()
}

// Send sends data using the active protocol
func (s *ProtocolSwitcher) Send(data []byte) (int, error) {
	// Send using the manager
	n, err := s.manager.Send(data)
	
	// Handle the result
	if err != nil {
		// Check if it's a timeout error
		if IsTimeoutError(err) {
			s.detector.RecordTimeout()
		} else {
			// For other errors, just record activity to reset the timeout counter
			s.detector.RecordActivity()
		}
		return n, err
	}
	
	// Record successful activity
	s.detector.RecordActivity()
	
	return n, nil
}

// Receive receives data using the active protocol
func (s *ProtocolSwitcher) Receive() ([]byte, error) {
	// Receive using the manager
	data, err := s.manager.Receive()
	
	// Handle the result
	if err != nil {
		// Check if it's a timeout error
		if IsTimeoutError(err) {
			s.detector.RecordTimeout()
		} else {
			// For other errors, just record activity to reset the timeout counter
			s.detector.RecordActivity()
		}
		return data, err
	}
	
	// Record successful activity
	s.detector.RecordActivity()
	
	return data, nil
}

// GetActiveProtocol returns the currently active protocol
func (s *ProtocolSwitcher) GetActiveProtocol() Protocol {
	return s.manager.GetActiveProtocol()
}

// GetProtocol retrieves a protocol by name
func (s *ProtocolSwitcher) GetProtocol(name string) (Protocol, error) {
	return s.manager.GetProtocol(name)
}

// SetOnSwitchHandler sets the handler to be called when a protocol switch occurs
func (s *ProtocolSwitcher) SetOnSwitchHandler(handler func(oldProtocol, newProtocol Protocol)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.onSwitch = handler
}

// SetSwitchStrategy sets the strategy for selecting the next protocol
func (s *ProtocolSwitcher) SetSwitchStrategy(strategy SwitchStrategy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.switchStrategy = strategy
}

// SetJitterRange sets the range for random jitter
func (s *ProtocolSwitcher) SetJitterRange(min, max int) {
	if min < 0 {
		min = 0
	}
	
	if max <= min {
		max = min + 30
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.jitterRange = [2]int{min, max}
}

// IsRunning returns true if the switcher is running
func (s *ProtocolSwitcher) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.running
}

// handleTimeout handles a timeout event from the detector
func (s *ProtocolSwitcher) handleTimeout(protocol Protocol) {
	// No action needed here, the detector will trigger a switch if needed
}

// handleSwitch handles a protocol switch event from the detector
func (s *ProtocolSwitcher) handleSwitch(oldProtocol, newProtocol Protocol) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Apply jitter before switching
	jitter := s.getRandomJitter()
	time.Sleep(time.Duration(jitter) * time.Second)
	
	// Switch the active protocol in the manager
	if newProtocol != nil {
		s.manager.SetActiveProtocol(newProtocol.GetName())
		
		// Call the switch handler if set
		if s.onSwitch != nil {
			s.onSwitch(oldProtocol, newProtocol)
		}
	}
}

// getRandomJitter returns a random jitter value in seconds
func (s *ProtocolSwitcher) getRandomJitter() int {
	min := s.jitterRange[0]
	max := s.jitterRange[1]
	
	if min == max {
		return min
	}
	
	return min + rand.Intn(max-min)
}

// SwitchToNextProtocol manually triggers a switch to the next protocol
func (s *ProtocolSwitcher) SwitchToNextProtocol() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Get the current active protocol
	oldProtocol := s.manager.GetActiveProtocol()
	if oldProtocol == nil {
		return errors.New("no active protocol")
	}
	
	// Switch to the next protocol based on the strategy
	var nextProtocol Protocol
	var err error
	
	switch s.switchStrategy {
	case StrategyRandom:
		nextProtocol, err = s.selectRandomProtocol(oldProtocol)
	case StrategyWeighted:
		nextProtocol, err = s.selectWeightedProtocol(oldProtocol)
	default: // StrategySequential
		nextProtocol, err = s.selectSequentialProtocol(oldProtocol)
	}
	
	if err != nil {
		return err
	}
	
	// Apply jitter before switching
	jitter := s.getRandomJitter()
	time.Sleep(time.Duration(jitter) * time.Second)
	
	// Switch to the new protocol
	err = s.manager.SetActiveProtocol(nextProtocol.GetName())
	if err != nil {
		return err
	}
	
	// Update the detector
	s.detector.SetActiveProtocol(nextProtocol)
	
	// Call the switch handler if set
	if s.onSwitch != nil {
		s.onSwitch(oldProtocol, nextProtocol)
	}
	
	return nil
}

// selectSequentialProtocol selects the next protocol in the fallback order
func (s *ProtocolSwitcher) selectSequentialProtocol(currentProtocol Protocol) (Protocol, error) {
	return s.manager.selectNextProtocolInFallbackOrder(currentProtocol.GetName())
}

// selectRandomProtocol selects a random protocol different from the current one
func (s *ProtocolSwitcher) selectRandomProtocol(currentProtocol Protocol) (Protocol, error) {
	// Get all protocols
	protocols := s.manager.getAllProtocols()
	if len(protocols) <= 1 {
		return nil, errors.New("no alternative protocols available")
	}
	
	// Filter out the current protocol
	var alternatives []Protocol
	for _, p := range protocols {
		if p.GetName() != currentProtocol.GetName() {
			alternatives = append(alternatives, p)
		}
	}
	
	if len(alternatives) == 0 {
		return nil, errors.New("no alternative protocols available")
	}
	
	// Select a random protocol
	return alternatives[rand.Intn(len(alternatives))], nil
}

// selectWeightedProtocol selects a protocol based on weighted success rates
// This is a placeholder implementation that falls back to random selection
// In a real implementation, this would track success rates for each protocol
func (s *ProtocolSwitcher) selectWeightedProtocol(currentProtocol Protocol) (Protocol, error) {
	// For now, just use random selection
	return s.selectRandomProtocol(currentProtocol)
}

// Helper method to get all protocols from the manager
func (m *ProtocolManager) getAllProtocols() []Protocol {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	protocols := make([]Protocol, 0, len(m.protocols))
	for _, p := range m.protocols {
		protocols = append(protocols, p)
	}
	
	return protocols
}

// Helper method to select the next protocol in the fallback order
func (m *ProtocolManager) selectNextProtocolInFallbackOrder(currentName string) (Protocol, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if len(m.protocols) <= 1 {
		return nil, errors.New("no alternative protocols available")
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
			return protocol, nil
		}
	}
	
	// If we couldn't find the next protocol, try the first one in the fallback order
	if len(m.fallbackOrder) > 0 {
		for _, name := range m.fallbackOrder {
			if name != currentName {
				if protocol, exists := m.protocols[name]; exists {
					return protocol, nil
				}
			}
		}
	}
	
	// If all else fails, use any protocol that's not the current one
	for name, protocol := range m.protocols {
		if name != currentName {
			return protocol, nil
		}
	}
	
	return nil, errors.New("no alternative protocols available")
}
