package protocol

import (
    "context"
    "sync"
    "time"
)

// ProtocolManager manages multiple protocols and handles automatic switching
type ProtocolManager struct {
    protocols       []Protocol
    currentProtocol Protocol
    mu              sync.RWMutex
    switchThreshold int
    failCount       int
}

// NewProtocolManager creates a new protocol manager
func NewProtocolManager(protocols []Protocol, switchThreshold int) *ProtocolManager {
    var currentProtocol Protocol
    if len(protocols) > 0 {
        currentProtocol = protocols[0]
    }
    
    return &ProtocolManager{
        protocols:       protocols,
        currentProtocol: currentProtocol,
        switchThreshold: switchThreshold,
        failCount:       0,
    }
}

// Connect connects using the current protocol
func (m *ProtocolManager) Connect(ctx context.Context) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if m.currentProtocol == nil {
        return ErrNoProtocolAvailable
    }
    
    err := m.currentProtocol.Connect(ctx)
    if err != nil {
        m.failCount++
        if m.failCount >= m.switchThreshold {
            m.switchProtocol()
            m.failCount = 0
        }
        return err
    }
    
    m.failCount = 0
    return nil
}

// Disconnect disconnects the current protocol
func (m *ProtocolManager) Disconnect() error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if m.currentProtocol == nil {
        return nil
    }
    
    return m.currentProtocol.Disconnect()
}

// Send sends data using the current protocol
func (m *ProtocolManager) Send(data []byte) error {
    m.mu.RLock()
    protocol := m.currentProtocol
    m.mu.RUnlock()
    
    if protocol == nil {
        return ErrNoProtocolAvailable
    }
    
    err := protocol.Send(data)
    if err != nil {
        m.mu.Lock()
        m.failCount++
        if m.failCount >= m.switchThreshold {
            m.switchProtocol()
            m.failCount = 0
        }
        m.mu.Unlock()
        return err
    }
    
    m.mu.Lock()
    m.failCount = 0
    m.mu.Unlock()
    return nil
}

// Receive receives data using the current protocol
func (m *ProtocolManager) Receive(timeout time.Duration) ([]byte, error) {
    m.mu.RLock()
    protocol := m.currentProtocol
    m.mu.RUnlock()
    
    if protocol == nil {
        return nil, ErrNoProtocolAvailable
    }
    
    data, err := protocol.Receive(timeout)
    if err != nil {
        m.mu.Lock()
        m.failCount++
        if m.failCount >= m.switchThreshold {
            m.switchProtocol()
            m.failCount = 0
        }
        m.mu.Unlock()
        return nil, err
    }
    
    m.mu.Lock()
    m.failCount = 0
    m.mu.Unlock()
    return data, nil
}

// GetCurrentProtocol returns the current protocol
func (m *ProtocolManager) GetCurrentProtocol() Protocol {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.currentProtocol
}

// switchProtocol switches to the next available protocol
func (m *ProtocolManager) switchProtocol() {
    if len(m.protocols) <= 1 {
        return
    }
    
    currentIndex := -1
    for i, p := range m.protocols {
        if p == m.currentProtocol {
            currentIndex = i
            break
        }
    }
    
    nextIndex := (currentIndex + 1) % len(m.protocols)
    m.currentProtocol = m.protocols[nextIndex]
}
