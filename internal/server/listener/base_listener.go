package listener

import (
	"context"
	"sync"

	"github.com/Cl0udRs4/dinot/internal/server/common"
)

// BaseListener provides common functionality for all listeners
type BaseListener struct {
	// Protocol is the name of the protocol
	Protocol string
	
	// Config is the listener configuration
	Config Config
	
	// status is the current status of the listener
	status Status
	
	// statusMutex protects the status field
	statusMutex sync.RWMutex
	
	// cancel is the function to cancel the listener context
	cancel context.CancelFunc
}

// NewBaseListener creates a new base listener
func NewBaseListener(protocol string, config Config) *BaseListener {
	return &BaseListener{
		Protocol: protocol,
		Config:   config,
		status:   StatusStopped,
	}
}

// GetProtocol returns the protocol name
func (b *BaseListener) GetProtocol() string {
	return b.Protocol
}

// GetStatus returns the current status of the listener
func (b *BaseListener) GetStatus() Status {
	b.statusMutex.RLock()
	defer b.statusMutex.RUnlock()
	return b.status
}

// setStatus sets the status of the listener
func (b *BaseListener) setStatus(status Status) {
	b.statusMutex.Lock()
	defer b.statusMutex.Unlock()
	b.status = status
}

// GetConfig returns the current configuration of the listener
func (b *BaseListener) GetConfig() Config {
	return b.Config
}

// UpdateConfig updates the listener configuration
func (b *BaseListener) UpdateConfig(config Config) error {
	if b.GetStatus() == StatusRunning {
		return common.NewServerError(common.ErrListenerAlreadyRunning, "cannot update config while listener is running", nil)
	}
	b.Config = config
	return nil
}

// Start is a placeholder that should be overridden by specific listeners
func (b *BaseListener) Start(ctx context.Context, handler ConnectionHandler) error {
	return common.NewServerError(common.ErrNotImplemented, "Start method not implemented", nil)
}

// Stop is a placeholder that should be overridden by specific listeners
func (b *BaseListener) Stop() error {
	if b.cancel != nil {
		b.cancel()
		b.cancel = nil
	}
	b.setStatus(StatusStopped)
	return nil
}

// ValidateConfig validates the listener configuration
func (b *BaseListener) ValidateConfig() error {
	if b.Config.Address == "" {
		return common.NewServerError(common.ErrInvalidConfig, "address cannot be empty", nil)
	}
	
	if b.Config.BufferSize <= 0 {
		b.Config.BufferSize = 4096 // Default buffer size
	}
	
	if b.Config.MaxConnections <= 0 {
		b.Config.MaxConnections = 100 // Default max connections
	}
	
	if b.Config.Timeout <= 0 {
		b.Config.Timeout = 30 // Default timeout in seconds
	}
	
	return nil
}
