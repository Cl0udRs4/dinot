package encryption

import (
	"testing"
	"time"
)

func TestSecurityManager(t *testing.T) {
	// Create a security manager with default configuration
	config := DefaultSecurityConfig()
	manager, err := NewSecurityManager(config)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Test starting and stopping the security manager
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start security manager: %v", err)
	}
	manager.Stop()

	// Test client registration
	clientID := "test-client"
	clientEnc := manager.RegisterClient(clientID)
	if clientEnc == nil {
		t.Fatal("Failed to register client")
	}

	// Test client unregistration
	manager.UnregisterClient(clientID)

	// Re-register client for further tests
	clientEnc = manager.RegisterClient(clientID)

	// Test authentication
	token, err := manager.GenerateToken(clientID, "client")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := manager.VerifyToken(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if claims.ClientID != clientID {
		t.Fatalf("Token verification returned wrong client ID: got %s, want %s", claims.ClientID, clientID)
	}

	// Test jitter
	jitter := manager.ApplyJitter()
	if config.Obfuscation.EnableJitter && (jitter < time.Duration(config.Obfuscation.MinJitter)*time.Millisecond || jitter > time.Duration(config.Obfuscation.MaxJitter)*time.Millisecond) {
		t.Fatalf("Jitter out of range: got %v, want between %v and %v", jitter, config.Obfuscation.MinJitter, config.Obfuscation.MaxJitter)
	}
}

func TestAuthentication(t *testing.T) {
	// Create an authenticator with default configuration
	config := DefaultAuthConfig()
	auth := NewAuthenticator(config)

	// Test HMAC generation and verification
	data := []byte("test data")
	hmac, err := auth.GenerateHMAC(data)
	if err != nil {
		t.Fatalf("Failed to generate HMAC: %v", err)
	}

	if err := auth.VerifyHMAC(data, hmac); err != nil {
		t.Fatalf("Failed to verify HMAC: %v", err)
	}

	// Test JWT generation and verification
	clientID := "test-client"
	role := "client"
	token, err := auth.GenerateJWT(clientID, role)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	claims, err := auth.VerifyJWT(token)
	if err != nil {
		t.Fatalf("Failed to verify JWT: %v", err)
	}

	if claims.ClientID != clientID {
		t.Fatalf("JWT verification returned wrong client ID: got %s, want %s", claims.ClientID, clientID)
	}

	if claims.Role != role {
		t.Fatalf("JWT verification returned wrong role: got %s, want %s", claims.Role, role)
	}

	// Test basic auth
	username := "admin"
	password := "password"
	basicAuth := auth.GenerateBasicAuth(username, password)

	if err := auth.VerifyBasicAuth(basicAuth, username, password); err != nil {
		t.Fatalf("Failed to verify basic auth: %v", err)
	}
}

func TestKeyRotation(t *testing.T) {
	// Create a key rotator with default configuration
	config := DefaultKeyRotationConfig()
	// Set shorter interval for testing
	config.Interval = 100 * time.Millisecond
	rotator := NewKeyRotator(config)

	// Test starting and stopping the key rotator
	if err := rotator.Start(); err != nil {
		t.Fatalf("Failed to start key rotator: %v", err)
	}

	// Register a client
	clientID := "test-client"
	clientEnc := NewClientEncryption(clientID)
	rotator.RegisterClient(clientID, clientEnc)

	// Wait for a key rotation
	time.Sleep(150 * time.Millisecond)

	// Check that a key rotation occurred
	lastRotate := rotator.GetLastRotateTime()
	if time.Since(lastRotate) > 150*time.Millisecond {
		t.Fatal("Key rotation did not occur")
	}

	// Test force rotation
	if err := rotator.ForceRotate(); err != nil {
		t.Fatalf("Failed to force key rotation: %v", err)
	}

	// Test setting rotation interval
	if err := rotator.SetRotationInterval(200 * time.Millisecond); err != nil {
		t.Fatalf("Failed to set rotation interval: %v", err)
	}

	// Test enabling and disabling key rotation
	rotator.Disable()
	if rotator.IsEnabled() {
		t.Fatal("Key rotation should be disabled")
	}

	rotator.Enable()
	if !rotator.IsEnabled() {
		t.Fatal("Key rotation should be enabled")
	}

	// Stop the key rotator
	rotator.Stop()

	// Test unregistering a client
	rotator.UnregisterClient(clientID)
}

func TestForwardSecrecy(t *testing.T) {
	// Create a forward secrecy handler with default configuration
	config := DefaultForwardSecrecyConfig()
	// Set shorter interval for testing
	config.KeyRotationInterval = 100 * time.Millisecond
	fs, err := NewForwardSecrecy(config)
	if err != nil {
		t.Fatalf("Failed to create forward secrecy handler: %v", err)
	}

	// Test starting and stopping the forward secrecy handler
	if err := fs.Start(); err != nil {
		t.Fatalf("Failed to start forward secrecy handler: %v", err)
	}

	// Get the public key
	publicKey := fs.GetPublicKey()
	if publicKey == nil {
		t.Fatal("Failed to get public key")
	}

	// Create another forward secrecy handler for the peer
	peerFS, err := NewForwardSecrecy(config)
	if err != nil {
		t.Fatalf("Failed to create peer forward secrecy handler: %v", err)
	}

	// Get the peer's public key
	peerPublicKey := peerFS.GetPublicKey()
	if peerPublicKey == nil {
		t.Fatal("Failed to get peer public key")
	}

	// Compute shared secrets
	sharedSecret1, err := fs.ComputeSharedSecret(peerPublicKey)
	if err != nil {
		t.Fatalf("Failed to compute shared secret: %v", err)
	}

	sharedSecret2, err := peerFS.ComputeSharedSecret(publicKey)
	if err != nil {
		t.Fatalf("Failed to compute peer shared secret: %v", err)
	}

	// Verify that both sides computed the same shared secret
	if len(sharedSecret1) != len(sharedSecret2) {
		t.Fatalf("Shared secret lengths differ: got %d and %d", len(sharedSecret1), len(sharedSecret2))
	}

	for i := 0; i < len(sharedSecret1); i++ {
		if sharedSecret1[i] != sharedSecret2[i] {
			t.Fatalf("Shared secrets differ at index %d: got %d and %d", i, sharedSecret1[i], sharedSecret2[i])
		}
	}

	// Wait for a key rotation
	time.Sleep(150 * time.Millisecond)

	// Check that a key rotation occurred
	lastRotate := fs.GetLastRotateTime()
	if time.Since(lastRotate) > 150*time.Millisecond {
		t.Fatal("Key rotation did not occur")
	}

	// Test force rotation
	if err := fs.ForceRotate(); err != nil {
		t.Fatalf("Failed to force key rotation: %v", err)
	}

	// Test setting rotation interval
	if err := fs.SetKeyRotationInterval(200 * time.Millisecond); err != nil {
		t.Fatalf("Failed to set rotation interval: %v", err)
	}

	// Test enabling and disabling forward secrecy
	fs.Disable()
	if fs.IsEnabled() {
		t.Fatal("Forward secrecy should be disabled")
	}

	fs.Enable()
	if !fs.IsEnabled() {
		t.Fatal("Forward secrecy should be enabled")
	}

	// Stop the forward secrecy handler
	fs.Stop()
}

func TestObfuscation(t *testing.T) {
	// Create an obfuscator with default configuration
	config := DefaultObfuscationConfig()
	obfuscator, err := NewObfuscator(config)
	if err != nil {
		t.Fatalf("Failed to create obfuscator: %v", err)
	}

	// Test padding
	data := []byte("test data")
	paddedData, err := obfuscator.AddPadding(data)
	if err != nil {
		t.Fatalf("Failed to add padding: %v", err)
	}

	if len(paddedData) <= len(data) {
		t.Fatalf("Padding did not increase data size: got %d, want > %d", len(paddedData), len(data))
	}

	unpaddedData, err := obfuscator.RemovePadding(paddedData)
	if err != nil {
		t.Fatalf("Failed to remove padding: %v", err)
	}

	if len(unpaddedData) != len(data) {
		t.Fatalf("Unpadded data length differs from original: got %d, want %d", len(unpaddedData), len(data))
	}

	for i := 0; i < len(data); i++ {
		if unpaddedData[i] != data[i] {
			t.Fatalf("Unpadded data differs from original at index %d: got %d, want %d", i, unpaddedData[i], data[i])
		}
	}

	// Test jitter
	jitter := obfuscator.ApplyJitter()
	if config.EnableJitter && (jitter < time.Duration(config.MinJitter)*time.Millisecond || jitter > time.Duration(config.MaxJitter)*time.Millisecond) {
		t.Fatalf("Jitter out of range: got %v, want between %v and %v", jitter, config.MinJitter, config.MaxJitter)
	}
}

func TestSignatureVerification(t *testing.T) {
	// Create a signature verifier
	verifier := NewSignatureVerifier()

	// Generate a key pair
	if err := verifier.GenerateKeyPair(2048); err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Get the public key
	publicKey, err := verifier.GetPublicKeyFromPrivate()
	if err != nil {
		t.Fatalf("Failed to get public key: %v", err)
	}

	// Sign some data
	data := []byte("test data")
	signature, err := verifier.Sign(data)
	if err != nil {
		t.Fatalf("Failed to sign data: %v", err)
	}

	// Verify the signature
	if err := verifier.VerifyWithKey(publicKey, data, signature); err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}

	// Modify the data and verify that the signature is invalid
	modifiedData := []byte("modified data")
	if err := verifier.VerifyWithKey(publicKey, modifiedData, signature); err == nil {
		t.Fatal("Signature verification should fail with modified data")
	}
}
