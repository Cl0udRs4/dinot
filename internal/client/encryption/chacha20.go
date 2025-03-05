package encryption

import (
    "crypto/cipher"
    "crypto/rand"
    "encoding/binary"
    "io"
    "sync"
    "sync/atomic"

    "golang.org/x/crypto/chacha20poly1305"
)

// ChaCha20Encrypter implements the Encrypter interface using ChaCha20-Poly1305
type ChaCha20Encrypter struct {
    key       []byte
    keyID     uint32
    aead      cipher.AEAD
    mu        sync.RWMutex
}

// NewChaCha20Encrypter creates a new ChaCha20 encrypter
func NewChaCha20Encrypter() (*ChaCha20Encrypter, error) {
    key, err := GenerateRandomBytes(chacha20poly1305.KeySize)
    if err != nil {
        return nil, err
    }
    
    aead, err := chacha20poly1305.New(key)
    if err != nil {
        return nil, err
    }
    
    return &ChaCha20Encrypter{
        key:   key,
        keyID: 1,
        aead:  aead,
    }, nil
}

// NewChaCha20EncrypterWithKey creates a new ChaCha20 encrypter with the provided key
func NewChaCha20EncrypterWithKey(key []byte) (*ChaCha20Encrypter, error) {
    if len(key) != chacha20poly1305.KeySize {
        return nil, ErrInvalidKey
    }
    
    aead, err := chacha20poly1305.New(key)
    if err != nil {
        return nil, err
    }
    
    return &ChaCha20Encrypter{
        key:   key,
        keyID: 1,
        aead:  aead,
    }, nil
}

// Encrypt encrypts the plaintext using ChaCha20-Poly1305
func (c *ChaCha20Encrypter) Encrypt(plaintext []byte) ([]byte, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if len(plaintext) == 0 {
        return nil, ErrInvalidData
    }
    
    // Create a nonce
    nonce := make([]byte, c.aead.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    
    // Prepend key ID to the nonce
    keyIDBytes := make([]byte, 4)
    binary.BigEndian.PutUint32(keyIDBytes, c.keyID)
    
    // Encrypt the plaintext
    ciphertext := c.aead.Seal(nil, nonce, plaintext, keyIDBytes)
    
    // Combine key ID, nonce, and ciphertext
    result := make([]byte, 4+len(nonce)+len(ciphertext))
    copy(result[0:4], keyIDBytes)
    copy(result[4:4+len(nonce)], nonce)
    copy(result[4+len(nonce):], ciphertext)
    
    return result, nil
}

// Decrypt decrypts the ciphertext using ChaCha20-Poly1305
func (c *ChaCha20Encrypter) Decrypt(ciphertext []byte) ([]byte, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if len(ciphertext) < 4+c.aead.NonceSize() {
        return nil, ErrInvalidData
    }
    
    // Extract key ID and verify
    keyID := binary.BigEndian.Uint32(ciphertext[0:4])
    if keyID != c.keyID {
        return nil, ErrInvalidKey
    }
    
    // Extract nonce
    nonce := ciphertext[4:4+c.aead.NonceSize()]
    
    // Extract actual ciphertext
    actualCiphertext := ciphertext[4+c.aead.NonceSize():]
    
    // Decrypt the ciphertext
    keyIDBytes := ciphertext[0:4]
    plaintext, err := c.aead.Open(nil, nonce, actualCiphertext, keyIDBytes)
    if err != nil {
        return nil, err
    }
    
    return plaintext, nil
}

// GetType returns the encryption type
func (c *ChaCha20Encrypter) GetType() EncryptionType {
    return EncryptionChaCha20
}

// GetKeyID returns the current key ID
func (c *ChaCha20Encrypter) GetKeyID() uint32 {
    return atomic.LoadUint32(&c.keyID)
}

// RotateKey generates a new key and returns the new key ID
func (c *ChaCha20Encrypter) RotateKey() (uint32, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    key, err := GenerateRandomBytes(chacha20poly1305.KeySize)
    if err != nil {
        return c.keyID, err
    }
    
    aead, err := chacha20poly1305.New(key)
    if err != nil {
        return c.keyID, err
    }
    
    c.key = key
    c.aead = aead
    atomic.AddUint32(&c.keyID, 1)
    
    return c.keyID, nil
}
