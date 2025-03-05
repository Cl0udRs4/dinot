package encryption

import (
    "crypto/ecdh"
    "crypto/rand"
    "crypto/sha256"
    "encoding/json"
    "errors"
)

var (
    // ErrInvalidPublicKey is returned when an invalid public key is provided
    ErrInvalidPublicKey = errors.New("invalid public key")
)

// KeyExchanger defines the interface for key exchange
type KeyExchanger interface {
    // GenerateKeyPair generates a new key pair and returns the public key
    GenerateKeyPair() ([]byte, error)
    
    // ComputeSharedSecret computes the shared secret using the peer's public key
    ComputeSharedSecret(peerPublicKey []byte) ([]byte, error)
    
    // GetPublicKey returns the current public key
    GetPublicKey() []byte
}

// ECDHKeyExchanger implements the KeyExchanger interface using ECDH
type ECDHKeyExchanger struct {
    curve      ecdh.Curve
    privateKey *ecdh.PrivateKey
    publicKey  *ecdh.PublicKey
}

// NewECDHKeyExchanger creates a new ECDH key exchanger
func NewECDHKeyExchanger() (*ECDHKeyExchanger, error) {
    curve := ecdh.P256()
    
    privateKey, err := curve.GenerateKey(rand.Reader)
    if err != nil {
        return nil, err
    }
    
    publicKey := privateKey.PublicKey()
    
    return &ECDHKeyExchanger{
        curve:      curve,
        privateKey: privateKey,
        publicKey:  publicKey,
    }, nil
}

// GenerateKeyPair generates a new key pair and returns the public key
func (e *ECDHKeyExchanger) GenerateKeyPair() ([]byte, error) {
    privateKey, err := e.curve.GenerateKey(rand.Reader)
    if err != nil {
        return nil, err
    }
    
    e.privateKey = privateKey
    e.publicKey = privateKey.PublicKey()
    
    return e.publicKey.Bytes(), nil
}

// ComputeSharedSecret computes the shared secret using the peer's public key
func (e *ECDHKeyExchanger) ComputeSharedSecret(peerPublicKeyBytes []byte) ([]byte, error) {
    peerPublicKey, err := e.curve.NewPublicKey(peerPublicKeyBytes)
    if err != nil {
        return nil, ErrInvalidPublicKey
    }
    
    sharedSecret, err := e.privateKey.ECDH(peerPublicKey)
    if err != nil {
        return nil, err
    }
    
    // Hash the shared secret to get a key of appropriate length
    hash := sha256.Sum256(sharedSecret)
    return hash[:], nil
}

// GetPublicKey returns the current public key
func (e *ECDHKeyExchanger) GetPublicKey() []byte {
    return e.publicKey.Bytes()
}

// KeyExchangeMessage represents a key exchange message
type KeyExchangeMessage struct {
    Type            string `json:"type"`
    EncryptionType  string `json:"encryption_type"`
    PublicKey       []byte `json:"public_key"`
    KeyRotationTime int64  `json:"key_rotation_time,omitempty"`
}

// NewKeyExchangeMessage creates a new key exchange message
func NewKeyExchangeMessage(encryptionType EncryptionType, publicKey []byte, keyRotationTime int64) *KeyExchangeMessage {
    return &KeyExchangeMessage{
        Type:            "key_exchange",
        EncryptionType:  string(encryptionType),
        PublicKey:       publicKey,
        KeyRotationTime: keyRotationTime,
    }
}

// ToJSON converts the key exchange message to JSON
func (m *KeyExchangeMessage) ToJSON() ([]byte, error) {
    return json.Marshal(m)
}

// FromJSON updates the key exchange message from JSON
func (m *KeyExchangeMessage) FromJSON(data []byte) error {
    return json.Unmarshal(data, m)
}
