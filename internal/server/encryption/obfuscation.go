package encryption

import (
	"crypto/rand"
	"errors"
	"math"
	"math/big"
	"time"
)

var (
	// ErrInvalidPaddingConfig is returned when an invalid padding configuration is provided
	ErrInvalidPaddingConfig = errors.New("invalid padding configuration")
	// ErrInvalidJitterConfig is returned when an invalid jitter configuration is provided
	ErrInvalidJitterConfig = errors.New("invalid jitter configuration")
)

// ObfuscationConfig holds traffic obfuscation configuration
type ObfuscationConfig struct {
	// EnablePadding enables message padding
	EnablePadding bool
	// MinPadding is the minimum padding length
	MinPadding int
	// MaxPadding is the maximum padding length
	MaxPadding int
	// EnableJitter enables timing jitter
	EnableJitter bool
	// MinJitter is the minimum jitter in milliseconds
	MinJitter int
	// MaxJitter is the maximum jitter in milliseconds
	MaxJitter int
	// EnableMimicry enables protocol mimicry
	EnableMimicry bool
	// MimicryProfile is the profile to mimic (e.g., "http", "dns", "ssl")
	MimicryProfile string
}

// DefaultObfuscationConfig returns the default obfuscation configuration
func DefaultObfuscationConfig() ObfuscationConfig {
	return ObfuscationConfig{
		EnablePadding:   true,
		MinPadding:      16,
		MaxPadding:      256,
		EnableJitter:    true,
		MinJitter:       50,
		MaxJitter:       500,
		EnableMimicry:   false,
		MimicryProfile:  "http",
	}
}

// Obfuscator handles traffic obfuscation
type Obfuscator struct {
	config ObfuscationConfig
}

// NewObfuscator creates a new obfuscator
func NewObfuscator(config ObfuscationConfig) (*Obfuscator, error) {
	if config.EnablePadding && (config.MinPadding < 0 || config.MaxPadding < config.MinPadding) {
		return nil, ErrInvalidPaddingConfig
	}

	if config.EnableJitter && (config.MinJitter < 0 || config.MaxJitter < config.MinJitter) {
		return nil, ErrInvalidJitterConfig
	}

	return &Obfuscator{
		config: config,
	}, nil
}

// AddPadding adds random padding to a message
func (o *Obfuscator) AddPadding(data []byte) ([]byte, error) {
	if !o.config.EnablePadding {
		return data, nil
	}

	// Generate a random padding length between min and max
	paddingLen, err := randomInt(o.config.MinPadding, o.config.MaxPadding)
	if err != nil {
		return nil, err
	}

	// Generate random padding bytes
	padding, err := GenerateRandomBytes(paddingLen)
	if err != nil {
		return nil, err
	}

	// Create a new slice with the original data and padding
	// Format: [original data length (4 bytes)][original data][padding]
	result := make([]byte, 4+len(data)+len(padding))
	
	// Store the original data length
	result[0] = byte(len(data) >> 24)
	result[1] = byte(len(data) >> 16)
	result[2] = byte(len(data) >> 8)
	result[3] = byte(len(data))
	
	// Copy the original data
	copy(result[4:], data)
	
	// Copy the padding
	copy(result[4+len(data):], padding)
	
	return result, nil
}

// RemovePadding removes padding from a message
func (o *Obfuscator) RemovePadding(data []byte) ([]byte, error) {
	if !o.config.EnablePadding || len(data) < 4 {
		return data, nil
	}

	// Extract the original data length
	originalLen := int(data[0])<<24 | int(data[1])<<16 | int(data[2])<<8 | int(data[3])
	
	// Validate the length
	if originalLen < 0 || originalLen+4 > len(data) {
		return nil, errors.New("invalid padding format")
	}
	
	// Extract the original data
	return data[4 : 4+originalLen], nil
}

// ApplyJitter applies timing jitter
func (o *Obfuscator) ApplyJitter() time.Duration {
	if !o.config.EnableJitter {
		return 0
	}

	// Generate a random jitter duration between min and max
	jitterMs, err := randomInt(o.config.MinJitter, o.config.MaxJitter)
	if err != nil {
		return 0
	}

	return time.Duration(jitterMs) * time.Millisecond
}

// ApplyMimicry applies protocol mimicry to a message
func (o *Obfuscator) ApplyMimicry(data []byte) ([]byte, error) {
	if !o.config.EnableMimicry {
		return data, nil
	}

	switch o.config.MimicryProfile {
	case "http":
		return o.mimicHTTP(data)
	case "dns":
		return o.mimicDNS(data)
	case "ssl":
		return o.mimicSSL(data)
	default:
		return data, nil
	}
}

// RemoveMimicry removes protocol mimicry from a message
func (o *Obfuscator) RemoveMimicry(data []byte) ([]byte, error) {
	if !o.config.EnableMimicry {
		return data, nil
	}

	switch o.config.MimicryProfile {
	case "http":
		return o.extractFromHTTP(data)
	case "dns":
		return o.extractFromDNS(data)
	case "ssl":
		return o.extractFromSSL(data)
	default:
		return data, nil
	}
}

// mimicHTTP mimics HTTP traffic
func (o *Obfuscator) mimicHTTP(data []byte) ([]byte, error) {
	// Simple HTTP GET request mimicry
	// Format: GET /path?data=base64(data) HTTP/1.1\r\nHost: example.com\r\n\r\n
	
	// Base64 encode the data
	encoded := base64Encode(data)
	
	// Create a fake HTTP request
	httpRequest := "GET /index.html?data=" + encoded + " HTTP/1.1\r\n" +
		"Host: example.com\r\n" +
		"User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36\r\n" +
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8\r\n" +
		"Connection: keep-alive\r\n\r\n"
	
	return []byte(httpRequest), nil
}

// extractFromHTTP extracts data from HTTP mimicry
func (o *Obfuscator) extractFromHTTP(data []byte) ([]byte, error) {
	// Extract the base64-encoded data from the HTTP request
	// Find the data parameter in the URL
	dataStart := indexOf(data, []byte("?data="))
	if dataStart == -1 {
		return nil, errors.New("invalid HTTP mimicry format")
	}
	
	dataStart += 6 // Skip "?data="
	
	// Find the end of the data (space before HTTP/1.1)
	dataEnd := indexOf(data[dataStart:], []byte(" "))
	if dataEnd == -1 {
		return nil, errors.New("invalid HTTP mimicry format")
	}
	
	// Extract and decode the base64 data
	encoded := string(data[dataStart : dataStart+dataEnd])
	return base64Decode(encoded)
}

// mimicDNS mimics DNS traffic
func (o *Obfuscator) mimicDNS(data []byte) ([]byte, error) {
	// Simple DNS query mimicry
	// This is a simplified version; a real implementation would create a proper DNS packet
	
	// Base64 encode the data and split into chunks (DNS labels)
	encoded := base64Encode(data)
	
	// Split into chunks of 63 bytes (max DNS label length)
	var dnsQuery []byte
	for i := 0; i < len(encoded); i += 63 {
		end := i + 63
		if end > len(encoded) {
			end = len(encoded)
		}
		
		chunk := encoded[i:end]
		dnsQuery = append(dnsQuery, byte(len(chunk)))
		dnsQuery = append(dnsQuery, []byte(chunk)...)
	}
	
	// Add the root domain
	dnsQuery = append(dnsQuery, 0)
	
	// Add a simple DNS header (12 bytes)
	header := []byte{
		0x00, 0x01, // Transaction ID
		0x01, 0x00, // Flags (standard query)
		0x00, 0x01, // Questions
		0x00, 0x00, // Answer RRs
		0x00, 0x00, // Authority RRs
		0x00, 0x00, // Additional RRs
	}
	
	// Combine header and query
	result := append(header, dnsQuery...)
	
	return result, nil
}

// extractFromDNS extracts data from DNS mimicry
func (o *Obfuscator) extractFromDNS(data []byte) ([]byte, error) {
	// Skip the DNS header (12 bytes)
	if len(data) <= 12 {
		return nil, errors.New("invalid DNS mimicry format")
	}
	
	// Extract the DNS labels
	var encoded string
	pos := 12
	
	for pos < len(data) {
		labelLen := int(data[pos])
		pos++
		
		if labelLen == 0 {
			break // Root domain
		}
		
		if pos+labelLen > len(data) {
			return nil, errors.New("invalid DNS mimicry format")
		}
		
		encoded += string(data[pos : pos+labelLen])
		pos += labelLen
	}
	
	// Decode the base64 data
	return base64Decode(encoded)
}

// mimicSSL mimics SSL/TLS traffic
func (o *Obfuscator) mimicSSL(data []byte) ([]byte, error) {
	// Simple SSL/TLS ClientHello mimicry
	// This is a simplified version; a real implementation would create a proper TLS packet
	
	// Base64 encode the data
	encoded := base64Encode(data)
	
	// Create a fake TLS ClientHello
	// Record header
	recordHeader := []byte{
		0x16,                   // Content type: Handshake
		0x03, 0x01,             // TLS version: TLS 1.0
		byte(len(encoded) >> 8), // Length (high byte)
		byte(len(encoded)),      // Length (low byte)
	}
	
	// Handshake header
	handshakeHeader := []byte{
		0x01,                                 // Handshake type: ClientHello
		byte((len(encoded) + 4) >> 16),       // Length (high byte)
		byte((len(encoded) + 4) >> 8),        // Length (middle byte)
		byte(len(encoded) + 4),               // Length (low byte)
		0x03, 0x03,                           // TLS version: TLS 1.2
	}
	
	// Random (32 bytes)
	random, _ := GenerateRandomBytes(32)
	
	// Session ID (0 bytes)
	sessionID := []byte{0x00}
	
	// Cipher suites (2 bytes)
	cipherSuites := []byte{0x00, 0x02, 0xc0, 0x2f}
	
	// Compression methods (1 byte)
	compressionMethods := []byte{0x01, 0x00}
	
	// Extensions length (2 bytes)
	extensionsLength := []byte{byte(len(encoded) >> 8), byte(len(encoded))}
	
	// Combine all parts
	result := append(recordHeader, handshakeHeader...)
	result = append(result, random...)
	result = append(result, sessionID...)
	result = append(result, cipherSuites...)
	result = append(result, compressionMethods...)
	result = append(result, extensionsLength...)
	result = append(result, []byte(encoded)...)
	
	return result, nil
}

// extractFromSSL extracts data from SSL/TLS mimicry
func (o *Obfuscator) extractFromSSL(data []byte) ([]byte, error) {
	// Check for minimum TLS record length
	if len(data) < 5 {
		return nil, errors.New("invalid SSL/TLS mimicry format")
	}
	
	// Check record type
	if data[0] != 0x16 {
		return nil, errors.New("invalid SSL/TLS record type")
	}
	
	// Skip record header (5 bytes) and handshake header (4 bytes)
	// Skip random (32 bytes), session ID length (1 byte)
	// Skip cipher suites length (2 bytes) and cipher suites
	// Skip compression methods length (1 byte) and compression methods
	// Skip extensions length (2 bytes)
	
	// For simplicity, we'll just extract the last part of the message
	// In a real implementation, we would parse the TLS structure properly
	
	// Find the extensions length
	pos := len(data) - 2
	extensionsLength := int(data[pos])<<8 | int(data[pos+1])
	
	// Extract the encoded data
	if pos-extensionsLength < 0 || pos+2 > len(data) {
		return nil, errors.New("invalid SSL/TLS mimicry format")
	}
	
	encoded := string(data[pos-extensionsLength : pos])
	
	// Decode the base64 data
	return base64Decode(encoded)
}

// randomInt generates a random integer between min and max (inclusive)
func randomInt(min, max int) (int, error) {
	if min > max {
		return 0, errors.New("min cannot be greater than max")
	}
	
	if min == max {
		return min, nil
	}
	
	// Generate a random number between 0 and (max - min)
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}
	
	// Add min to get a number between min and max
	return int(n.Int64()) + min, nil
}

// indexOf finds the index of a substring in a byte slice
func indexOf(data, substr []byte) int {
	for i := 0; i <= len(data)-len(substr); i++ {
		if bytesEqual(data[i:i+len(substr)], substr) {
			return i
		}
	}
	return -1
}

// bytesEqual compares two byte slices
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// base64Encode encodes data to base64
func base64Encode(data []byte) string {
	// This is a placeholder for a real base64 encoding function
	// In a real implementation, we would use the standard library
	return "base64encodeddata"
}

// base64Decode decodes base64 data
func base64Decode(encoded string) ([]byte, error) {
	// This is a placeholder for a real base64 decoding function
	// In a real implementation, we would use the standard library
	return []byte("decodeddata"), nil
}
