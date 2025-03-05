package generator

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Cl0udRs4/dinot/internal/builder/config"
	"github.com/Cl0udRs4/dinot/internal/builder/signature"
)

func TestGenerator(t *testing.T) {
	// Create a temporary output directory
	outputDir, err := os.MkdirTemp("", "generator-test-")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(outputDir)

	// Create a test configuration
	cfg := &config.BuilderConfig{
		Protocols:  []string{"tcp", "udp"},
		Domain:     "",
		Servers:    map[string]string{"tcp": "localhost:8080", "udp": "localhost:8081"},
		Modules:    []string{"shell"},
		Encryption: "aes",
		Debug:      true,
		Version:    "1.0.0-test",
		Signature:  false,
	}

	// Create a generator
	generator := NewGenerator(cfg, outputDir)

	// Generate client code
	clientCode, err := generator.generateClientCode()
	if err != nil {
		t.Fatalf("Failed to generate client code: %v", err)
	}

	// Print the client code for debugging
	t.Logf("Generated client code: %s", string(clientCode))

	// Verify that the client code contains the expected configuration
	expectedStrings := []string{
		`"tcp"`,
		`"udp"`,
		`"localhost:8080"`,
		`"localhost:8081"`,
		`"github.com/Cl0udRs4/dinot/internal/client/module/shell"`,
		`"aes"`,
		`Debug:             true`,
		`Version   = "1.0.0-test"`,
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains(clientCode, []byte(expected)) {
			t.Errorf("Client code does not contain expected string: %s", expected)
		}
	}
}

func TestGeneratorWithSignature(t *testing.T) {
	// Create a temporary output directory
	outputDir, err := os.MkdirTemp("", "generator-test-")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(outputDir)

	// Create a temporary keys directory
	keysDir, err := os.MkdirTemp("", "keys-test-")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(keysDir)

	privateKeyPath := filepath.Join(keysDir, "private.pem")
	publicKeyPath := filepath.Join(keysDir, "public.pem")

	// Create a signature manager
	sm := signature.NewSignatureManager(privateKeyPath, publicKeyPath)
	if err := sm.GenerateKeyPair(2048); err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create a test configuration
	cfg := &config.BuilderConfig{
		Protocols:  []string{"tcp"},
		Domain:     "",
		Servers:    map[string]string{"tcp": "localhost:8080"},
		Modules:    []string{"shell"},
		Encryption: "aes",
		Debug:      false,
		Version:    "1.0.0-test",
		Signature:  true,
	}

	// Create a generator
	generator := NewGenerator(cfg, outputDir)
	generator.SetSignatureManager(sm)

	// Generate client code
	clientCode, err := generator.generateClientCode()
	if err != nil {
		t.Fatalf("Failed to generate client code: %v", err)
	}

	// Sign the client code
	signature, err := sm.SignCode(clientCode)
	if err != nil {
		t.Fatalf("Failed to sign client code: %v", err)
	}

	// Verify the signature
	if err := sm.VerifySignature(clientCode, signature); err != nil {
		t.Fatalf("Signature verification failed: %v", err)
	}
}
