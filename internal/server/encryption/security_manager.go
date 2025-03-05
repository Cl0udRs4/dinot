package encryption

import (
	"sync"
	"time"
)

// SecurityConfig holds the configuration for all security features
type SecurityConfig struct {
	// Authentication configuration
	Auth AuthConfig
	// Key rotation configuration
	KeyRotation KeyRotationConfig
	// Forward secrecy configuration
	ForwardSecrecy ForwardSecrecyConfig
	// Obfuscation configuration
	Obfuscation ObfuscationConfig
}

// DefaultSecurityConfig returns the default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		Auth:          DefaultAuthConfig(),
		KeyRotation:   DefaultKeyRotationConfig(),
		ForwardSecrecy: DefaultForwardSecrecyConfig(),
		Obfuscation:   DefaultObfuscationConfig(),
	}
}

// SecurityManager manages all security features
type SecurityManager struct {
	config         SecurityConfig
	authenticator  *Authenticator
	keyRotator     *KeyRotator
	forwardSecrecy *ForwardSecrecy
	obfuscator     *Obfuscator
	signatureVerifier *SignatureVerifier
	messageProcessor *MessageProcessor
	mu             sync.RWMutex
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(config SecurityConfig) (*SecurityManager, error) {
	// Create authenticator
	authenticator := NewAuthenticator(config.Auth)

	// Create key rotator
	keyRotator := NewKeyRotator(config.KeyRotation)

	// Create forward secrecy handler
	forwardSecrecy, err := NewForwardSecrecy(config.ForwardSecrecy)
	if err != nil {
		return nil, err
	}

	// Create obfuscator
	obfuscator, err := NewObfuscator(config.Obfuscation)
	if err != nil {
		return nil, err
	}

	// Create signature verifier
	signatureVerifier := NewSignatureVerifier()

	// Create message processor
	messageProcessor := NewMessageProcessor()

	return &SecurityManager{
		config:         config,
		authenticator:  authenticator,
		keyRotator:     keyRotator,
		forwardSecrecy: forwardSecrecy,
		obfuscator:     obfuscator,
		signatureVerifier: signatureVerifier,
		messageProcessor: messageProcessor,
		mu:             sync.RWMutex{},
	}, nil
}

// Start starts all security features
func (m *SecurityManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Start key rotation
	if m.config.KeyRotation.Enabled {
		if err := m.keyRotator.Start(); err != nil {
			return err
		}
	}

	// Start forward secrecy
	if m.config.ForwardSecrecy.Enabled {
		if err := m.forwardSecrecy.Start(); err != nil {
			return err
		}
	}

	return nil
}

// Stop stops all security features
func (m *SecurityManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop key rotation
	m.keyRotator.Stop()

	// Stop forward secrecy
	m.forwardSecrecy.Stop()
}

// GetAuthenticator returns the authenticator
func (m *SecurityManager) GetAuthenticator() *Authenticator {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.authenticator
}

// GetKeyRotator returns the key rotator
func (m *SecurityManager) GetKeyRotator() *KeyRotator {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.keyRotator
}

// GetForwardSecrecy returns the forward secrecy handler
func (m *SecurityManager) GetForwardSecrecy() *ForwardSecrecy {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.forwardSecrecy
}

// GetObfuscator returns the obfuscator
func (m *SecurityManager) GetObfuscator() *Obfuscator {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.obfuscator
}

// GetSignatureVerifier returns the signature verifier
func (m *SecurityManager) GetSignatureVerifier() *SignatureVerifier {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.signatureVerifier
}

// GetMessageProcessor returns the message processor
func (m *SecurityManager) GetMessageProcessor() *MessageProcessor {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.messageProcessor
}

// RegisterClient registers a client with all security components
func (m *SecurityManager) RegisterClient(clientID string) *ClientEncryption {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Register with message processor
	clientEnc := m.messageProcessor.RegisterClient(clientID)

	// Register with key rotator
	m.keyRotator.RegisterClient(clientID, clientEnc)

	return clientEnc
}

// UnregisterClient unregisters a client from all security components
func (m *SecurityManager) UnregisterClient(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Unregister from key rotator
	m.keyRotator.UnregisterClient(clientID)
}

// ProcessIncomingMessage processes an incoming message with all security features
func (m *SecurityManager) ProcessIncomingMessage(clientID string, data []byte) ([]byte, error) {
	// First, remove any obfuscation
	if m.config.Obfuscation.EnableMimicry {
		var err error
		data, err = m.obfuscator.RemoveMimicry(data)
		if err != nil {
			return nil, err
		}
	}

	if m.config.Obfuscation.EnablePadding {
		var err error
		data, err = m.obfuscator.RemovePadding(data)
		if err != nil {
			return nil, err
		}
	}

	// Then process with message processor (handles encryption)
	return m.messageProcessor.ProcessIncomingMessage(clientID, data)
}

// ProcessOutgoingMessage processes an outgoing message with all security features
func (m *SecurityManager) ProcessOutgoingMessage(clientID string, data []byte) ([]byte, error) {
	// First process with message processor (handles encryption)
	data, err := m.messageProcessor.ProcessOutgoingMessage(clientID, data)
	if err != nil {
		return nil, err
	}

	// Then add padding if enabled
	if m.config.Obfuscation.EnablePadding {
		data, err = m.obfuscator.AddPadding(data)
		if err != nil {
			return nil, err
		}
	}

	// Finally add mimicry if enabled
	if m.config.Obfuscation.EnableMimicry {
		data, err = m.obfuscator.ApplyMimicry(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// VerifyModuleSignature verifies a module signature
func (m *SecurityManager) VerifyModuleSignature(moduleName string, moduleData, signature []byte) error {
	return m.signatureVerifier.Verify(moduleName, moduleData, signature)
}

// GenerateToken generates an authentication token for a client
func (m *SecurityManager) GenerateToken(clientID, role string) (string, error) {
	return m.authenticator.GenerateJWT(clientID, role)
}

// VerifyToken verifies an authentication token
func (m *SecurityManager) VerifyToken(token string) (*Claims, error) {
	return m.authenticator.VerifyJWT(token)
}

// ApplyJitter applies timing jitter and returns the delay duration
func (m *SecurityManager) ApplyJitter() time.Duration {
	return m.obfuscator.ApplyJitter()
}
