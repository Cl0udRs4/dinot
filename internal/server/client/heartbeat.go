package client

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// HeartbeatMonitor monitors client heartbeats and updates their status
type HeartbeatMonitor struct {
	// clientManager is the client manager to monitor
	clientManager *ClientManager
	
	// checkInterval is how often to check for offline clients
	checkInterval time.Duration
	
	// timeout is how long to wait after a missed heartbeat before marking a client as offline
	timeout time.Duration
	
	// minRandomInterval is the minimum random heartbeat interval
	minRandomInterval time.Duration
	
	// maxRandomInterval is the maximum random heartbeat interval
	maxRandomInterval time.Duration
	
	// useRandomIntervals determines whether to use random heartbeat intervals
	useRandomIntervals bool
	
	// ctx is the context for controlling the monitor's lifecycle
	ctx context.Context
	
	// cancel is the function to cancel the monitor's context
	cancel context.CancelFunc
	
	// wg is used to wait for the monitor to shut down
	wg sync.WaitGroup
	
	// mu protects concurrent access to the monitor's configuration
	mu sync.RWMutex
}

// NewHeartbeatMonitor creates a new heartbeat monitor
func NewHeartbeatMonitor(clientManager *ClientManager, checkInterval, timeout time.Duration) *HeartbeatMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &HeartbeatMonitor{
		clientManager:      clientManager,
		checkInterval:      checkInterval,
		timeout:            timeout,
		minRandomInterval:  1 * time.Second,
		maxRandomInterval:  24 * time.Hour,
		useRandomIntervals: false,
		ctx:                ctx,
		cancel:             cancel,
	}
}

// Start starts the heartbeat monitor
func (m *HeartbeatMonitor) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.wg.Add(1)
	go m.monitorHeartbeats()
}

// Stop stops the heartbeat monitor
func (m *HeartbeatMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.cancel()
	m.wg.Wait()
}

// SetCheckInterval sets the interval at which to check for offline clients
func (m *HeartbeatMonitor) SetCheckInterval(interval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.checkInterval = interval
}

// SetTimeout sets how long to wait after a missed heartbeat before marking a client as offline
func (m *HeartbeatMonitor) SetTimeout(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.timeout = timeout
}

// EnableRandomIntervals enables random heartbeat intervals for clients
func (m *HeartbeatMonitor) EnableRandomIntervals(min, max time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.minRandomInterval = min
	m.maxRandomInterval = max
	m.useRandomIntervals = true
}

// DisableRandomIntervals disables random heartbeat intervals for clients
func (m *HeartbeatMonitor) DisableRandomIntervals() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.useRandomIntervals = false
}

// AssignRandomInterval assigns a random heartbeat interval to a client
func (m *HeartbeatMonitor) AssignRandomInterval(clientID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if !m.useRandomIntervals {
		return nil
	}
	
	client, err := m.clientManager.GetClient(clientID)
	if err != nil {
		return err
	}
	
	// Generate a random duration between min and max
	randomDuration := m.minRandomInterval + 
		time.Duration(rand.Int63n(int64(m.maxRandomInterval-m.minRandomInterval)))
	
	client.SetHeartbeatInterval(randomDuration)
	return nil
}

// monitorHeartbeats is the main goroutine that monitors client heartbeats
func (m *HeartbeatMonitor) monitorHeartbeats() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.checkClients()
		case <-m.ctx.Done():
			return
		}
	}
}

// checkClients checks all clients for heartbeat timeouts
func (m *HeartbeatMonitor) checkClients() {
	m.mu.RLock()
	timeout := m.timeout
	m.mu.RUnlock()
	
	_ = m.clientManager.CheckOfflineClients(timeout)
	
	// Here you could add additional logic to handle offline clients,
	// such as logging, sending notifications, or attempting to reconnect
}

// ProcessHeartbeat processes a heartbeat from a client
func (m *HeartbeatMonitor) ProcessHeartbeat(clientID string) error {
	// Update the client's last seen time
	err := m.clientManager.UpdateClientLastSeen(clientID)
	if err != nil {
		return err
	}
	
	// If the client was offline, mark it as online
	client, err := m.clientManager.GetClient(clientID)
	if err != nil {
		return err
	}
	
	if client.Status == StatusOffline {
		client.UpdateStatus(StatusOnline, "")
	}
	
	// Assign a new random interval if enabled
	m.mu.RLock()
	useRandom := m.useRandomIntervals
	m.mu.RUnlock()
	
	if useRandom {
		m.AssignRandomInterval(clientID)
	}
	
	return nil
}
