package template

import (
	"bytes"
	"testing"
)

func TestGenerateClientCode(t *testing.T) {
	// Test case 1: Basic configuration
	params1 := map[string]interface{}{
		"ClientID":          "test-client",
		"Protocols":         []string{"tcp", "udp"},
		"Domain":            "",
		"Servers":           map[string]string{"tcp": "localhost:8080", "udp": "localhost:8081"},
		"Modules":           []string{"shell"},
		"Encryption":        "aes",
		"Debug":             false,
		"Version":           "1.0.0",
		"BuildTime":         "2025-03-05T10:00:00Z",
		"HeartbeatInterval": "time.Second * 60",
		"Signature":         "",
	}

	code1, err := GenerateClientCode(params1)
	if err != nil {
		t.Fatalf("Failed to generate client code: %v", err)
	}

	// Verify that the code contains the expected configuration
	expectedStrings1 := []string{
		`Protocols:         []string{ "tcp", "udp",  }`,
		`Servers:           map[string]string{ "tcp": "localhost:8080", "udp": "localhost:8081",  }`,
		`"github.com/Cl0udRs4/dinot/internal/client/module/shell"`,
		`Encryption:        "aes"`,
		`Debug:             false`,
		`Version   = "1.0.0"`,
		`BuildTime = "2025-03-05T10:00:00Z"`,
	}

	for _, expected := range expectedStrings1 {
		if !bytes.Contains(code1, []byte(expected)) {
			t.Errorf("Client code does not contain expected string: %s", expected)
		}
	}

	// Test case 2: DNS protocol with domain
	params2 := map[string]interface{}{
		"ClientID":          "test-client",
		"Protocols":         []string{"dns"},
		"Domain":            "example.com",
		"Servers":           map[string]string{"dns": "8.8.8.8"},
		"Modules":           []string{"shell"},
		"Encryption":        "chacha20",
		"Debug":             true,
		"Version":           "1.0.0",
		"BuildTime":         "2025-03-05T10:00:00Z",
		"HeartbeatInterval": "time.Second * 60",
		"Signature":         "test-signature",
	}

	code2, err := GenerateClientCode(params2)
	if err != nil {
		t.Fatalf("Failed to generate client code: %v", err)
	}

	// Verify that the code contains the expected configuration
	expectedStrings2 := []string{
		`Protocols:         []string{ "dns",  }`,
		`Domain:            "example.com"`,
		`Servers:           map[string]string{ "dns": "8.8.8.8",  }`,
		`Encryption:        "chacha20"`,
		`Debug:             true`,
		`Signature = "test-signature"`,
	}

	for _, expected := range expectedStrings2 {
		if !bytes.Contains(code2, []byte(expected)) {
			t.Errorf("Client code does not contain expected string: %s", expected)
		}
	}
}
