package encryption

import (
    "testing"
)

func TestGenerateRandomBytes(t *testing.T) {
    tests := []struct {
        name   string
        length int
    }{
        {"Zero length", 0},
        {"Small length", 16},
        {"Medium length", 64},
        {"Large length", 1024},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            bytes, err := GenerateRandomBytes(tt.length)
            if err != nil {
                t.Fatalf("GenerateRandomBytes(%d) error = %v", tt.length, err)
            }
            
            if len(bytes) != tt.length {
                t.Errorf("GenerateRandomBytes(%d) returned %d bytes, want %d", tt.length, len(bytes), tt.length)
            }
            
            // For non-zero length, check that the bytes are not all zeros
            if tt.length > 0 {
                allZeros := true
                for _, b := range bytes {
                    if b != 0 {
                        allZeros = false
                        break
                    }
                }
                
                if allZeros {
                    t.Errorf("GenerateRandomBytes(%d) returned all zeros, expected random data", tt.length)
                }
            }
        })
    }
}

func TestEncryptionTypes(t *testing.T) {
    tests := []struct {
        name string
        encType EncryptionType
        expected string
    }{
        {"None", EncryptionNone, "none"},
        {"AES", EncryptionAES, "aes"},
        {"ChaCha20", EncryptionChaCha20, "chacha20"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if string(tt.encType) != tt.expected {
                t.Errorf("EncryptionType %v = %v, want %v", tt.name, tt.encType, tt.expected)
            }
        })
    }
}
