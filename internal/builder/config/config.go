package config

import (
    "strings"
)

// BuilderConfig represents the configuration for the Builder tool
type BuilderConfig struct {
    // Required parameters
    Protocols []string
    Domain    string
    Servers   map[string]string
    Modules   []string

    // Optional parameters
    Encryption string
    Debug      bool
    Version    string
    Signature  bool
}

// NewBuilderConfig creates a new BuilderConfig with default values
func NewBuilderConfig() *BuilderConfig {
    return &BuilderConfig{
        Protocols:  []string{},
        Domain:     "",
        Servers:    make(map[string]string),
        Modules:    []string{},
        Encryption: "aes",
        Debug:      false,
        Version:    "1.0.0",
        Signature:  false,
    }
}

// ParseProtocols parses the protocol string into a slice of protocols
func ParseProtocols(protocolStr string) []string {
    if protocolStr == "" {
        return []string{}
    }
    return strings.Split(protocolStr, ",")
}

// ParseServers parses the server string into a map of protocol to server address
func ParseServers(serverStr string) map[string]string {
    servers := make(map[string]string)
    if serverStr == "" {
        return servers
    }

    serverPairs := strings.Split(serverStr, ",")
    for _, pair := range serverPairs {
        parts := strings.SplitN(pair, ":", 2)
        if len(parts) == 2 {
            protocol := parts[0]
            address := parts[1]
            servers[protocol] = address
        }
    }

    return servers
}

// ParseModules parses the module string into a slice of modules
func ParseModules(moduleStr string) []string {
    if moduleStr == "" {
        return []string{}
    }
    return strings.Split(moduleStr, ",")
}
