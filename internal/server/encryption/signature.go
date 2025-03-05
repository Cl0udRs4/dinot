package encryption

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	// ErrInvalidSignature is returned when a signature is invalid
	ErrInvalidSignature = errors.New("invalid signature")
	// ErrInvalidKey is returned when a key is invalid
	ErrInvalidKeyFormat = errors.New("invalid key format")
	// ErrKeyNotFound is returned when a key is not found
	ErrKeyNotFound = errors.New("key not found")
)

// SignatureVerifier handles signature verification
type SignatureVerifier struct {
	publicKeys map[string]*rsa.PublicKey
	privateKey *rsa.PrivateKey
}

// NewSignatureVerifier creates a new signature verifier
func NewSignatureVerifier() *SignatureVerifier {
	return &SignatureVerifier{
		publicKeys: make(map[string]*rsa.PublicKey),
	}
}

// LoadPrivateKeyFromFile loads a private key from a file
func (v *SignatureVerifier) LoadPrivateKeyFromFile(filePath string) error {
	keyData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return ErrInvalidKeyFormat
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	v.privateKey = privateKey
	return nil
}

// LoadPublicKeyFromFile loads a public key from a file
func (v *SignatureVerifier) LoadPublicKeyFromFile(name, filePath string) error {
	keyData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return ErrInvalidKeyFormat
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return err
	}

	v.publicKeys[name] = publicKey
	return nil
}

// LoadPublicKeysFromDirectory loads all public keys from a directory
func (v *SignatureVerifier) LoadPublicKeysFromDirectory(dirPath string) error {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) != ".pem" {
			continue
		}

		name := filepath.Base(file.Name())
		name = name[:len(name)-4] // Remove .pem extension

		err := v.LoadPublicKeyFromFile(name, filepath.Join(dirPath, file.Name()))
		if err != nil {
			// Log error but continue with other keys
			continue
		}
	}

	return nil
}

// GenerateKeyPair generates a new RSA key pair
func (v *SignatureVerifier) GenerateKeyPair(bits int) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}

	v.privateKey = privateKey
	return nil
}

// SavePrivateKeyToFile saves the private key to a file
func (v *SignatureVerifier) SavePrivateKeyToFile(filePath string) error {
	if v.privateKey == nil {
		return errors.New("no private key available")
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(v.privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, block)
}

// SavePublicKeyToFile saves the public key to a file
func (v *SignatureVerifier) SavePublicKeyToFile(filePath string) error {
	if v.privateKey == nil {
		return errors.New("no private key available")
	}

	publicKey := &v.privateKey.PublicKey
	keyBytes := x509.MarshalPKCS1PublicKey(publicKey)
	block := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: keyBytes,
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, block)
}

// Sign signs data with the private key
func (v *SignatureVerifier) Sign(data []byte) ([]byte, error) {
	if v.privateKey == nil {
		return nil, errors.New("no private key available")
	}

	hashed := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, v.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// Verify verifies a signature with a public key
func (v *SignatureVerifier) Verify(keyName string, data, signature []byte) error {
	publicKey, ok := v.publicKeys[keyName]
	if !ok {
		return ErrKeyNotFound
	}

	hashed := sha256.Sum256(data)
	err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		return ErrInvalidSignature
	}

	return nil
}

// VerifyWithKey verifies a signature with a specific public key
func (v *SignatureVerifier) VerifyWithKey(publicKey *rsa.PublicKey, data, signature []byte) error {
	if publicKey == nil {
		return errors.New("no public key provided")
	}

	hashed := sha256.Sum256(data)
	err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		return ErrInvalidSignature
	}

	return nil
}

// GetPublicKey returns a public key by name
func (v *SignatureVerifier) GetPublicKey(name string) (*rsa.PublicKey, error) {
	publicKey, ok := v.publicKeys[name]
	if !ok {
		return nil, ErrKeyNotFound
	}

	return publicKey, nil
}

// GetPublicKeyFromPrivate returns the public key from the private key
func (v *SignatureVerifier) GetPublicKeyFromPrivate() (*rsa.PublicKey, error) {
	if v.privateKey == nil {
		return nil, errors.New("no private key available")
	}

	return &v.privateKey.PublicKey, nil
}
