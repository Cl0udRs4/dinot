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
    messageProcessor *encryption.MessageProcessor
    clientManager    *client.Manager
    logger           *logging.Logger
    mu               sync.RWMutex
}

// NewEncryptedListener creates a new encrypted listener
func NewEncryptedListener(baseListener Listener, clientManager *client.Manager, logger *logging.Logger) *EncryptedListener {
    return &EncryptedListener{
        baseListener:     baseListener,
        messageProcessor: encryption.NewMessageProcessor(),
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
    
    // Register the client for encryption
    clientEnc := l.messageProcessor.RegisterClient(clientID)
    
    // Register the client with the client manager
    c := client.NewClient(clientID, conn.RemoteAddr().String(), l.GetProtocol())
    l.clientManager.RegisterClient(c)
    
    l.logger.Info("New encrypted connection", map[string]interface{}{
        "client_id":  clientID,
        "remote_addr": conn.RemoteAddr().String(),
        "protocol":    l.GetProtocol(),
    })
    
    // Handle the connection
    go l.handleClient(conn, clientID)
}

// handleClient handles a client connection with encryption support
func (l *EncryptedListener) handleClient(conn net.Conn, clientID string) {
    defer func() {
        conn.Close()
        l.clientManager.UnregisterClient(clientID)
        l.logger.Info("Client disconnected", map[string]interface{}{
            "client_id": clientID,
        })
    }()
    
    buffer := make([]byte, 4096)
    
    for {
        // Set read deadline
        conn.SetReadDeadline(time.Now().Add(30 * time.Second))
        
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
        processedData, err := l.messageProcessor.ProcessIncomingMessage(clientID, data)
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
        encryptedResponse, err := l.messageProcessor.ProcessOutgoingMessage(clientID, response)
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
    clientEnc, err := l.messageProcessor.GetClientEncryption(clientID)
    if err == nil {
        encType := clientEnc.GetEncryptionType()
        l.logger.Info("Client encryption", map[string]interface{}{
            "client_id":      clientID,
            "encryption_type": encType,
        })
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
    clientEnc, err := l.messageProcessor.GetClientEncryption(clientID)
    if err == nil {
        encType = string(clientEnc.GetEncryptionType())
    }
    
    // Create status response
    status := struct {
        Status      string   `json:"status"`
        ClientID    string   `json:"client_id"`
        Hostname    string   `json:"hostname"`
        OS          string   `json:"os"`
        Arch        string   `json:"arch"`
        Modules     []string `json:"modules"`
        Protocols   []string `json:"protocols"`
        Encryption  string   `json:"encryption"`
        LastActivity string   `json:"last_activity"`
    }{
        Status:      "success",
        ClientID:    c.ID,
        Hostname:    c.Hostname,
        OS:          c.OS,
        Arch:        c.Arch,
        Modules:     c.Modules,
        Protocols:   c.Protocols,
        Encryption:  encType,
        LastActivity: c.LastActivity.Format(time.RFC3339),
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
