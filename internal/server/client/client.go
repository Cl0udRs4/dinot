package client

import (
	"encoding/json"
	"sync"
	"time"
)

// ClientStatus represents the current status of a client
type ClientStatus string

const (
	// StatusOnline indicates the client is online and responsive
	StatusOnline ClientStatus = "online"
	// StatusOffline indicates the client is offline or unresponsive
	StatusOffline ClientStatus = "offline"
	// StatusBusy indicates the client is online but currently executing a task
	StatusBusy ClientStatus = "busy"
	// StatusError indicates the client is experiencing an error
	StatusError ClientStatus = "error"
)

// Client represents a connected client in the C2 system
type Client struct {
	// ID is the unique identifier for the client
	ID string `json:"id"`
	
	// Name is a human-readable name for the client
	Name string `json:"name"`
	
	// IPAddress is the client's IP address
	IPAddress string `json:"ip_address"`
	
	// OS is the operating system of the client
	OS string `json:"os"`
	
	// Architecture is the CPU architecture of the client
	Architecture string `json:"architecture"`
	
	// RegisteredAt is when the client first registered with the server
	RegisteredAt time.Time `json:"registered_at"`
	
	// LastSeen is when the client was last seen (heartbeat or communication)
	LastSeen time.Time `json:"last_seen"`
	
	// Status is the current status of the client
	Status ClientStatus `json:"status"`
	
	// SupportedModules is a list of modules supported by this client
	SupportedModules []string `json:"supported_modules"`
	
	// ActiveModules is a list of currently active modules on this client
	ActiveModules []string `json:"active_modules"`
	
	// Protocol is the communication protocol being used by this client
	Protocol string `json:"protocol"`
	
	// HeartbeatInterval is the interval at which this client sends heartbeats
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
	
	// ErrorMessage contains the last error message if Status is StatusError
	ErrorMessage string `json:"error_message,omitempty"`
	
	// mu protects concurrent access to the client data
	mu sync.RWMutex
}

// NewClient creates a new client with the given ID and initial data
func NewClient(id, name, ipAddress, os, arch string, supportedModules []string, protocol string) *Client {
	now := time.Now()
	return &Client{
		ID:                id,
		Name:              name,
		IPAddress:         ipAddress,
		OS:                os,
		Architecture:      arch,
		RegisteredAt:      now,
		LastSeen:          now,
		Status:            StatusOnline,
		SupportedModules:  supportedModules,
		ActiveModules:     []string{},
		Protocol:          protocol,
		HeartbeatInterval: 60 * time.Second, // Default heartbeat interval
	}
}

// UpdateStatus updates the client's status and last seen time
func (c *Client) UpdateStatus(status ClientStatus, errorMsg string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.Status = status
	c.LastSeen = time.Now()
	
	if status == StatusError {
		c.ErrorMessage = errorMsg
	} else {
		c.ErrorMessage = ""
	}
}

// UpdateLastSeen updates the client's last seen time
func (c *Client) UpdateLastSeen() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.LastSeen = time.Now()
}

// SetHeartbeatInterval sets the client's heartbeat interval
func (c *Client) SetHeartbeatInterval(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.HeartbeatInterval = interval
}

// AddActiveModule adds a module to the list of active modules
func (c *Client) AddActiveModule(module string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Check if the module is supported
	supported := false
	for _, m := range c.SupportedModules {
		if m == module {
			supported = true
			break
		}
	}
	
	if !supported {
		return false
	}
	
	// Check if the module is already active
	for _, m := range c.ActiveModules {
		if m == module {
			return true // Already active
		}
	}
	
	c.ActiveModules = append(c.ActiveModules, module)
	return true
}

// RemoveActiveModule removes a module from the list of active modules
func (c *Client) RemoveActiveModule(module string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for i, m := range c.ActiveModules {
		if m == module {
			// Remove the module by replacing it with the last element and truncating
			c.ActiveModules[i] = c.ActiveModules[len(c.ActiveModules)-1]
			c.ActiveModules = c.ActiveModules[:len(c.ActiveModules)-1]
			return true
		}
	}
	
	return false // Module was not active
}

// IsModuleActive checks if a module is active
func (c *Client) IsModuleActive(module string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	for _, m := range c.ActiveModules {
		if m == module {
			return true
		}
	}
	
	return false
}

// ToJSON converts the client to a JSON string
func (c *Client) ToJSON() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return json.Marshal(c)
}

// FromJSON updates the client from a JSON string
func (c *Client) FromJSON(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	return json.Unmarshal(data, c)
}
