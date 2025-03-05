package client

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrClientNotFound is returned when a client with the specified ID is not found
	ErrClientNotFound = errors.New("client not found")
	
	// ErrClientAlreadyExists is returned when trying to register a client with an ID that already exists
	ErrClientAlreadyExists = errors.New("client already exists")
)

// ClientManager manages all connected clients
type ClientManager struct {
	// clients maps client IDs to Client objects
	clients map[string]*Client
	
	// exceptionManager manages exception reports
	exceptionManager *ExceptionManager
	
	// mu protects concurrent access to the clients map
	mu sync.RWMutex
}

// NewClientManager creates a new client manager
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:          make(map[string]*Client),
		exceptionManager: NewExceptionManager(),
	}
}

// RegisterClient registers a new client with the manager
func (m *ClientManager) RegisterClient(client *Client) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if a client with this ID already exists
	if _, exists := m.clients[client.ID]; exists {
		return ErrClientAlreadyExists
	}
	
	// Add the client to the map
	m.clients[client.ID] = client
	return nil
}

// UnregisterClient removes a client from the manager
func (m *ClientManager) UnregisterClient(clientID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if the client exists
	if _, exists := m.clients[clientID]; !exists {
		return ErrClientNotFound
	}
	
	// Remove the client from the map
	delete(m.clients, clientID)
	return nil
}

// GetClient retrieves a client by ID
func (m *ClientManager) GetClient(clientID string) (*Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	client, exists := m.clients[clientID]
	if !exists {
		return nil, ErrClientNotFound
	}
	
	return client, nil
}

// GetAllClients returns a slice of all registered clients
func (m *ClientManager) GetAllClients() []*Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	clients := make([]*Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	
	return clients
}

// GetClientsByStatus returns a slice of clients with the specified status
func (m *ClientManager) GetClientsByStatus(status ClientStatus) []*Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	clients := make([]*Client, 0)
	for _, client := range m.clients {
		if client.Status == status {
			clients = append(clients, client)
		}
	}
	
	return clients
}

// UpdateClientStatus updates the status of a client
func (m *ClientManager) UpdateClientStatus(clientID string, status ClientStatus, errorMsg string) error {
	client, err := m.GetClient(clientID)
	if err != nil {
		return err
	}
	
	client.UpdateStatus(status, errorMsg)
	return nil
}

// UpdateClientLastSeen updates the last seen time of a client
func (m *ClientManager) UpdateClientLastSeen(clientID string) error {
	client, err := m.GetClient(clientID)
	if err != nil {
		return err
	}
	
	client.UpdateLastSeen()
	return nil
}

// CheckOfflineClients checks for clients that haven't sent a heartbeat in a while
// and marks them as offline
func (m *ClientManager) CheckOfflineClients(timeout time.Duration) []*Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	now := time.Now()
	offlineClients := make([]*Client, 0)
	
	for _, client := range m.clients {
		// Skip already offline clients
		if client.Status == StatusOffline {
			continue
		}
		
		// Check if the client has exceeded its heartbeat interval plus timeout
		if now.Sub(client.LastSeen) > (client.HeartbeatInterval + timeout) {
			client.UpdateStatus(StatusOffline, "Heartbeat timeout exceeded")
			offlineClients = append(offlineClients, client)
		}
	}
	
	return offlineClients
}

// Count returns the total number of registered clients
func (m *ClientManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return len(m.clients)
}

// CountByStatus returns the number of clients with the specified status
func (m *ClientManager) CountByStatus(status ClientStatus) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	count := 0
	for _, client := range m.clients {
		if client.Status == status {
			count++
		}
	}
	
	return count
}

// ReportException reports an exception for a client
func (m *ClientManager) ReportException(clientID, message string, severity ExceptionSeverity, module, stackTrace string, additionalInfo map[string]string) (*ExceptionReport, error) {
	// Check if the client exists
	client, err := m.GetClient(clientID)
	if err != nil {
		return nil, err
	}
	
	// Create a new exception report
	report := &ExceptionReport{
		ID:             fmt.Sprintf("%s-%d", clientID, time.Now().UnixNano()),
		ClientID:       clientID,
		Timestamp:      time.Now(),
		Message:        message,
		Severity:       severity,
		Module:         module,
		StackTrace:     stackTrace,
		AdditionalInfo: additionalInfo,
	}
	
	// Add the report to the exception manager
	m.exceptionManager.AddReport(report)
	
	// Update the client's status if the severity is high enough
	if severity == SeverityError || severity == SeverityCritical {
		client.UpdateStatus(StatusError, message)
	}
	
	return report, nil
}

// GetExceptionReports retrieves all exception reports for a client
func (m *ClientManager) GetExceptionReports(clientID string) ([]*ExceptionReport, error) {
	// Check if the client exists
	_, err := m.GetClient(clientID)
	if err != nil {
		return nil, err
	}
	
	// Get the client's exception reports
	return m.exceptionManager.GetClientReports(clientID), nil
}

// GetAllExceptionReports retrieves all exception reports
func (m *ClientManager) GetAllExceptionReports() []*ExceptionReport {
	return m.exceptionManager.GetAllReports()
}
