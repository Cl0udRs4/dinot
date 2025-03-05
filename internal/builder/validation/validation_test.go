package validation

import (
    "testing"

    "github.com/Cl0udRs4/dinot/internal/builder/config"
)

func TestValidateConfig(t *testing.T) {
    tests := []struct {
        name    string
        cfg     *config.BuilderConfig
        wantErr bool
    }{
        {
            name: "Valid config",
            cfg: &config.BuilderConfig{
                Protocols:  []string{"tcp", "udp"},
                Domain:     "",
                Servers:    map[string]string{"tcp": "localhost:8080", "udp": "localhost:8081"},
                Modules:    []string{"shell"},
                Encryption: "aes",
                Debug:      false,
                Version:    "1.0.0",
                Signature:  false,
            },
            wantErr: false,
        },
        {
            name: "Empty protocols",
            cfg: &config.BuilderConfig{
                Protocols:  []string{},
                Domain:     "",
                Servers:    map[string]string{"tcp": "localhost:8080"},
                Modules:    []string{"shell"},
                Encryption: "aes",
                Debug:      false,
                Version:    "1.0.0",
                Signature:  false,
            },
            wantErr: true,
        },
        {
            name: "Invalid protocol",
            cfg: &config.BuilderConfig{
                Protocols:  []string{"invalid"},
                Domain:     "",
                Servers:    map[string]string{"invalid": "localhost:8080"},
                Modules:    []string{"shell"},
                Encryption: "aes",
                Debug:      false,
                Version:    "1.0.0",
                Signature:  false,
            },
            wantErr: true,
        },
        {
            name: "DNS protocol without domain",
            cfg: &config.BuilderConfig{
                Protocols:  []string{"dns"},
                Domain:     "",
                Servers:    map[string]string{"dns": "localhost:8080"},
                Modules:    []string{"shell"},
                Encryption: "aes",
                Debug:      false,
                Version:    "1.0.0",
                Signature:  false,
            },
            wantErr: true,
        },
        {
            name: "Empty servers",
            cfg: &config.BuilderConfig{
                Protocols:  []string{"tcp"},
                Domain:     "",
                Servers:    map[string]string{},
                Modules:    []string{"shell"},
                Encryption: "aes",
                Debug:      false,
                Version:    "1.0.0",
                Signature:  false,
            },
            wantErr: true,
        },
        {
            name: "Missing server for protocol",
            cfg: &config.BuilderConfig{
                Protocols:  []string{"tcp", "udp"},
                Domain:     "",
                Servers:    map[string]string{"tcp": "localhost:8080"},
                Modules:    []string{"shell"},
                Encryption: "aes",
                Debug:      false,
                Version:    "1.0.0",
                Signature:  false,
            },
            wantErr: true,
        },
        {
            name: "Empty modules",
            cfg: &config.BuilderConfig{
                Protocols:  []string{"tcp"},
                Domain:     "",
                Servers:    map[string]string{"tcp": "localhost:8080"},
                Modules:    []string{},
                Encryption: "aes",
                Debug:      false,
                Version:    "1.0.0",
                Signature:  false,
            },
            wantErr: true,
        },
        {
            name: "Invalid encryption",
            cfg: &config.BuilderConfig{
                Protocols:  []string{"tcp"},
                Domain:     "",
                Servers:    map[string]string{"tcp": "localhost:8080"},
                Modules:    []string{"shell"},
                Encryption: "invalid",
                Debug:      false,
                Version:    "1.0.0",
                Signature:  false,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateConfig(tt.cfg)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestIsValidProtocol(t *testing.T) {
    tests := []struct {
        name     string
        protocol string
        want     bool
    }{
        {
            name:     "Valid protocol - tcp",
            protocol: "tcp",
            want:     true,
        },
        {
            name:     "Valid protocol - udp",
            protocol: "udp",
            want:     true,
        },
        {
            name:     "Valid protocol - ws",
            protocol: "ws",
            want:     true,
        },
        {
            name:     "Valid protocol - icmp",
            protocol: "icmp",
            want:     true,
        },
        {
            name:     "Valid protocol - dns",
            protocol: "dns",
            want:     true,
        },
        {
            name:     "Invalid protocol",
            protocol: "invalid",
            want:     false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := isValidProtocol(tt.protocol)
            if got != tt.want {
                t.Errorf("isValidProtocol() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestIsValidEncryption(t *testing.T) {
    tests := []struct {
        name    string
        encType string
        want    bool
    }{
        {
            name:    "Valid encryption - none",
            encType: "none",
            want:    true,
        },
        {
            name:    "Valid encryption - aes",
            encType: "aes",
            want:    true,
        },
        {
            name:    "Valid encryption - chacha20",
            encType: "chacha20",
            want:    true,
        },
        {
            name:    "Invalid encryption",
            encType: "invalid",
            want:    false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := isValidEncryption(tt.encType)
            if got != tt.want {
                t.Errorf("isValidEncryption() = %v, want %v", got, tt.want)
            }
        })
    }
}
