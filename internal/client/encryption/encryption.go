package encryption

import (
    "crypto/rand"
    "errors"
    "io"
)

// EncryptionType represents the type of encryption
type EncryptionType string

const (
    // EncryptionNone represents no encryption
    EncryptionNone EncryptionType = "none"
    // EncryptionAES represents AES encryption
    EncryptionAES EncryptionType = "aes"
    // EncryptionChaCha20 represents ChaCha20 encryption
    EncryptionChaCha20 EncryptionType = "chacha20"
)

var (
    // ErrInvalidKey is returned when an invalid key is provided
    ErrInvalidKey = errors.New("invalid encryption key")
    // ErrInvalidData is returned when invalid data is provided
    ErrInvalidData = errors.New("invalid data for encryption/decryption")
    // ErrInvalidNonce is returned when an invalid nonce is provided
    ErrInvalidNonce = errors.New("invalid nonce")
)

// Encrypter defines the interface for encryption and decryption
type Encrypter interface {
    // Encrypt encrypts the plaintext and returns the ciphertext
    Encrypt(plaintext []byte) ([]byte, error)
    
    // Decrypt decrypts the ciphertext and returns the plaintext
    Decrypt(ciphertext []byte) ([]byte, error)
    
    // GetType returns the encryption type
    GetType() EncryptionType
    
    // GetKeyID returns the current key ID
    GetKeyID() uint32
    
    // RotateKey generates a new key and returns the new key ID
    RotateKey() (uint32, error)
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
