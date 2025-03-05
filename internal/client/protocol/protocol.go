package protocol

import (
    "context"
    "time"

    "github.com/Cl0udRs4/dinot/internal/client/encryption"
)

// Protocol defines the interface for client communication protocols
type Protocol interface {
    // Connect establishes a connection to the server
    Connect(ctx context.Context) error
    
    // Disconnect closes the connection to the server
    Disconnect() error
    
    // Send sends data to the server
    Send(data []byte) error
    
    // Receive receives data from the server with timeout
    Receive(timeout time.Duration) ([]byte, error)
    
    // IsConnected returns true if the protocol is connected
    IsConnected() bool
    
    // GetName returns the protocol name
    GetName() string
    
    // GetEncryptionType returns the encryption type
    GetEncryptionType() encryption.EncryptionType
    
    // SetEncryptionType sets the encryption type
    SetEncryptionType(encType encryption.EncryptionType) error
    
    // GetEncrypter returns the current encrypter
    GetEncrypter() encryption.Encrypter
    
    // SetEncrypter sets the encrypter
    SetEncrypter(encrypter encryption.Encrypter) error
}

// BaseProtocol provides common functionality for all protocols
type BaseProtocol struct {
    Name           string
    Connected      bool
    Timeout        time.Duration
    EncryptionType encryption.EncryptionType
    Encrypter      encryption.Encrypter
}

// GetName returns the protocol name
func (p *BaseProtocol) GetName() string {
    return p.Name
}

// IsConnected returns true if the protocol is connected
func (p *BaseProtocol) IsConnected() bool {
    return p.Connected
}

// GetEncryptionType returns the encryption type
func (p *BaseProtocol) GetEncryptionType() encryption.EncryptionType {
    return p.EncryptionType
}

// SetEncryptionType sets the encryption type
func (p *BaseProtocol) SetEncryptionType(encType encryption.EncryptionType) error {
    p.EncryptionType = encType
    return nil
}

// GetEncrypter returns the current encrypter
func (p *BaseProtocol) GetEncrypter() encryption.Encrypter {
    return p.Encrypter
}

// SetEncrypter sets the encrypter
func (p *BaseProtocol) SetEncrypter(encrypter encryption.Encrypter) error {
    p.Encrypter = encrypter
    return nil
}
