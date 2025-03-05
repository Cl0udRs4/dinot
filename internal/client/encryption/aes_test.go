package encryption

import (
    "strings"
    "testing"
)

func TestAESEncrypterCreation(t *testing.T) {
    tests := []struct {
        name    string
        keySize int
        wantErr bool
    }{
        {"Valid 128-bit key", 16, false},
        {"Valid 192-bit key", 24, false},
        {"Valid 256-bit key", 32, false},
        {"Invalid key size", 20, true},
        {"Zero key size", 0, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            encrypter, err := NewAESEncrypter(tt.keySize)
            
            if tt.wantErr {
                if err == nil {
                    t.Errorf("NewAESEncrypter(%d) expected error, got nil", tt.keySize)
                }
                return
            }
            
            if err != nil {
                t.Fatalf("NewAESEncrypter(%d) error = %v", tt.keySize, err)
            }
            
            if encrypter == nil {
                t.Fatalf("NewAESEncrypter(%d) returned nil encrypter", tt.keySize)
            }
            
            if encrypter.GetType() != EncryptionAES {
                t.Errorf("encrypter.GetType() = %v, want %v", encrypter.GetType(), EncryptionAES)
            }
            
            if encrypter.GetKeyID() != 1 {
                t.Errorf("encrypter.GetKeyID() = %v, want %v", encrypter.GetKeyID(), 1)
            }
        })
    }
}

func TestAESEncrypterWithKey(t *testing.T) {
    tests := []struct {
        name    string
        keySize int
        wantErr bool
    }{
        {"Valid 128-bit key", 16, false},
        {"Valid 192-bit key", 24, false},
        {"Valid 256-bit key", 32, false},
        {"Invalid key size", 20, true},
        {"Zero key size", 0, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            key, err := GenerateRandomBytes(tt.keySize)
            if err != nil && !tt.wantErr {
                t.Fatalf("GenerateRandomBytes(%d) error = %v", tt.keySize, err)
            }
            
            encrypter, err := NewAESEncrypterWithKey(key)
            
            if tt.wantErr {
                if err == nil {
                    t.Errorf("NewAESEncrypterWithKey() expected error, got nil")
                }
                return
            }
            
            if err != nil {
                t.Fatalf("NewAESEncrypterWithKey() error = %v", err)
            }
            
            if encrypter == nil {
                t.Fatalf("NewAESEncrypterWithKey() returned nil encrypter")
            }
            
            if encrypter.GetType() != EncryptionAES {
                t.Errorf("encrypter.GetType() = %v, want %v", encrypter.GetType(), EncryptionAES)
            }
            
            if encrypter.GetKeyID() != 1 {
                t.Errorf("encrypter.GetKeyID() = %v, want %v", encrypter.GetKeyID(), 1)
            }
        })
    }
}

func TestAESEncrypterEncryptDecrypt(t *testing.T) {
    tests := []struct {
        name      string
        keySize   int
        plaintext string
        wantErr   bool
    }{
        {"128-bit key", 16, "Hello, world!", false},
        {"192-bit key", 24, "This is a test message", false},
        {"256-bit key", 32, "Lorem ipsum dolor sit amet", false},
        {"Empty message", 32, "", true},
        {"Long message", 32, strings.Repeat("A", 1000), false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            encrypter, err := NewAESEncrypter(tt.keySize)
            if err != nil {
                t.Fatalf("Failed to create AES encrypter: %v", err)
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
            if len(ciphertext) <= 4+encrypter.gcm.NonceSize() {
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

func TestAESEncrypterKeyRotation(t *testing.T) {
    encrypter, err := NewAESEncrypter(32)
    if err != nil {
        t.Fatalf("Failed to create AES encrypter: %v", err)
    }
    
    initialKeyID := encrypter.GetKeyID()
    plaintext := []byte("Test message for key rotation")
    
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
