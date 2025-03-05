package encryption

import (
    "testing"
    "time"
)

func TestMessageCreation(t *testing.T) {
    tests := []struct {
        name          string
        encryptionType EncryptionType
        keyID         uint32
        payload       []byte
    }{
        {"AES encryption", EncryptionAES, 1, []byte("Test payload")},
        {"ChaCha20 encryption", EncryptionChaCha20, 2, []byte("Another test payload")},
        {"No encryption", EncryptionNone, 0, []byte("Unencrypted payload")},
        {"Empty payload", EncryptionAES, 3, []byte{}},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            message := NewMessage(tt.encryptionType, tt.keyID, tt.payload)
            
            if message.Header.Version != 1 {
                t.Errorf("message.Header.Version = %v, want %v", message.Header.Version, 1)
            }
            
            if message.Header.Encryption != tt.encryptionType {
                t.Errorf("message.Header.Encryption = %v, want %v", message.Header.Encryption, tt.encryptionType)
            }
            
            if message.Header.KeyID != tt.keyID {
                t.Errorf("message.Header.KeyID = %v, want %v", message.Header.KeyID, tt.keyID)
            }
            
            if message.Header.Timestamp <= 0 {
                t.Errorf("message.Header.Timestamp = %v, should be > 0", message.Header.Timestamp)
            }
            
            if string(message.Payload) != string(tt.payload) {
                t.Errorf("message.Payload = %v, want %v", string(message.Payload), string(tt.payload))
            }
        })
    }
}

func TestMessageJSONConversion(t *testing.T) {
    tests := []struct {
        name          string
        encryptionType EncryptionType
        keyID         uint32
        payload       []byte
    }{
        {"AES encryption", EncryptionAES, 1, []byte("Test payload")},
        {"ChaCha20 encryption", EncryptionChaCha20, 2, []byte("Another test payload")},
        {"No encryption", EncryptionNone, 0, []byte("Unencrypted payload")},
        {"Empty payload", EncryptionAES, 3, []byte{}},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            message := NewMessage(tt.encryptionType, tt.keyID, tt.payload)
            
            // Set a fixed timestamp for testing
            message.Header.Timestamp = time.Now().Unix()
            
            // Convert to JSON
            jsonData, err := message.ToJSON()
            if err != nil {
                t.Fatalf("ToJSON() error = %v", err)
            }
            
            // Parse from JSON
            var parsedMessage Message
            err = parsedMessage.FromJSON(jsonData)
            if err != nil {
                t.Fatalf("FromJSON() error = %v", err)
            }
            
            // Check that the parsed message matches the original
            if parsedMessage.Header.Version != message.Header.Version {
                t.Errorf("parsedMessage.Header.Version = %v, want %v", parsedMessage.Header.Version, message.Header.Version)
            }
            
            if parsedMessage.Header.Encryption != message.Header.Encryption {
                t.Errorf("parsedMessage.Header.Encryption = %v, want %v", parsedMessage.Header.Encryption, message.Header.Encryption)
            }
            
            if parsedMessage.Header.KeyID != message.Header.KeyID {
                t.Errorf("parsedMessage.Header.KeyID = %v, want %v", parsedMessage.Header.KeyID, message.Header.KeyID)
            }
            
            if parsedMessage.Header.Timestamp != message.Header.Timestamp {
                t.Errorf("parsedMessage.Header.Timestamp = %v, want %v", parsedMessage.Header.Timestamp, message.Header.Timestamp)
            }
            
            if string(parsedMessage.Payload) != string(message.Payload) {
                t.Errorf("parsedMessage.Payload = %v, want %v", string(parsedMessage.Payload), string(message.Payload))
            }
        })
    }
}

func TestMessageWithEncryption(t *testing.T) {
    // Create an AES encrypter
    aesEncrypter, err := NewAESEncrypter(32)
    if err != nil {
        t.Fatalf("Failed to create AES encrypter: %v", err)
    }
    
    // Create a ChaCha20 encrypter
    chaCha20Encrypter, err := NewChaCha20Encrypter()
    if err != nil {
        t.Fatalf("Failed to create ChaCha20 encrypter: %v", err)
    }
    
    tests := []struct {
        name      string
        encrypter Encrypter
        plaintext string
    }{
        {"AES encryption", aesEncrypter, "Test message for AES encryption"},
        {"ChaCha20 encryption", chaCha20Encrypter, "Test message for ChaCha20 encryption"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Encrypt the plaintext
            ciphertext, err := tt.encrypter.Encrypt([]byte(tt.plaintext))
            if err != nil {
                t.Fatalf("Encrypt() error = %v", err)
            }
            
            // Create a message with the encrypted payload
            message := NewMessage(tt.encrypter.GetType(), tt.encrypter.GetKeyID(), ciphertext)
            
            // Convert to JSON
            jsonData, err := message.ToJSON()
            if err != nil {
                t.Fatalf("ToJSON() error = %v", err)
            }
            
            // Parse from JSON
            var parsedMessage Message
            err = parsedMessage.FromJSON(jsonData)
            if err != nil {
                t.Fatalf("FromJSON() error = %v", err)
            }
            
            // Decrypt the payload
            decrypted, err := tt.encrypter.Decrypt(parsedMessage.Payload)
            if err != nil {
                t.Fatalf("Decrypt() error = %v", err)
            }
            
            if string(decrypted) != tt.plaintext {
                t.Errorf("Decrypted payload = %v, want %v", string(decrypted), tt.plaintext)
            }
        })
    }
}
