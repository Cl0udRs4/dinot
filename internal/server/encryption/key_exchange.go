package encryption

import (
    "encoding/json"
    "errors"
    "time"

    clientenc "github.com/Cl0udRs4/dinot/internal/client/encryption"
)

// KeyExchangeHandler handles key exchange with clients
type KeyExchangeHandler struct {
    clients map[string]*ClientEncryption
}

// NewKeyExchangeHandler creates a new key exchange handler
func NewKeyExchangeHandler() *KeyExchangeHandler {
    return &KeyExchangeHandler{
        clients: make(map[string]*ClientEncryption),
    }
}

// RegisterClient registers a client for key exchange
func (h *KeyExchangeHandler) RegisterClient(clientID string) *ClientEncryption {
    clientEnc := NewClientEncryption(clientID)
    h.clients[clientID] = clientEnc
    return clientEnc
}

// GetClientEncryption gets the encryption state for a client
func (h *KeyExchangeHandler) GetClientEncryption(clientID string) (*ClientEncryption, error) {
    clientEnc, ok := h.clients[clientID]
    if !ok {
        return nil, errors.New("client not registered for encryption")
    }
    return clientEnc, nil
}

// HandleKeyExchange handles a key exchange message from a client
func (h *KeyExchangeHandler) HandleKeyExchange(clientID string, data []byte) ([]byte, error) {
    // Get client encryption state
    clientEnc, err := h.GetClientEncryption(clientID)
    if err != nil {
        return nil, err
    }
    
    // Parse the key exchange message
    var keyExchangeMsg clientenc.KeyExchangeMessage
    err = keyExchangeMsg.FromJSON(data)
    if err != nil {
        return nil, err
    }
    
    // Set the encryption type
    encType := EncryptionType(keyExchangeMsg.EncryptionType)
    err = clientEnc.SetEncryptionType(encType)
    if err != nil {
        return nil, err
    }
    
    // Generate a new key pair
    publicKey, err := clientEnc.KeyExchanger.GenerateKeyPair()
    if err != nil {
        return nil, err
    }
    
    // Compute the shared secret
    sharedSecret, err := clientEnc.KeyExchanger.ComputeSharedSecret(keyExchangeMsg.PublicKey)
    if err != nil {
        return nil, err
    }
    
    // Create the appropriate encrypter based on the encryption type
    var encrypter clientenc.Encrypter
    switch encType {
    case EncryptionAES:
        encrypter, err = clientenc.NewAESEncrypterWithKey(sharedSecret)
        if err != nil {
            return nil, err
        }
    case EncryptionChaCha20:
        encrypter, err = clientenc.NewChaCha20EncrypterWithKey(sharedSecret)
        if err != nil {
            return nil, err
        }
    default:
        return nil, ErrUnsupportedEncryption
    }
    
    // Set the encrypter
    err = clientEnc.SetEncrypter(encrypter)
    if err != nil {
        return nil, err
    }
    
    // Create a response message
    responseMsg := clientenc.NewKeyExchangeMessage(
        encType,
        publicKey,
        time.Now().Add(24*time.Hour).Unix(),
    )
    
    // Convert to JSON
    responseData, err := responseMsg.ToJSON()
    if err != nil {
        return nil, err
    }
    
    return responseData, nil
}
