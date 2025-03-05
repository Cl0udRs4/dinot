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
    clientManager    *client.ClientManager
    logger           *logging.Logger
    mu               sync.RWMutex
}

// NewEncryptedListener creates a new encrypted listener
func NewEncryptedListener(baseListener Listener, clientManager *client.ClientManager, logger *logging.Logger) *EncryptedListener {
    return &EncryptedListener{
        baseListener:     baseListener,
        messageProcessor: encryption.NewMessageProcessor(),
        clientManager:    clientManager,
        logger:           logger,
    }
}

// Start starts the encrypted listener
func (l *EncryptedListener) Start(ctx context.Context) error {
    return l.baseListener.Start(ctx, l.HandleConnection)
}

// Stop stops the encrypted listener
func (l *EncryptedListener) Stop() error {
    return l.baseListener.Stop()
}

// GetProtocol returns the protocol of the base listener
func (l *EncryptedListener) GetProtocol() string {
    return l.baseListener.GetProtocol()
}

// GetConfig returns the configuration of the base listener
func (l *EncryptedListener) GetConfig() Config {
    return l.baseListener.GetConfig()
}

// GetStatus returns the status of the base listener
func (l *EncryptedListener) GetStatus() Status {
    return l.baseListener.GetStatus()
}

// UpdateConfig updates the configuration of the base listener
func (l *EncryptedListener) UpdateConfig(config Config) error {
    return l.baseListener.UpdateConfig(config)
}

// HandleConnection handles a new connection with encryption support
func (l *EncryptedListener) HandleConnection(conn net.Conn) {
    // Register the client
    clientID := fmt.Sprintf("%s-%d", conn.RemoteAddr().String(), time.Now().UnixNano())
    
    // Register the client for encryption
    clientEnc := l.messageProcessor.RegisterClient(clientID)
    
    // Register the client with the client manager
    c := client.NewClient(clientID, "Client-"+clientID, conn.RemoteAddr().String(), "unknown", "unknown", []string{}, l.GetProtocol())
    l.clientManager.RegisterClient(c)
    
    // Remove the unused variable warning
    _ = clientEnc
    
    // Log new connection
    fmt.Printf("New encrypted connection: client_id=%s, remote_addr=%s, protocol=%s\n", 
        clientID, conn.RemoteAddr().String(), l.GetProtocol())
    
    // Handle the connection
    go l.handleClient(conn, clientID)
}

// handleClient handles a client connection with encryption support
func (l *EncryptedListener) handleClient(conn net.Conn, clientID string) {
    defer func() {
        conn.Close()
        l.clientManager.UnregisterClient(clientID)
        fmt.Printf("Client disconnected: client_id=%s\n", clientID)
    }()
    
    buffer := make([]byte, 4096)
    
    for {
        // Set read deadline
        conn.SetReadDeadline(time.Now().Add(30 * time.Second))
        
        // Read data from the connection
        n, err := conn.Read(buffer)
        if err != nil {
            fmt.Printf("Error reading from connection: client_id=%s, error=%s\n", clientID, err.Error())
            break
        }
        
        // Process the incoming message
        data := buffer[:n]
        processedData, err := l.messageProcessor.ProcessIncomingMessage(clientID, data)
        if err != nil {
            fmt.Printf("Error processing incoming message: client_id=%s, error=%s\n", clientID, err.Error())
            continue
        }
        
        // Handle the processed message
        response, err := l.handleMessage(clientID, processedData)
        if err != nil {
            fmt.Printf("Error handling message: client_id=%s, error=%s\n", clientID, err.Error())
            continue
        }
        
        // Process the outgoing message
        encryptedResponse, err := l.messageProcessor.ProcessOutgoingMessage(clientID, response)
        if err != nil {
            fmt.Printf("Error processing outgoing message: client_id=%s, error=%s\n", clientID, err.Error())
            continue
        }
        
        // Send the response
        _, err = conn.Write(encryptedResponse)
        if err != nil {
            fmt.Printf("Error writing to connection: client_id=%s, error=%s\n", clientID, err.Error())
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
    
    // Update the client's last seen time
    c.UpdateLastSeen()
    
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
    
    // Update client information - directly update fields since setter methods don't exist
    // Note: In a real implementation, we would add these setter methods to the Client struct
    c.Name = registerParams.Hostname
    c.OS = registerParams.OS
    c.Architecture = registerParams.Arch
    c.SupportedModules = registerParams.Modules
    // We don't have a direct field for protocols in the Client struct
    
    fmt.Printf("Client registered: client_id=%s, hostname=%s, os=%s, arch=%s\n", 
        clientID, registerParams.Hostname, registerParams.OS, registerParams.Arch)
    
    // Get client encryption info
    clientEnc, err := l.messageProcessor.GetClientEncryption(clientID)
    if err == nil {
        encType := clientEnc.GetEncryptionType()
        fmt.Printf("Client encryption: client_id=%s, encryption_type=%s\n", 
            clientID, encType)
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
        Name        string   `json:"name"`
        OS          string   `json:"os"`
        Arch        string   `json:"arch"`
        Modules     []string `json:"modules"`
        Protocol    string   `json:"protocol"`
        Encryption  string   `json:"encryption"`
        LastSeen    string   `json:"last_seen"`
    }{
        Status:      "success",
        ClientID:    c.ID,
        Name:        c.Name,
        OS:          c.OS,
        Arch:        c.Architecture,
        Modules:     c.SupportedModules,
        Protocol:    c.Protocol,
        Encryption:  encType,
        LastSeen:    c.LastSeen.Format(time.RFC3339),
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
    
    // Update the client's last seen time
    c.UpdateLastSeen()
    
    return []byte(`{"status":"success","message":"heartbeat received"}`), nil
}
