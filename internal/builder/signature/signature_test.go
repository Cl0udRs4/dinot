package signature

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSignatureManager(t *testing.T) {
	// Create a temporary directory for keys
	tempDir, err := os.MkdirTemp("", "signature-test-")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	// Create a signature manager
	sm := NewSignatureManager(privateKeyPath, publicKeyPath)

	// Generate a key pair
	if err := sm.GenerateKeyPair(2048); err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Verify that the key files were created
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		t.Fatalf("Private key file was not created")
	}

	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		t.Fatalf("Public key file was not created")
	}

	// Test signing and verification
	testData := []byte("test data")
	signature, err := sm.SignCode(testData)
	if err != nil {
		t.Fatalf("Failed to sign data: %v", err)
	}

	// Verify the signature
	if err := sm.VerifySignature(testData, signature); err != nil {
		t.Fatalf("Signature verification failed: %v", err)
	}

	// Test verification with modified data
	modifiedData := []byte("modified data")
	if err := sm.VerifySignature(modifiedData, signature); err != ErrSignatureVerificationFailed {
		t.Fatalf("Expected signature verification to fail, got: %v", err)
	}

	// Create a new signature manager and load the keys
	sm2 := NewSignatureManager(privateKeyPath, publicKeyPath)
	if err := sm2.LoadKeys(); err != nil {
		t.Fatalf("Failed to load keys: %v", err)
	}

	// Verify the signature with the loaded keys
	if err := sm2.VerifySignature(testData, signature); err != nil {
		t.Fatalf("Signature verification failed with loaded keys: %v", err)
	}
}
