package protocol

import (
    "context"
    "time"
)

// Protocol defines the interface for client communication protocols
type Protocol interface {
    // Connect establishes a connection to the server
    Connect(ctx context.Context) error
    
    // Disconnect closes the connection to the server
    Disconnect() error
    
    // Send sends data to the server
    Send(data []byte) error
    
    // Receive receives data from the server with timeout
    Receive(timeout time.Duration) ([]byte, error)
    
    // IsConnected returns true if the protocol is connected
    IsConnected() bool
    
    // GetName returns the protocol name
    GetName() string
}

// BaseProtocol provides common functionality for all protocols
type BaseProtocol struct {
    Name      string
    Connected bool
    Timeout   time.Duration
}

// GetName returns the protocol name
func (p *BaseProtocol) GetName() string {
    return p.Name
}

// IsConnected returns true if the protocol is connected
func (p *BaseProtocol) IsConnected() bool {
    return p.Connected
}
