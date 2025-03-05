package encryption

import (
    "strings"
    "testing"

    "golang.org/x/crypto/chacha20poly1305"
)

func TestChaCha20EncrypterCreation(t *testing.T) {
    encrypter, err := NewChaCha20Encrypter()
    if err != nil {
        t.Fatalf("NewChaCha20Encrypter() error = %v", err)
    }
    
    if encrypter == nil {
        t.Fatalf("NewChaCha20Encrypter() returned nil encrypter")
    }
    
    if encrypter.GetType() != EncryptionChaCha20 {
        t.Errorf("encrypter.GetType() = %v, want %v", encrypter.GetType(), EncryptionChaCha20)
    }
    
    if encrypter.GetKeyID() != 1 {
        t.Errorf("encrypter.GetKeyID() = %v, want %v", encrypter.GetKeyID(), 1)
    }
    
    if len(encrypter.key) != chacha20poly1305.KeySize {
        t.Errorf("encrypter.key length = %v, want %v", len(encrypter.key), chacha20poly1305.KeySize)
    }
}

func TestChaCha20EncrypterWithKey(t *testing.T) {
    tests := []struct {
        name    string
        keySize int
        wantErr bool
    }{
        {"Valid key", chacha20poly1305.KeySize, false},
        {"Invalid key size", chacha20poly1305.KeySize - 1, true},
        {"Invalid key size", chacha20poly1305.KeySize + 1, true},
        {"Zero key size", 0, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            key, err := GenerateRandomBytes(tt.keySize)
            if err != nil && !tt.wantErr {
                t.Fatalf("GenerateRandomBytes(%d) error = %v", tt.keySize, err)
            }
            
            encrypter, err := NewChaCha20EncrypterWithKey(key)
            
            if tt.wantErr {
                if err == nil {
                    t.Errorf("NewChaCha20EncrypterWithKey() expected error, got nil")
                }
                return
            }
            
            if err != nil {
                t.Fatalf("NewChaCha20EncrypterWithKey() error = %v", err)
            }
            
            if encrypter == nil {
                t.Fatalf("NewChaCha20EncrypterWithKey() returned nil encrypter")
            }
            
            if encrypter.GetType() != EncryptionChaCha20 {
                t.Errorf("encrypter.GetType() = %v, want %v", encrypter.GetType(), EncryptionChaCha20)
            }
            
            if encrypter.GetKeyID() != 1 {
                t.Errorf("encrypter.GetKeyID() = %v, want %v", encrypter.GetKeyID(), 1)
            }
        })
    }
}

func TestChaCha20EncrypterEncryptDecrypt(t *testing.T) {
    tests := []struct {
        name      string
        plaintext string
        wantErr   bool
    }{
        {"Simple message", "Hello, world!", false},
        {"Medium message", "This is a test message for ChaCha20 encryption", false},
        {"Long message", strings.Repeat("A", 1000), false},
        {"Empty message", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            encrypter, err := NewChaCha20Encrypter()
            if err != nil {
                t.Fatalf("Failed to create ChaCha20 encrypter: %v", err)
            }
            
            ciphertext, err := encrypter.Encrypt([]byte(tt.plaintext))
            if tt.wantErr {
                if err == nil {
                    t.Errorf("Encrypt() expected error for empty plaintext, got nil")
                }
                return
            }
            
            if err != nil {
                t.Fatalf("Encrypt() error = %v", err)
            }
            
            // Ciphertext should be different from plaintext
            if string(ciphertext) == tt.plaintext {
                t.Errorf("Encrypt() returned plaintext, expected ciphertext")
            }
            
            // Ciphertext should include key ID and nonce
            if len(ciphertext) <= 4+encrypter.aead.NonceSize() {
                t.Errorf("Ciphertext too short, missing key ID or nonce")
            }
            
            decrypted, err := encrypter.Decrypt(ciphertext)
            if err != nil {
                t.Fatalf("Decrypt() error = %v", err)
            }
            
            if string(decrypted) != tt.plaintext {
                t.Errorf("Decrypt() = %v, want %v", string(decrypted), tt.plaintext)
            }
        })
    }
}

func TestChaCha20EncrypterKeyRotation(t *testing.T) {
    encrypter, err := NewChaCha20Encrypter()
    if err != nil {
        t.Fatalf("Failed to create ChaCha20 encrypter: %v", err)
    }
    
    initialKeyID := encrypter.GetKeyID()
    plaintext := []byte("Test message for ChaCha20 key rotation")
    
    // Encrypt with initial key
    ciphertext, err := encrypter.Encrypt(plaintext)
    if err != nil {
        t.Fatalf("Encrypt() error = %v", err)
    }
    
    // Decrypt with initial key
    decrypted, err := encrypter.Decrypt(ciphertext)
    if err != nil {
        t.Fatalf("Decrypt() error = %v", err)
    }
    
    if string(decrypted) != string(plaintext) {
        t.Errorf("Decrypt() = %v, want %v", string(decrypted), string(plaintext))
    }
    
    // Rotate key
    newKeyID, err := encrypter.RotateKey()
    if err != nil {
        t.Fatalf("RotateKey() error = %v", err)
    }
    
    if newKeyID <= initialKeyID {
        t.Errorf("RotateKey() = %v, want > %v", newKeyID, initialKeyID)
    }
    
    // Encrypt with new key
    newCiphertext, err := encrypter.Encrypt(plaintext)
    if err != nil {
        t.Fatalf("Encrypt() after rotation error = %v", err)
    }
    
    // Decrypt with new key
    newDecrypted, err := encrypter.Decrypt(newCiphertext)
    if err != nil {
        t.Fatalf("Decrypt() after rotation error = %v", err)
    }
    
    if string(newDecrypted) != string(plaintext) {
        t.Errorf("Decrypt() after rotation = %v, want %v", string(newDecrypted), string(plaintext))
    }
    
    // Try to decrypt old ciphertext with new key (should fail)
    _, err = encrypter.Decrypt(ciphertext)
    if err == nil {
        t.Errorf("Decrypt() old ciphertext with new key should fail")
    }
}
