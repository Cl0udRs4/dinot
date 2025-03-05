package encryption

import (
    "crypto/rand"
    "errors"
    "io"
    "sync"

    clientenc "github.com/Cl0udRs4/dinot/internal/client/encryption"
)

// EncryptionType represents the type of encryption
type EncryptionType = clientenc.EncryptionType

const (
    // EncryptionNone represents no encryption
    EncryptionNone = clientenc.EncryptionNone
    // EncryptionAES represents AES encryption
    EncryptionAES = clientenc.EncryptionAES
    // EncryptionChaCha20 represents ChaCha20 encryption
    EncryptionChaCha20 = clientenc.EncryptionChaCha20
)

var (
    // ErrInvalidKey is returned when an invalid key is provided
    ErrInvalidKey = errors.New("invalid encryption key")
    // ErrInvalidData is returned when invalid data is provided
    ErrInvalidData = errors.New("invalid data for encryption/decryption")
    // ErrInvalidNonce is returned when an invalid nonce is provided
    ErrInvalidNonce = errors.New("invalid nonce")
    // ErrUnsupportedEncryption is returned when an unsupported encryption type is requested
    ErrUnsupportedEncryption = errors.New("unsupported encryption type")
)

// ClientEncryption represents the encryption state for a client
type ClientEncryption struct {
    ClientID      string
    EncryptionType EncryptionType
    Encrypter     clientenc.Encrypter
    KeyExchanger  clientenc.KeyExchanger
    mu            sync.RWMutex
}

// NewClientEncryption creates a new client encryption state
func NewClientEncryption(clientID string) *ClientEncryption {
    keyExchanger, _ := clientenc.NewECDHKeyExchanger()
    
    return &ClientEncryption{
        ClientID:      clientID,
        EncryptionType: EncryptionNone,
        KeyExchanger:  keyExchanger,
    }
}

// SetEncryptionType sets the encryption type
func (c *ClientEncryption) SetEncryptionType(encType EncryptionType) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.EncryptionType = encType
    return nil
}

// GetEncryptionType returns the encryption type
func (c *ClientEncryption) GetEncryptionType() EncryptionType {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    return c.EncryptionType
}

// SetEncrypter sets the encrypter
func (c *ClientEncryption) SetEncrypter(encrypter clientenc.Encrypter) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.Encrypter = encrypter
    return nil
}

// GetEncrypter returns the encrypter
func (c *ClientEncryption) GetEncrypter() clientenc.Encrypter {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    return c.Encrypter
}

// Encrypt encrypts the plaintext
func (c *ClientEncryption) Encrypt(plaintext []byte) ([]byte, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if c.EncryptionType == EncryptionNone || c.Encrypter == nil {
        return plaintext, nil
    }
    
    return c.Encrypter.Encrypt(plaintext)
}

// Decrypt decrypts the ciphertext
func (c *ClientEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if c.EncryptionType == EncryptionNone || c.Encrypter == nil {
        return ciphertext, nil
    }
    
    return c.Encrypter.Decrypt(ciphertext)
}

// GenerateRandomBytes generates random bytes of the specified length
func GenerateRandomBytes(length int) ([]byte, error) {
    bytes := make([]byte, length)
    _, err := io.ReadFull(rand.Reader, bytes)
    if err != nil {
        return nil, err
    }
    return bytes, nil
}
