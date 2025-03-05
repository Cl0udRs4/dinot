package listener

import (
	"testing"
)

func TestDNSListener_GetProtocol(t *testing.T) {
	// Create a test config
	config := DNSConfig{
		Config: Config{
			Address:        "127.0.0.1:53",
			BufferSize:     1024,
			MaxConnections: 10,
			Timeout:        30,
		},
		Domain:      "example.com",
		TTL:         60,
		RecordTypes: []string{"A", "TXT"},
	}

	// Create a DNS listener
	listener := NewDNSListener(config)

	// Test protocol
	if listener.GetProtocol() != "dns" {
		t.Errorf("Expected protocol \"dns\", got \"%s\"", listener.GetProtocol())
	}

	// Test config
	if listener.GetConfig().Address != config.Config.Address {
		t.Errorf("Expected address \"%s\", got \"%s\"", config.Config.Address, listener.GetConfig().Address)
	}
}
