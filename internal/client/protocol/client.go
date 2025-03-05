package protocol

import (
	"context"
	"sync"
	"time"
)

// Client represents a client that can communicate with a server using multiple protocols
type Client struct {
	// switcher is the protocol switcher
	switcher *ProtocolSwitcher
	
	// ctx is the context for the client
	ctx context.Context
	
	// cancel is the function to cancel the client context
	cancel context.CancelFunc
	
	// mu protects concurrent access to the client
	mu sync.RWMutex
	
	// connected indicates if the client is connected
	connected bool
	
	// onConnect is called when the client connects
	onConnect func()
	
	// onDisconnect is called when the client disconnects
	onDisconnect func(err error)
	
	// onProtocolSwitch is called when the client switches protocols
	onProtocolSwitch func(oldProtocol, newProtocol Protocol)
	
	// onError is called when the client encounters an error
	onError func(err error)
}

// ClientConfig defines the configuration for the client
type ClientConfig struct {
	// Protocols is a map of protocol name to protocol configuration
	Protocols map[string]Config
	
	// PrimaryProtocol is the preferred protocol to use
	PrimaryProtocol string
	
	// FallbackOrder is the order to try protocols when the active one fails
	FallbackOrder []string
	
	// SwitchStrategy determines how to select the next protocol
	SwitchStrategy SwitchStrategy
	
	// SwitchThreshold is the number of consecutive failures before switching protocols
	SwitchThreshold int
	
	// MinSwitchInterval is the minimum time between protocol switches in seconds
	MinSwitchInterval int
	
	// TimeoutThreshold is the number of consecutive timeouts before switching protocols
	TimeoutThreshold int
	
	// CheckInterval is the interval between timeout checks in seconds
	CheckInterval int
	
	// MaxInactivity is the maximum allowed inactivity period in seconds
	MaxInactivity int
	
	// JitterMin is the minimum jitter in seconds
	JitterMin int
	
	// JitterMax is the maximum jitter in seconds
	JitterMax int
}

// NewClient creates a new client
func NewClient(config ClientConfig) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create the protocol switcher
	switcher := NewProtocolSwitcher(ProtocolSwitcherConfig{
		SwitchStrategy: config.SwitchStrategy,
		JitterMin:      config.JitterMin,
		JitterMax:      config.JitterMax,
		TimeoutDetectorConfig: TimeoutDetectorConfig{
			TimeoutThreshold: config.TimeoutThreshold,
			CheckInterval:    config.CheckInterval,
			MaxInactivity:    config.MaxInactivity,
		},
		ManagerConfig: ManagerConfig{
			PrimaryProtocol:    config.PrimaryProtocol,
			FallbackOrder:      config.FallbackOrder,
			SwitchThreshold:    config.SwitchThreshold,
			MinSwitchInterval:  config.MinSwitchInterval,
		},
	})
	
	client := &Client{
		switcher:  switcher,
		ctx:       ctx,
		cancel:    cancel,
		connected: false,
	}
	
	// Set up the protocol switch handler
	switcher.SetOnSwitchHandler(client.handleProtocolSwitch)
	
	// Register protocols
	for name, protocolConfig := range config.Protocols {
		var protocol Protocol
		
		switch name {
		case "tcp":
			protocol = NewTCPProtocol(protocolConfig)
		case "udp":
			protocol = NewUDPProtocol(protocolConfig)
		case "ws":
			protocol = NewWSProtocol(protocolConfig)
		case "icmp":
			protocol = NewICMPProtocol(protocolConfig)
		case "dns":
			protocol = NewDNSProtocol(protocolConfig)
		default:
			continue // Skip unknown protocols
		}
		
		err := switcher.RegisterProtocol(protocol)
		if err != nil {
			return nil, err
		}
	}
	
	return client, nil
}

// Connect connects the client to the server
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.connected {
		return nil
	}
	
	// Start the protocol switcher
	err := c.switcher.Start()
	if err != nil {
		return err
	}
	
	// Connect using the active protocol
	err = c.switcher.Connect(c.ctx)
	if err != nil {
		c.switcher.Stop()
		return err
	}
	
	c.connected = true
	
	// Call the connect handler if set
	if c.onConnect != nil {
		go c.onConnect()
	}
	
	return nil
}

// Disconnect disconnects the client from the server
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.connected {
		return nil
	}
	
	// Stop the protocol switcher
	c.switcher.Stop()
	
	// Disconnect using the active protocol
	err := c.switcher.Disconnect()
	
	c.connected = false
	
	// Call the disconnect handler if set
	if c.onDisconnect != nil {
		go c.onDisconnect(err)
	}
	
	return err
}

// Send sends data to the server
func (c *Client) Send(data []byte) (int, error) {
	c.mu.RLock()
	connected := c.connected
	c.mu.RUnlock()
	
	if !connected {
		return 0, ErrNotConnected
	}
	
	n, err := c.switcher.Send(data)
	
	if err != nil && c.onError != nil {
		go c.onError(err)
	}
	
	return n, err
}

// Receive receives data from the server
func (c *Client) Receive() ([]byte, error) {
	c.mu.RLock()
	connected := c.connected
	c.mu.RUnlock()
	
	if !connected {
		return nil, ErrNotConnected
	}
	
	data, err := c.switcher.Receive()
	
	if err != nil && c.onError != nil {
		go c.onError(err)
	}
	
	return data, err
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return c.connected
}

// GetActiveProtocol returns the currently active protocol
func (c *Client) GetActiveProtocol() Protocol {
	return c.switcher.GetActiveProtocol()
}

// SwitchProtocol manually switches to the next protocol
func (c *Client) SwitchProtocol() error {
	c.mu.RLock()
	connected := c.connected
	c.mu.RUnlock()
	
	if !connected {
		return ErrNotConnected
	}
	
	return c.switcher.SwitchToNextProtocol()
}

// SetOnConnectHandler sets the handler to be called when the client connects
func (c *Client) SetOnConnectHandler(handler func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.onConnect = handler
}

// SetOnDisconnectHandler sets the handler to be called when the client disconnects
func (c *Client) SetOnDisconnectHandler(handler func(err error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.onDisconnect = handler
}

// SetOnProtocolSwitchHandler sets the handler to be called when the client switches protocols
func (c *Client) SetOnProtocolSwitchHandler(handler func(oldProtocol, newProtocol Protocol)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.onProtocolSwitch = handler
}

// SetOnErrorHandler sets the handler to be called when the client encounters an error
func (c *Client) SetOnErrorHandler(handler func(err error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.onError = handler
}

// handleProtocolSwitch handles a protocol switch event from the switcher
func (c *Client) handleProtocolSwitch(oldProtocol, newProtocol Protocol) {
	// Call the protocol switch handler if set
	if c.onProtocolSwitch != nil {
		go c.onProtocolSwitch(oldProtocol, newProtocol)
	}
}

// Close closes the client and releases all resources
func (c *Client) Close() error {
	// Disconnect if connected
	if c.IsConnected() {
		c.Disconnect()
	}
	
	// Cancel the context
	c.cancel()
	
	return nil
}
