package encryption

import (
    "testing"
)

func TestECDHKeyExchangerCreation(t *testing.T) {
    keyExchanger, err := NewECDHKeyExchanger()
    if err != nil {
        t.Fatalf("NewECDHKeyExchanger() error = %v", err)
    }
    
    if keyExchanger == nil {
        t.Fatalf("NewECDHKeyExchanger() returned nil key exchanger")
    }
    
    publicKey := keyExchanger.GetPublicKey()
    if len(publicKey) == 0 {
        t.Errorf("keyExchanger.GetPublicKey() returned empty public key")
    }
}

func TestECDHKeyExchangerGenerateKeyPair(t *testing.T) {
    keyExchanger, err := NewECDHKeyExchanger()
    if err != nil {
        t.Fatalf("NewECDHKeyExchanger() error = %v", err)
    }
    
    initialPublicKey := keyExchanger.GetPublicKey()
    
    newPublicKey, err := keyExchanger.GenerateKeyPair()
    if err != nil {
        t.Fatalf("GenerateKeyPair() error = %v", err)
    }
    
    if len(newPublicKey) == 0 {
        t.Errorf("GenerateKeyPair() returned empty public key")
    }
    
    // New public key should be different from initial public key
    if string(newPublicKey) == string(initialPublicKey) {
        t.Errorf("GenerateKeyPair() returned same public key")
    }
}

func TestECDHKeyExchangerComputeSharedSecret(t *testing.T) {
    // Create two key exchangers
    keyExchanger1, err := NewECDHKeyExchanger()
    if err != nil {
        t.Fatalf("NewECDHKeyExchanger() error = %v", err)
    }
    
    keyExchanger2, err := NewECDHKeyExchanger()
    if err != nil {
        t.Fatalf("NewECDHKeyExchanger() error = %v", err)
    }
    
    // Generate key pairs
    publicKey1, err := keyExchanger1.GenerateKeyPair()
    if err != nil {
        t.Fatalf("GenerateKeyPair() error = %v", err)
    }
    
    publicKey2, err := keyExchanger2.GenerateKeyPair()
    if err != nil {
        t.Fatalf("GenerateKeyPair() error = %v", err)
    }
    
    // Compute shared secrets
    sharedSecret1, err := keyExchanger1.ComputeSharedSecret(publicKey2)
    if err != nil {
        t.Fatalf("ComputeSharedSecret() error = %v", err)
    }
    
    sharedSecret2, err := keyExchanger2.ComputeSharedSecret(publicKey1)
    if err != nil {
        t.Fatalf("ComputeSharedSecret() error = %v", err)
    }
    
    // Shared secrets should be the same
    if string(sharedSecret1) != string(sharedSecret2) {
        t.Errorf("Shared secrets do not match")
    }
    
    // Test with invalid public key
    invalidPublicKey := []byte("invalid public key")
    _, err = keyExchanger1.ComputeSharedSecret(invalidPublicKey)
    if err == nil {
        t.Errorf("ComputeSharedSecret() with invalid public key should return error")
    }
}

func TestKeyExchangeMessage(t *testing.T) {
    // Create a key exchange message
    publicKey := []byte("test public key")
    keyRotationTime := int64(1234567890)
    message := NewKeyExchangeMessage(EncryptionAES, publicKey, keyRotationTime)
    
    if message.Type != "key_exchange" {
        t.Errorf("message.Type = %v, want %v", message.Type, "key_exchange")
    }
    
    if message.EncryptionType != string(EncryptionAES) {
        t.Errorf("message.EncryptionType = %v, want %v", message.EncryptionType, string(EncryptionAES))
    }
    
    if string(message.PublicKey) != string(publicKey) {
        t.Errorf("message.PublicKey = %v, want %v", string(message.PublicKey), string(publicKey))
    }
    
    if message.KeyRotationTime != keyRotationTime {
        t.Errorf("message.KeyRotationTime = %v, want %v", message.KeyRotationTime, keyRotationTime)
    }
    
    // Convert to JSON and back
    jsonData, err := message.ToJSON()
    if err != nil {
        t.Fatalf("ToJSON() error = %v", err)
    }
    
    var parsedMessage KeyExchangeMessage
    err = parsedMessage.FromJSON(jsonData)
    if err != nil {
        t.Fatalf("FromJSON() error = %v", err)
    }
    
    if parsedMessage.Type != message.Type {
        t.Errorf("parsedMessage.Type = %v, want %v", parsedMessage.Type, message.Type)
    }
    
    if parsedMessage.EncryptionType != message.EncryptionType {
        t.Errorf("parsedMessage.EncryptionType = %v, want %v", parsedMessage.EncryptionType, message.EncryptionType)
    }
    
    if string(parsedMessage.PublicKey) != string(message.PublicKey) {
        t.Errorf("parsedMessage.PublicKey = %v, want %v", string(parsedMessage.PublicKey), string(message.PublicKey))
    }
    
    if parsedMessage.KeyRotationTime != message.KeyRotationTime {
        t.Errorf("parsedMessage.KeyRotationTime = %v, want %v", parsedMessage.KeyRotationTime, message.KeyRotationTime)
    }
}
