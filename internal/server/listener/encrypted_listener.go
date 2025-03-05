package listener

import (
    "context"
    "encoding/json"
    "fmt"
    "net"
    "sync"
    "time"

    "github.com/Cl0udRs4/dinot/internal/server/client"
    "github.com/Cl0udRs4/dinot/internal/server/encryption"
    "github.com/Cl0udRs4/dinot/internal/server/logging"
)

// EncryptedListener wraps a Listener with encryption support
type EncryptedListener struct {
    baseListener     Listener
    securityManager  *encryption.SecurityManager
    clientManager    *client.Manager
    logger           *logging.Logger
    mu               sync.RWMutex
}

// NewEncryptedListener creates a new encrypted listener
func NewEncryptedListener(baseListener Listener, clientManager *client.Manager, logger *logging.Logger) *EncryptedListener {
    // Create security manager with default configuration
    securityConfig := encryption.DefaultSecurityConfig()
    securityManager, err := encryption.NewSecurityManager(securityConfig)
    if err != nil {
        logger.Error("Failed to create security manager", map[string]interface{}{
            "error": err.Error(),
        })
        // Fall back to a basic message processor if security manager creation fails
        return &EncryptedListener{
            baseListener:     baseListener,
            securityManager:  nil,
            clientManager:    clientManager,
            logger:           logger,
        }
    }
    
    // Start the security manager
    if err := securityManager.Start(); err != nil {
        logger.Error("Failed to start security manager", map[string]interface{}{
            "error": err.Error(),
        })
    }
    
    return &EncryptedListener{
        baseListener:     baseListener,
        securityManager:  securityManager,
        clientManager:    clientManager,
        logger:           logger,
    }
}

// Start starts the encrypted listener
func (l *EncryptedListener) Start(ctx context.Context) error {
    return l.baseListener.Start(ctx)
}

// Stop stops the encrypted listener
func (l *EncryptedListener) Stop() error {
    return l.baseListener.Stop()
}

// GetProtocol returns the protocol of the base listener
func (l *EncryptedListener) GetProtocol() string {
    return l.baseListener.GetProtocol()
}

// GetAddress returns the address of the base listener
func (l *EncryptedListener) GetAddress() string {
    return l.baseListener.GetAddress()
}

// HandleConnection handles a new connection with encryption support
func (l *EncryptedListener) HandleConnection(conn net.Conn) {
    // Register the client
    clientID := fmt.Sprintf("%s-%d", conn.RemoteAddr().String(), time.Now().UnixNano())
    
    // Register the client with the security manager
    var clientEnc *encryption.ClientEncryption
    if l.securityManager != nil {
        clientEnc = l.securityManager.RegisterClient(clientID)
    }
    
    // Register the client with the client manager
    c := client.NewClient(clientID, conn.RemoteAddr().String(), l.GetProtocol())
    l.clientManager.RegisterClient(c)
    
    l.logger.Info("New encrypted connection", map[string]interface{}{
        "client_id":  clientID,
        "remote_addr": conn.RemoteAddr().String(),
        "protocol":    l.GetProtocol(),
        "encryption":  clientEnc != nil,
    })
    
    // Handle the connection
    go l.handleClient(conn, clientID)
}

// handleClient handles a client connection with encryption support
func (l *EncryptedListener) handleClient(conn net.Conn, clientID string) {
    defer func() {
        conn.Close()
        // Unregister client from both client manager and security manager
        l.clientManager.UnregisterClient(clientID)
        if l.securityManager != nil {
            l.securityManager.UnregisterClient(clientID)
        }
        l.logger.Info("Client disconnected", map[string]interface{}{
            "client_id": clientID,
        })
    }()
    
    buffer := make([]byte, 4096)
    
    for {
        // Apply jitter to read deadline if security manager is available
        readDeadline := 30 * time.Second
        if l.securityManager != nil {
            jitter := l.securityManager.ApplyJitter()
            if jitter > 0 {
                readDeadline += jitter
            }
        }
        
        // Set read deadline
        conn.SetReadDeadline(time.Now().Add(readDeadline))
        
        // Read data from the connection
        n, err := conn.Read(buffer)
        if err != nil {
            l.logger.Error("Error reading from connection", map[string]interface{}{
                "client_id": clientID,
                "error":     err.Error(),
            })
            break
        }
        
        // Process the incoming message
        data := buffer[:n]
        var processedData []byte
        
        if l.securityManager != nil {
            // Use security manager for processing if available
            processedData, err = l.securityManager.ProcessIncomingMessage(clientID, data)
        } else {
            // Fall back to message processor if security manager is not available
            messageProcessor := encryption.NewMessageProcessor()
            processedData, err = messageProcessor.ProcessIncomingMessage(clientID, data)
        }
        
        if err != nil {
            l.logger.Error("Error processing incoming message", map[string]interface{}{
                "client_id": clientID,
                "error":     err.Error(),
            })
            continue
        }
        
        // Handle the processed message
        response, err := l.handleMessage(clientID, processedData)
        if err != nil {
            l.logger.Error("Error handling message", map[string]interface{}{
                "client_id": clientID,
                "error":     err.Error(),
            })
            continue
        }
        
        // Process the outgoing message
        var encryptedResponse []byte
        
        if l.securityManager != nil {
            // Use security manager for processing if available
            encryptedResponse, err = l.securityManager.ProcessOutgoingMessage(clientID, response)
        } else {
            // Fall back to message processor if security manager is not available
            messageProcessor := encryption.NewMessageProcessor()
            encryptedResponse, err = messageProcessor.ProcessOutgoingMessage(clientID, response)
        }
        
        if err != nil {
            l.logger.Error("Error processing outgoing message", map[string]interface{}{
                "client_id": clientID,
                "error":     err.Error(),
            })
            continue
        }
        
        // Send the response
        _, err = conn.Write(encryptedResponse)
        if err != nil {
            l.logger.Error("Error writing to connection", map[string]interface{}{
                "client_id": clientID,
                "error":     err.Error(),
            })
            break
        }
    }
}

// handleMessage handles a processed message
func (l *EncryptedListener) handleMessage(clientID string, data []byte) ([]byte, error) {
    // Try to parse as a command message
    var message struct {
        Type    string          `json:"type"`
        Command string          `json:"command"`
        Params  json.RawMessage `json:"params"`
    }
    
    err := json.Unmarshal(data, &message)
    if err != nil {
        return nil, err
    }
    
    // Handle different message types
    switch message.Type {
    case "command":
        return l.handleCommandMessage(clientID, message.Command, message.Params)
    case "heartbeat":
        return l.handleHeartbeatMessage(clientID)
    default:
        return []byte(`{"status":"error","message":"unknown message type"}`), nil
    }
}

// handleCommandMessage handles a command message
func (l *EncryptedListener) handleCommandMessage(clientID, command string, params json.RawMessage) ([]byte, error) {
    // Get the client
    c, err := l.clientManager.GetClient(clientID)
    if err != nil {
        return nil, err
    }
    
    // Update the client's last activity time
    c.UpdateLastActivity()
    
    // Handle different commands
    switch command {
    case "register":
        return l.handleRegisterCommand(clientID, params)
    case "status":
        return l.handleStatusCommand(clientID)
    default:
        return []byte(`{"status":"error","message":"unknown command"}`), nil
    }
}

// handleRegisterCommand handles a register command
func (l *EncryptedListener) handleRegisterCommand(clientID string, params json.RawMessage) ([]byte, error) {
    var registerParams struct {
        Hostname  string   `json:"hostname"`
        OS        string   `json:"os"`
        Arch      string   `json:"arch"`
        Modules   []string `json:"modules"`
        Protocols []string `json:"protocols"`
    }
    
    err := json.Unmarshal(params, &registerParams)
    if err != nil {
        return nil, err
    }
    
    // Get the client
    c, err := l.clientManager.GetClient(clientID)
    if err != nil {
        return nil, err
    }
    
    // Update client information
    c.SetHostname(registerParams.Hostname)
    c.SetOS(registerParams.OS)
    c.SetArch(registerParams.Arch)
    c.SetModules(registerParams.Modules)
    c.SetProtocols(registerParams.Protocols)
    
    l.logger.Info("Client registered", map[string]interface{}{
        "client_id": clientID,
        "hostname":  registerParams.Hostname,
        "os":        registerParams.OS,
        "arch":      registerParams.Arch,
        "modules":   registerParams.Modules,
        "protocols": registerParams.Protocols,
    })
    
    // Get client encryption info
    var encType string = "none"
    
    if l.securityManager != nil {
        // Get encryption info from security manager
        messageProcessor := l.securityManager.GetMessageProcessor()
        if messageProcessor != nil {
            clientEnc, err := messageProcessor.GetClientEncryption(clientID)
            if err == nil && clientEnc != nil {
                encType = string(clientEnc.GetEncryptionType())
                l.logger.Info("Client encryption", map[string]interface{}{
                    "client_id":       clientID,
                    "encryption_type": encType,
                })
            }
        }
        
        // Generate authentication token for the client if JWT is enabled
        authenticator := l.securityManager.GetAuthenticator()
        if authenticator != nil {
            token, err := authenticator.GenerateJWT(clientID, "client")
            if err == nil {
                // Include token in response
                response := struct {
                    Status  string `json:"status"`
                    Message string `json:"message"`
                    Token   string `json:"token"`
                }{
                    Status:  "success",
                    Message: "client registered",
                    Token:   token,
                }
                return json.Marshal(response)
            }
        }
    } else {
        // Fall back to message processor if security manager is not available
        messageProcessor := encryption.NewMessageProcessor()
        clientEnc, err := messageProcessor.GetClientEncryption(clientID)
        if err == nil && clientEnc != nil {
            encType = string(clientEnc.GetEncryptionType())
            l.logger.Info("Client encryption", map[string]interface{}{
                "client_id":       clientID,
                "encryption_type": encType,
            })
        }
    }
    
    return []byte(`{"status":"success","message":"client registered"}`), nil
}

// handleStatusCommand handles a status command
func (l *EncryptedListener) handleStatusCommand(clientID string) ([]byte, error) {
    // Get the client
    c, err := l.clientManager.GetClient(clientID)
    if err != nil {
        return nil, err
    }
    
    // Get client encryption info
    encType := "none"
    securityEnabled := false
    
    if l.securityManager != nil {
        securityEnabled = true
        // Get encryption info from security manager
        messageProcessor := l.securityManager.GetMessageProcessor()
        if messageProcessor != nil {
            clientEnc, err := messageProcessor.GetClientEncryption(clientID)
            if err == nil && clientEnc != nil {
                encType = string(clientEnc.GetEncryptionType())
            }
        }
    } else {
        // Fall back to message processor if security manager is not available
        messageProcessor := encryption.NewMessageProcessor()
        clientEnc, err := messageProcessor.GetClientEncryption(clientID)
        if err == nil && clientEnc != nil {
            encType = string(clientEnc.GetEncryptionType())
        }
    }
    
    // Create status response
    status := struct {
        Status         string   `json:"status"`
        ClientID       string   `json:"client_id"`
        Hostname       string   `json:"hostname"`
        OS             string   `json:"os"`
        Arch           string   `json:"arch"`
        Modules        []string `json:"modules"`
        Protocols      []string `json:"protocols"`
        Encryption     string   `json:"encryption"`
        LastActivity   string   `json:"last_activity"`
        SecurityEnabled bool     `json:"security_enabled"`
    }{
        Status:         "success",
        ClientID:       c.ID,
        Hostname:       c.Hostname,
        OS:             c.OS,
        Arch:           c.Arch,
        Modules:        c.Modules,
        Protocols:      c.Protocols,
        Encryption:     encType,
        LastActivity:   c.LastActivity.Format(time.RFC3339),
        SecurityEnabled: securityEnabled,
    }
    
    return json.Marshal(status)
}

// handleHeartbeatMessage handles a heartbeat message
func (l *EncryptedListener) handleHeartbeatMessage(clientID string) ([]byte, error) {
    // Get the client
    c, err := l.clientManager.GetClient(clientID)
    if err != nil {
        return nil, err
    }
    
    // Update the client's last activity time
    c.UpdateLastActivity()
    
    return []byte(`{"status":"success","message":"heartbeat received"}`), nil
}
