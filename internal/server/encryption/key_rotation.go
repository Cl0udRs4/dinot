package encryption

import (
	"errors"
	"sync"
	"time"

	clientenc "github.com/Cl0udRs4/dinot/internal/client/encryption"
)

var (
	// ErrKeyRotationDisabled is returned when key rotation is disabled
	ErrKeyRotationDisabled = errors.New("key rotation is disabled")
	// ErrInvalidRotationInterval is returned when an invalid rotation interval is provided
	ErrInvalidRotationInterval = errors.New("invalid rotation interval")
)

// KeyRotationConfig holds key rotation configuration
type KeyRotationConfig struct {
	// Enabled indicates whether key rotation is enabled
	Enabled bool
	// Interval is the interval between key rotations
	Interval time.Duration
	// GracePeriod is the period during which old keys are still valid after rotation
	GracePeriod time.Duration
	// MaxKeys is the maximum number of old keys to keep
	MaxKeys int
}

// DefaultKeyRotationConfig returns the default key rotation configuration
func DefaultKeyRotationConfig() KeyRotationConfig {
	return KeyRotationConfig{
		Enabled:     true,
		Interval:    24 * time.Hour,
		GracePeriod: 1 * time.Hour,
		MaxKeys:     5,
	}
}

// KeyRotator handles key rotation
type KeyRotator struct {
	config     KeyRotationConfig
	clients    map[string]*ClientEncryption
	stopChan   chan struct{}
	mu         sync.RWMutex
	isRunning  bool
	lastRotate time.Time
}

// NewKeyRotator creates a new key rotator
func NewKeyRotator(config KeyRotationConfig) *KeyRotator {
	return &KeyRotator{
		config:     config,
		clients:    make(map[string]*ClientEncryption),
		stopChan:   make(chan struct{}),
		isRunning:  false,
		lastRotate: time.Now(),
	}
}

// RegisterClient registers a client for key rotation
func (r *KeyRotator) RegisterClient(clientID string, clientEnc *ClientEncryption) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[clientID] = clientEnc
}

// UnregisterClient unregisters a client from key rotation
func (r *KeyRotator) UnregisterClient(clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, clientID)
}

// Start starts the key rotation process
func (r *KeyRotator) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.config.Enabled {
		return ErrKeyRotationDisabled
	}

	if r.isRunning {
		return nil
	}

	r.isRunning = true
	go r.rotationLoop()
	return nil
}

// Stop stops the key rotation process
func (r *KeyRotator) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isRunning {
		return
	}

	close(r.stopChan)
	r.isRunning = false
}

// ForceRotate forces a key rotation for all clients
func (r *KeyRotator) ForceRotate() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.config.Enabled {
		return ErrKeyRotationDisabled
	}

	for _, clientEnc := range r.clients {
		if err := r.rotateClientKey(clientEnc); err != nil {
			// Log error but continue with other clients
			continue
		}
	}

	r.lastRotate = time.Now()
	return nil
}

// rotationLoop runs the key rotation loop
func (r *KeyRotator) rotationLoop() {
	ticker := time.NewTicker(r.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = r.ForceRotate()
		case <-r.stopChan:
			return
		}
	}
}

// rotateClientKey rotates the key for a client
func (r *KeyRotator) rotateClientKey(clientEnc *ClientEncryption) error {
	encrypter := clientEnc.GetEncrypter()
	if encrypter == nil {
		return nil // No encrypter to rotate
	}

	// Generate a new key
	var newEncrypter clientenc.Encrypter
	var err error

	switch clientEnc.GetEncryptionType() {
	case EncryptionAES:
		// For AES, we'll use the default key size (32 bytes for AES-256)
		_, ok := encrypter.(*clientenc.AESEncrypter)
		if !ok {
			return errors.New("invalid encrypter type")
		}
		keySize := 32 // Use AES-256 by default
		newEncrypter, err = clientenc.NewAESEncrypter(keySize)
	case EncryptionChaCha20:
		newEncrypter, err = clientenc.NewChaCha20Encrypter()
	default:
		return ErrUnsupportedEncryption
	}

	if err != nil {
		return err
	}

	// Set the new encrypter
	return clientEnc.SetEncrypter(newEncrypter)
}

// GetLastRotateTime returns the time of the last key rotation
func (r *KeyRotator) GetLastRotateTime() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastRotate
}

// SetRotationInterval sets the key rotation interval
func (r *KeyRotator) SetRotationInterval(interval time.Duration) error {
	if interval <= 0 {
		return ErrInvalidRotationInterval
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.config.Interval = interval
	return nil
}

// IsEnabled returns whether key rotation is enabled
func (r *KeyRotator) IsEnabled() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config.Enabled
}

// Enable enables key rotation
func (r *KeyRotator) Enable() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config.Enabled = true
}

// Disable disables key rotation
func (r *KeyRotator) Disable() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config.Enabled = false
}
