package encryption

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"sync"
	"time"

	clientenc "github.com/Cl0udRs4/dinot/internal/client/encryption"
)

var (
	// ErrInvalidCurve is returned when an invalid curve is provided
	ErrInvalidCurve = errors.New("invalid elliptic curve")
	// ErrInvalidPublicKey is returned when an invalid public key is provided
	ErrInvalidPublicKey = errors.New("invalid public key")
	// ErrInvalidPrivateKey is returned when an invalid private key is provided
	ErrInvalidPrivateKey = errors.New("invalid private key")
)

// ForwardSecrecyConfig holds forward secrecy configuration
type ForwardSecrecyConfig struct {
	// Enabled indicates whether forward secrecy is enabled
	Enabled bool
	// Curve is the elliptic curve to use (P256, P384, P521)
	Curve string
	// KeyRotationInterval is the interval between key rotations
	KeyRotationInterval time.Duration
}

// DefaultForwardSecrecyConfig returns the default forward secrecy configuration
func DefaultForwardSecrecyConfig() ForwardSecrecyConfig {
	return ForwardSecrecyConfig{
		Enabled:            true,
		Curve:              "P256",
		KeyRotationInterval: 1 * time.Hour,
	}
}

// ForwardSecrecy handles perfect forward secrecy
type ForwardSecrecy struct {
	config      ForwardSecrecyConfig
	curve       ecdh.Curve
	privateKey  *ecdh.PrivateKey
	publicKey   *ecdh.PublicKey
	keyHistory  []*ecdh.PrivateKey
	mu          sync.RWMutex
	stopChan    chan struct{}
	isRunning   bool
	lastRotate  time.Time
}

// NewForwardSecrecy creates a new forward secrecy handler
func NewForwardSecrecy(config ForwardSecrecyConfig) (*ForwardSecrecy, error) {
	var curve ecdh.Curve
	var err error

	switch config.Curve {
	case "P256":
		curve = ecdh.P256()
	case "P384":
		curve = ecdh.P384()
	case "P521":
		curve = ecdh.P521()
	default:
		return nil, ErrInvalidCurve
	}

	fs := &ForwardSecrecy{
		config:     config,
		curve:      curve,
		keyHistory: make([]*ecdh.PrivateKey, 0),
		stopChan:   make(chan struct{}),
		isRunning:  false,
		lastRotate: time.Now(),
	}

	// Generate initial key pair
	if err = fs.generateKeyPair(); err != nil {
		return nil, err
	}

	return fs, nil
}

// Start starts the key rotation process
func (fs *ForwardSecrecy) Start() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.config.Enabled {
		return errors.New("forward secrecy is disabled")
	}

	if fs.isRunning {
		return nil
	}

	fs.isRunning = true
	go fs.rotationLoop()
	return nil
}

// Stop stops the key rotation process
func (fs *ForwardSecrecy) Stop() {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.isRunning {
		return
	}

	close(fs.stopChan)
	fs.isRunning = false
}

// GetPublicKey returns the current public key
func (fs *ForwardSecrecy) GetPublicKey() []byte {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if fs.publicKey == nil {
		return nil
	}

	return fs.publicKey.Bytes()
}

// ComputeSharedSecret computes a shared secret with a peer's public key
func (fs *ForwardSecrecy) ComputeSharedSecret(peerPublicKeyBytes []byte) ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if fs.privateKey == nil {
		return nil, ErrInvalidPrivateKey
	}

	peerPublicKey, err := fs.curve.NewPublicKey(peerPublicKeyBytes)
	if err != nil {
		return nil, ErrInvalidPublicKey
	}

	sharedSecret, err := fs.privateKey.ECDH(peerPublicKey)
	if err != nil {
		return nil, err
	}

	// Hash the shared secret for better security
	hash := sha256.Sum256(sharedSecret)
	return hash[:], nil
}

// TryComputeSharedSecretWithHistory tries to compute a shared secret with a peer's public key using current and historical keys
func (fs *ForwardSecrecy) TryComputeSharedSecretWithHistory(peerPublicKeyBytes []byte) ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// First try with the current key
	if fs.privateKey != nil {
		peerPublicKey, err := fs.curve.NewPublicKey(peerPublicKeyBytes)
		if err == nil {
			sharedSecret, err := fs.privateKey.ECDH(peerPublicKey)
			if err == nil {
				hash := sha256.Sum256(sharedSecret)
				return hash[:], nil
			}
		}
	}

	// Try with historical keys
	for _, privateKey := range fs.keyHistory {
		peerPublicKey, err := fs.curve.NewPublicKey(peerPublicKeyBytes)
		if err != nil {
			continue
		}

		sharedSecret, err := privateKey.ECDH(peerPublicKey)
		if err != nil {
			continue
		}

		hash := sha256.Sum256(sharedSecret)
		return hash[:], nil
	}

	return nil, errors.New("failed to compute shared secret with any key")
}

// ForceRotate forces a key rotation
func (fs *ForwardSecrecy) ForceRotate() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.config.Enabled {
		return errors.New("forward secrecy is disabled")
	}

	// Add current key to history
	if fs.privateKey != nil {
		fs.keyHistory = append(fs.keyHistory, fs.privateKey)
		
		// Limit history size
		if len(fs.keyHistory) > 5 {
			fs.keyHistory = fs.keyHistory[len(fs.keyHistory)-5:]
		}
	}

	// Generate new key pair
	if err := fs.generateKeyPair(); err != nil {
		return err
	}

	fs.lastRotate = time.Now()
	return nil
}

// generateKeyPair generates a new key pair
func (fs *ForwardSecrecy) generateKeyPair() error {
	privateKey, err := fs.curve.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	fs.privateKey = privateKey
	fs.publicKey = privateKey.PublicKey()
	return nil
}

// rotationLoop runs the key rotation loop
func (fs *ForwardSecrecy) rotationLoop() {
	ticker := time.NewTicker(fs.config.KeyRotationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = fs.ForceRotate()
		case <-fs.stopChan:
			return
		}
	}
}

// GetLastRotateTime returns the time of the last key rotation
func (fs *ForwardSecrecy) GetLastRotateTime() time.Time {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.lastRotate
}

// SetKeyRotationInterval sets the key rotation interval
func (fs *ForwardSecrecy) SetKeyRotationInterval(interval time.Duration) error {
	if interval <= 0 {
		return errors.New("invalid rotation interval")
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.config.KeyRotationInterval = interval
	return nil
}

// IsEnabled returns whether forward secrecy is enabled
func (fs *ForwardSecrecy) IsEnabled() bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.config.Enabled
}

// Enable enables forward secrecy
func (fs *ForwardSecrecy) Enable() {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.config.Enabled = true
}

// Disable disables forward secrecy
func (fs *ForwardSecrecy) Disable() {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.config.Enabled = false
}
