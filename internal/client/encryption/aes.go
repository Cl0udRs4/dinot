package encryption

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/binary"
    "io"
    "sync"
    "sync/atomic"
)

// AESEncrypter implements the Encrypter interface using AES-GCM
type AESEncrypter struct {
    key       []byte
    keyID     uint32
    gcm       cipher.AEAD
    keySize   int
    mu        sync.RWMutex
}

// NewAESEncrypter creates a new AES encrypter with the specified key size
func NewAESEncrypter(keySize int) (*AESEncrypter, error) {
    if keySize != 16 && keySize != 24 && keySize != 32 {
        return nil, ErrInvalidKey
    }
    
    key, err := GenerateRandomBytes(keySize)
    if err != nil {
        return nil, err
    }
    
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    return &AESEncrypter{
        key:     key,
        keyID:   1,
        gcm:     gcm,
        keySize: keySize,
    }, nil
}

// NewAESEncrypterWithKey creates a new AES encrypter with the provided key
func NewAESEncrypterWithKey(key []byte) (*AESEncrypter, error) {
    keySize := len(key)
    if keySize != 16 && keySize != 24 && keySize != 32 {
        return nil, ErrInvalidKey
    }
    
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    return &AESEncrypter{
        key:     key,
        keyID:   1,
        gcm:     gcm,
        keySize: keySize,
    }, nil
}

// Encrypt encrypts the plaintext using AES-GCM
func (a *AESEncrypter) Encrypt(plaintext []byte) ([]byte, error) {
    a.mu.RLock()
    defer a.mu.RUnlock()
    
    if len(plaintext) == 0 {
        return nil, ErrInvalidData
    }
    
    // Create a nonce
    nonce := make([]byte, a.gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    
    // Prepend key ID to the nonce
    keyIDBytes := make([]byte, 4)
    binary.BigEndian.PutUint32(keyIDBytes, a.keyID)
    
    // Encrypt the plaintext
    ciphertext := a.gcm.Seal(nil, nonce, plaintext, keyIDBytes)
    
    // Combine key ID, nonce, and ciphertext
    result := make([]byte, 4+len(nonce)+len(ciphertext))
    copy(result[0:4], keyIDBytes)
    copy(result[4:4+len(nonce)], nonce)
    copy(result[4+len(nonce):], ciphertext)
    
    return result, nil
}

// Decrypt decrypts the ciphertext using AES-GCM
func (a *AESEncrypter) Decrypt(ciphertext []byte) ([]byte, error) {
    a.mu.RLock()
    defer a.mu.RUnlock()
    
    if len(ciphertext) < 4+a.gcm.NonceSize() {
        return nil, ErrInvalidData
    }
    
    // Extract key ID and verify
    keyID := binary.BigEndian.Uint32(ciphertext[0:4])
    if keyID != a.keyID {
        return nil, ErrInvalidKey
    }
    
    // Extract nonce
    nonce := ciphertext[4:4+a.gcm.NonceSize()]
    
    // Extract actual ciphertext
    actualCiphertext := ciphertext[4+a.gcm.NonceSize():]
    
    // Decrypt the ciphertext
    keyIDBytes := ciphertext[0:4]
    plaintext, err := a.gcm.Open(nil, nonce, actualCiphertext, keyIDBytes)
    if err != nil {
        return nil, err
    }
    
    return plaintext, nil
}

// GetType returns the encryption type
func (a *AESEncrypter) GetType() EncryptionType {
    return EncryptionAES
}

// GetKeyID returns the current key ID
func (a *AESEncrypter) GetKeyID() uint32 {
    return atomic.LoadUint32(&a.keyID)
}

// RotateKey generates a new key and returns the new key ID
func (a *AESEncrypter) RotateKey() (uint32, error) {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    key, err := GenerateRandomBytes(a.keySize)
    if err != nil {
        return a.keyID, err
    }
    
    block, err := aes.NewCipher(key)
    if err != nil {
        return a.keyID, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return a.keyID, err
    }
    
    a.key = key
    a.gcm = gcm
    atomic.AddUint32(&a.keyID, 1)
    
    return a.keyID, nil
}
