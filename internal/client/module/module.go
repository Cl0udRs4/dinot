package module

import (
    "context"
    "encoding/json"
)

// Module defines the interface for client modules
type Module interface {
    // GetName returns the module name
    GetName() string
    
    // Execute executes the module with the given parameters
    Execute(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
    
    // Init initializes the module
    Init() error
    
    // Cleanup performs cleanup when the module is unloaded
    Cleanup() error
}

// BaseModule provides common functionality for all modules
type BaseModule struct {
    Name string
}

// GetName returns the module name
func (m *BaseModule) GetName() string {
    return m.Name
}

// Init initializes the module
func (m *BaseModule) Init() error {
    return nil
}

// Cleanup performs cleanup when the module is unloaded
func (m *BaseModule) Cleanup() error {
    return nil
}
