package validation

import (
    "fmt"

    "github.com/Cl0udRs4/dinot/internal/builder/config"
)

// ValidateConfig validates the BuilderConfig
func ValidateConfig(cfg *config.BuilderConfig) error {
    // Validate required parameters
    if len(cfg.Protocols) == 0 {
        return fmt.Errorf("protocol is required")
    }

    // Validate protocols
    for _, protocol := range cfg.Protocols {
        if !isValidProtocol(protocol) {
            return fmt.Errorf("invalid protocol: %s", protocol)
        }
    }

    // Validate domain if DNS protocol is specified
    if contains(cfg.Protocols, "dns") && cfg.Domain == "" {
        return fmt.Errorf("domain is required for DNS protocol")
    }

    // Validate servers
    if len(cfg.Servers) == 0 {
        return fmt.Errorf("servers are required")
    }

    // Validate server addresses for each protocol
    for _, protocol := range cfg.Protocols {
        if _, ok := cfg.Servers[protocol]; !ok {
            return fmt.Errorf("server address for protocol %s is required", protocol)
        }
    }

    // Validate modules
    if len(cfg.Modules) == 0 {
        return fmt.Errorf("modules are required")
    }

    // Validate encryption
    if !isValidEncryption(cfg.Encryption) {
        return fmt.Errorf("invalid encryption: %s", cfg.Encryption)
    }

    return nil
}

// isValidProtocol checks if the protocol is valid
func isValidProtocol(protocol string) bool {
    validProtocols := []string{"tcp", "udp", "ws", "icmp", "dns"}
    return contains(validProtocols, protocol)
}

// isValidEncryption checks if the encryption is valid
func isValidEncryption(encType string) bool {
    return encType == "none" ||
        encType == "aes" ||
        encType == "chacha20"
}

// contains checks if a string is in a slice
func contains(slice []string, str string) bool {
    for _, s := range slice {
        if s == str {
            return true
        }
    }
    return false
}
