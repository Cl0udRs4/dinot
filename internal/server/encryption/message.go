package encryption

import (
    "errors"

    clientenc "github.com/Cl0udRs4/dinot/internal/client/encryption"
)

var (
    // ErrInvalidMessageFormat is returned when a message has an invalid format
    ErrInvalidMessageFormat = errors.New("invalid message format")
)

// MessageProcessor processes encrypted messages
type MessageProcessor struct {
    keyExchangeHandler *KeyExchangeHandler
}

// NewMessageProcessor creates a new message processor
func NewMessageProcessor() *MessageProcessor {
    return &MessageProcessor{
        keyExchangeHandler: NewKeyExchangeHandler(),
    }
}

// ProcessIncomingMessage processes an incoming message from a client
func (p *MessageProcessor) ProcessIncomingMessage(clientID string, data []byte) ([]byte, error) {
    // Try to parse as a key exchange message
    var keyExchangeMsg clientenc.KeyExchangeMessage
    err := keyExchangeMsg.FromJSON(data)
    if err == nil && keyExchangeMsg.Type == "key_exchange" {
        // Handle key exchange
        return p.keyExchangeHandler.HandleKeyExchange(clientID, data)
    }
    
    // Get client encryption state
    clientEnc, err := p.keyExchangeHandler.GetClientEncryption(clientID)
    if err != nil {
        return nil, err
    }
    
    // If encryption is not enabled, return the data as is
    if clientEnc.GetEncryptionType() == EncryptionNone {
        return data, nil
    }
    
    // Try to parse as an encrypted message
    var message clientenc.Message
    err = message.FromJSON(data)
    if err != nil {
        return nil, ErrInvalidMessageFormat
    }
    
    // Check if the encryption type matches
    if message.Header.Encryption != clientEnc.GetEncryptionType() {
        return nil, ErrUnsupportedEncryption
    }
    
    // Decrypt the payload
    decryptedData, err := clientEnc.Decrypt(message.Payload)
    if err != nil {
        return nil, err
    }
    
    return decryptedData, nil
}

// ProcessOutgoingMessage processes an outgoing message to a client
func (p *MessageProcessor) ProcessOutgoingMessage(clientID string, data []byte) ([]byte, error) {
    // Get client encryption state
    clientEnc, err := p.keyExchangeHandler.GetClientEncryption(clientID)
    if err != nil {
        return nil, err
    }
    
    // If encryption is not enabled, return the data as is
    if clientEnc.GetEncryptionType() == EncryptionNone {
        return data, nil
    }
    
    // Encrypt the data
    encryptedData, err := clientEnc.Encrypt(data)
    if err != nil {
        return nil, err
    }
    
    // Create a message with the encrypted payload
    message := clientenc.NewMessage(
        clientEnc.GetEncryptionType(),
        clientEnc.GetEncrypter().GetKeyID(),
        encryptedData,
    )
    
    // Convert to JSON
    messageData, err := message.ToJSON()
    if err != nil {
        return nil, err
    }
    
    return messageData, nil
}

// RegisterClient registers a client for encryption
func (p *MessageProcessor) RegisterClient(clientID string) *ClientEncryption {
    return p.keyExchangeHandler.RegisterClient(clientID)
}

// GetClientEncryption gets the encryption state for a client
func (p *MessageProcessor) GetClientEncryption(clientID string) (*ClientEncryption, error) {
    return p.keyExchangeHandler.GetClientEncryption(clientID)
}
