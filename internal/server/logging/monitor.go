package logging

import (
	"context"
	"sync"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/client"
)

// MonitorConfig represents the configuration for the monitoring manager
type MonitorConfig struct {
	// CheckInterval is the interval at which to check for exceptions
	CheckInterval time.Duration
	// ReconnectInterval is the interval at which to attempt reconnection
	ReconnectInterval time.Duration
	// MaxReconnectAttempts is the maximum number of reconnection attempts
	MaxReconnectAttempts int
}

// MonitorManager manages exception monitoring and automatic reconnection
type MonitorManager struct {
	// logger is the logger instance
	logger Logger
	// clientManager is the client manager instance
	clientManager *client.ClientManager
	// config is the monitor configuration
	config MonitorConfig
	// ctx is the context for cancellation
	ctx context.Context
	// cancel is the cancel function for the context
	cancel context.CancelFunc
	// wg is the wait group for goroutines
	wg sync.WaitGroup
	// mu protects concurrent access to the monitor manager
	mu sync.RWMutex
	// running indicates whether the monitor is running
	running bool
}

// NewMonitorManager creates a new monitor manager
func NewMonitorManager(logger Logger, clientManager *client.ClientManager, config MonitorConfig) *MonitorManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &MonitorManager{
		logger:        logger,
		clientManager: clientManager,
		config:        config,
		ctx:           ctx,
		cancel:        cancel,
		running:       false,
	}
}

// Start starts the monitor manager
func (m *MonitorManager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return
	}

	m.running = true
	m.wg.Add(1)
	go m.monitorExceptions()

	m.logger.Info("Monitor manager started", map[string]interface{}{
		"check_interval":        m.config.CheckInterval,
		"reconnect_interval":    m.config.ReconnectInterval,
		"max_reconnect_attempts": m.config.MaxReconnectAttempts,
	})
}

// Stop stops the monitor manager
func (m *MonitorManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.cancel()
	m.wg.Wait()
	m.running = false

	m.logger.Info("Monitor manager stopped", nil)
}

// monitorExceptions monitors for exceptions and attempts reconnection
func (m *MonitorManager) monitorExceptions() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkExceptions()
		}
	}
}

// checkExceptions checks for exceptions and attempts reconnection
func (m *MonitorManager) checkExceptions() {
	// Get all clients with error status
	clients := m.clientManager.GetClientsByStatus(client.StatusError)
	if len(clients) == 0 {
		return
	}

	m.logger.Info("Found clients with error status", map[string]interface{}{
		"count": len(clients),
	})

	// Get all exception reports
	exceptions := m.clientManager.GetAllExceptionReports()
	exceptionMap := make(map[string][]*client.ExceptionReport)
	for _, exception := range exceptions {
		exceptionMap[exception.ClientID] = append(exceptionMap[exception.ClientID], exception)
	}

	// Attempt reconnection for each client with error status
	for _, c := range clients {
		clientExceptions := exceptionMap[c.ID]
		if len(clientExceptions) == 0 {
			continue
		}

		// Log the exceptions
		m.logger.Error("Client has exceptions", map[string]interface{}{
			"client_id":     c.ID,
			"client_name":   c.Name,
			"client_status": c.Status,
			"exception_count": len(clientExceptions),
			"last_exception": clientExceptions[len(clientExceptions)-1].Message,
		})

		// Attempt reconnection
		go m.attemptReconnection(c)
	}
}

// attemptReconnection attempts to reconnect to a client
func (m *MonitorManager) attemptReconnection(c *client.Client) {
	m.logger.Info("Attempting reconnection", map[string]interface{}{
		"client_id":   c.ID,
		"client_name": c.Name,
	})

	// Simulate reconnection attempts
	for i := 0; i < m.config.MaxReconnectAttempts; i++ {
		// Check if the client is already back online
		updatedClient, err := m.clientManager.GetClient(c.ID)
		if err != nil {
			m.logger.Error("Failed to get client", map[string]interface{}{
				"client_id": c.ID,
				"error":     err.Error(),
			})
			return
		}

		if updatedClient.Status != client.StatusError {
			m.logger.Info("Client is back online", map[string]interface{}{
				"client_id":   c.ID,
				"client_name": c.Name,
				"client_status": updatedClient.Status,
			})
			return
		}

		// Wait before next attempt
		time.Sleep(m.config.ReconnectInterval)

		// Simulate a reconnection attempt
		// In a real implementation, this would attempt to establish a new connection
		// For now, we'll just log the attempt and randomly succeed or fail
		m.logger.Info("Reconnection attempt", map[string]interface{}{
			"client_id":   c.ID,
			"client_name": c.Name,
			"attempt":     i + 1,
			"max_attempts": m.config.MaxReconnectAttempts,
		})

		// For demonstration purposes, let's say the last attempt always succeeds
		if i == m.config.MaxReconnectAttempts-1 {
			err := m.clientManager.UpdateClientStatus(c.ID, client.StatusOnline, "")
			if err != nil {
				m.logger.Error("Failed to update client status", map[string]interface{}{
					"client_id": c.ID,
					"error":     err.Error(),
				})
				return
			}

			m.logger.Info("Reconnection successful", map[string]interface{}{
				"client_id":   c.ID,
				"client_name": c.Name,
			})
			return
		}
	}

	m.logger.Error("Reconnection failed after maximum attempts", map[string]interface{}{
		"client_id":   c.ID,
		"client_name": c.Name,
		"max_attempts": m.config.MaxReconnectAttempts,
	})
}
