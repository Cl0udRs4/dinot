package signature

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	// ErrSignatureVerificationFailed is returned when signature verification fails
	ErrSignatureVerificationFailed = errors.New("signature verification failed")

	// ErrKeyNotFound is returned when the key file is not found
	ErrKeyNotFound = errors.New("key file not found")
)

// SignatureManager manages client code signing and verification
type SignatureManager struct {
	privateKeyPath string
	publicKeyPath  string
	privateKey     *rsa.PrivateKey
	publicKey      *rsa.PublicKey
}

// NewSignatureManager creates a new signature manager
func NewSignatureManager(privateKeyPath, publicKeyPath string) *SignatureManager {
	return &SignatureManager{
		privateKeyPath: privateKeyPath,
		publicKeyPath:  publicKeyPath,
	}
}

// LoadKeys loads the private and public keys
func (m *SignatureManager) LoadKeys() error {
	// Load private key if path is provided
	if m.privateKeyPath != "" {
		privateKeyBytes, err := ioutil.ReadFile(m.privateKeyPath)
		if err != nil {
			if os.IsNotExist(err) {
				return ErrKeyNotFound
			}
			return err
		}

		block, _ := pem.Decode(privateKeyBytes)
		if block == nil {
			return fmt.Errorf("failed to decode PEM block containing private key")
		}

		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return err
		}

		m.privateKey = privateKey
	}

	// Load public key if path is provided
	if m.publicKeyPath != "" {
		publicKeyBytes, err := ioutil.ReadFile(m.publicKeyPath)
		if err != nil {
			if os.IsNotExist(err) {
				return ErrKeyNotFound
			}
			return err
		}

		block, _ := pem.Decode(publicKeyBytes)
		if block == nil {
			return fmt.Errorf("failed to decode PEM block containing public key")
		}

		publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return err
		}

		m.publicKey = publicKey
	}

	return nil
}

// GenerateKeyPair generates a new RSA key pair
func (m *SignatureManager) GenerateKeyPair(bits int) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}

	m.privateKey = privateKey
	m.publicKey = &privateKey.PublicKey

	// Create directories if they don't exist
	if m.privateKeyPath != "" {
		if err := os.MkdirAll(filepath.Dir(m.privateKeyPath), 0755); err != nil {
			return err
		}
	}

	if m.publicKeyPath != "" {
		if err := os.MkdirAll(filepath.Dir(m.publicKeyPath), 0755); err != nil {
			return err
		}
	}

	// Save private key
	if m.privateKeyPath != "" {
		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		privateKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		})

		if err := ioutil.WriteFile(m.privateKeyPath, privateKeyPEM, 0600); err != nil {
			return err
		}
	}

	// Save public key
	if m.publicKeyPath != "" {
		publicKeyBytes := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
		publicKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: publicKeyBytes,
		})

		if err := ioutil.WriteFile(m.publicKeyPath, publicKeyPEM, 0644); err != nil {
			return err
		}
	}

	return nil
}

// SignCode signs the provided code
func (m *SignatureManager) SignCode(code []byte) (string, error) {
	if m.privateKey == nil {
		return "", fmt.Errorf("private key not loaded")
	}

	hashed := sha256.Sum256(code)
	signature, err := rsa.SignPKCS1v15(rand.Reader, m.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignature verifies the signature of the provided code
func (m *SignatureManager) VerifySignature(code []byte, signatureStr string) error {
	if m.publicKey == nil {
		return fmt.Errorf("public key not loaded")
	}

	signature, err := base64.StdEncoding.DecodeString(signatureStr)
	if err != nil {
		return err
	}

	hashed := sha256.Sum256(code)
	err = rsa.VerifyPKCS1v15(m.publicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		return ErrSignatureVerificationFailed
	}

	return nil
}
