package protocol

import (
    "context"
    "encoding/json"
    "sync"
    "time"

    "github.com/Cl0udRs4/dinot/internal/client/encryption"
)

// ProtocolManager manages multiple protocols and handles automatic switching
type ProtocolManager struct {
    protocols           []Protocol
    currentProtocol     Protocol
    mu                  sync.RWMutex
    switchThreshold     int
    failCount           int
    keyExchanger        encryption.KeyExchanger
    encryptionType      encryption.EncryptionType
    keyRotationInterval time.Duration
    lastKeyRotation     time.Time
}

// NewProtocolManager creates a new protocol manager
func NewProtocolManager(protocols []Protocol, switchThreshold int) *ProtocolManager {
    var currentProtocol Protocol
    if len(protocols) > 0 {
        currentProtocol = protocols[0]
    }
    
    keyExchanger, _ := encryption.NewECDHKeyExchanger()
    
    return &ProtocolManager{
        protocols:           protocols,
        currentProtocol:     currentProtocol,
        switchThreshold:     switchThreshold,
        failCount:           0,
        keyExchanger:        keyExchanger,
        encryptionType:      encryption.EncryptionNone,
        keyRotationInterval: 24 * time.Hour,
        lastKeyRotation:     time.Now(),
    }
}

// SetEncryptionType sets the encryption type for the protocol manager
func (m *ProtocolManager) SetEncryptionType(encType encryption.EncryptionType) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.encryptionType = encType
}

// SetKeyRotationInterval sets the key rotation interval
func (m *ProtocolManager) SetKeyRotationInterval(interval time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.keyRotationInterval = interval
}

// Connect connects using the current protocol
func (m *ProtocolManager) Connect(ctx context.Context) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if m.currentProtocol == nil {
        return ErrNoProtocolAvailable
    }
    
    err := m.currentProtocol.Connect(ctx)
    if err != nil {
        m.failCount++
        if m.failCount >= m.switchThreshold {
            m.switchProtocol()
            m.failCount = 0
        }
        return err
    }
    
    m.failCount = 0
    
    // Perform key exchange if encryption is enabled
    if m.encryptionType != encryption.EncryptionNone {
        err = m.performKeyExchange(ctx)
        if err != nil {
            return err
        }
    }
    
    return nil
}

// performKeyExchange performs key exchange with the server
func (m *ProtocolManager) performKeyExchange(ctx context.Context) error {
    // Generate a new key pair
    publicKey, err := m.keyExchanger.GenerateKeyPair()
    if err != nil {
        return err
    }
    
    // Create a key exchange message
    keyExchangeMsg := encryption.NewKeyExchangeMessage(
        m.encryptionType,
        publicKey,
        time.Now().Add(m.keyRotationInterval).Unix(),
    )
    
    // Convert to JSON
    keyExchangeData, err := keyExchangeMsg.ToJSON()
    if err != nil {
        return err
    }
    
    // Send the key exchange message
    err = m.currentProtocol.Send(keyExchangeData)
    if err != nil {
        return err
    }
    
    // Receive the server's response
    responseData, err := m.currentProtocol.Receive(30 * time.Second)
    if err != nil {
        return err
    }
    
    // Parse the response
    var serverKeyExchangeMsg encryption.KeyExchangeMessage
    err = serverKeyExchangeMsg.FromJSON(responseData)
    if err != nil {
        return err
    }
    
    // Compute the shared secret
    sharedSecret, err := m.keyExchanger.ComputeSharedSecret(serverKeyExchangeMsg.PublicKey)
    if err != nil {
        return err
    }
    
    // Create the appropriate encrypter based on the encryption type
    var encrypter encryption.Encrypter
    switch m.encryptionType {
    case encryption.EncryptionAES:
        encrypter, err = encryption.NewAESEncrypterWithKey(sharedSecret)
        if err != nil {
            return err
        }
    case encryption.EncryptionChaCha20:
        encrypter, err = encryption.NewChaCha20EncrypterWithKey(sharedSecret)
        if err != nil {
            return err
        }
    default:
        return ErrInvalidMessageType
    }
    
    // Set the encrypter for the current protocol
    err = m.currentProtocol.SetEncrypter(encrypter)
    if err != nil {
        return err
    }
    
    // Set the encryption type for the current protocol
    err = m.currentProtocol.SetEncryptionType(m.encryptionType)
    if err != nil {
        return err
    }
    
    // Update last key rotation time
    m.lastKeyRotation = time.Now()
    
    return nil
}

// checkKeyRotation checks if key rotation is needed and performs it if necessary
func (m *ProtocolManager) checkKeyRotation() error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    // Check if key rotation is needed
    if time.Since(m.lastKeyRotation) < m.keyRotationInterval {
        return nil
    }
    
    // Perform key rotation
    if m.currentProtocol == nil || !m.currentProtocol.IsConnected() {
        return ErrNotConnected
    }
    
    encrypter := m.currentProtocol.GetEncrypter()
    if encrypter == nil {
        return ErrInvalidMessageType
    }
    
    // Rotate the key
    _, err := encrypter.RotateKey()
    if err != nil {
        return err
    }
    
    // Update last key rotation time
    m.lastKeyRotation = time.Now()
    
    return nil
}

// Disconnect disconnects the current protocol
func (m *ProtocolManager) Disconnect() error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if m.currentProtocol == nil {
        return nil
    }
    
    return m.currentProtocol.Disconnect()
}

// Send sends data using the current protocol
func (m *ProtocolManager) Send(data []byte) error {
    m.mu.RLock()
    protocol := m.currentProtocol
    encryptionType := m.encryptionType
    m.mu.RUnlock()
    
    if protocol == nil {
        return ErrNoProtocolAvailable
    }
    
    // Check if key rotation is needed
    if encryptionType != encryption.EncryptionNone {
        err := m.checkKeyRotation()
        if err != nil {
            return err
        }
    }
    
    // If encryption is enabled, encrypt the data
    if encryptionType != encryption.EncryptionNone {
        encrypter := protocol.GetEncrypter()
        if encrypter == nil {
            return ErrInvalidMessageType
        }
        
        // Encrypt the data
        encryptedData, err := encrypter.Encrypt(data)
        if err != nil {
            return err
        }
        
        // Create a message with the encrypted payload
        message := encryption.NewMessage(encryptionType, encrypter.GetKeyID(), encryptedData)
        
        // Convert to JSON
        messageData, err := message.ToJSON()
        if err != nil {
            return err
        }
        
        data = messageData
    }
    
    err := protocol.Send(data)
    if err != nil {
        m.mu.Lock()
        m.failCount++
        if m.failCount >= m.switchThreshold {
            m.switchProtocol()
            m.failCount = 0
        }
        m.mu.Unlock()
        return err
    }
    
    m.mu.Lock()
    m.failCount = 0
    m.mu.Unlock()
    return nil
}

// Receive receives data using the current protocol
func (m *ProtocolManager) Receive(timeout time.Duration) ([]byte, error) {
    m.mu.RLock()
    protocol := m.currentProtocol
    encryptionType := m.encryptionType
    m.mu.RUnlock()
    
    if protocol == nil {
        return nil, ErrNoProtocolAvailable
    }
    
    data, err := protocol.Receive(timeout)
    if err != nil {
        m.mu.Lock()
        m.failCount++
        if m.failCount >= m.switchThreshold {
            m.switchProtocol()
            m.failCount = 0
        }
        m.mu.Unlock()
        return nil, err
    }
    
    // If encryption is enabled, decrypt the data
    if encryptionType != encryption.EncryptionNone {
        encrypter := protocol.GetEncrypter()
        if encrypter == nil {
            return nil, ErrInvalidMessageType
        }
        
        // Parse the message
        var message encryption.Message
        err = message.FromJSON(data)
        if err != nil {
            return nil, err
        }
        
        // Check if the encryption type matches
        if message.Header.Encryption != encryptionType {
            return nil, ErrInvalidMessageType
        }
        
        // Decrypt the payload
        decryptedData, err := encrypter.Decrypt(message.Payload)
        if err != nil {
            return nil, err
        }
        
        data = decryptedData
    }
    
    m.mu.Lock()
    m.failCount = 0
    m.mu.Unlock()
    return data, nil
}

// GetCurrentProtocol returns the current protocol
func (m *ProtocolManager) GetCurrentProtocol() Protocol {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.currentProtocol
}

// switchProtocol switches to the next available protocol
func (m *ProtocolManager) switchProtocol() {
    if len(m.protocols) <= 1 {
        return
    }
    
    currentIndex := -1
    for i, p := range m.protocols {
        if p == m.currentProtocol {
            currentIndex = i
            break
        }
    }
    
    nextIndex := (currentIndex + 1) % len(m.protocols)
    m.currentProtocol = m.protocols[nextIndex]
}
