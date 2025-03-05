package protocol

import (
	"context"
	"encoding/base64"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DNSProtocol implements the Protocol interface for DNS
type DNSProtocol struct {
	*BaseProtocol
	client      *dns.Client
	domain      string
	queryType   uint16
	maxDataSize int
	timeout     time.Duration
}

// NewDNSProtocol creates a new DNS protocol instance
func NewDNSProtocol(config Config) *DNSProtocol {
	return &DNSProtocol{
		BaseProtocol: NewBaseProtocol("dns", config),
		client:       &dns.Client{},
		queryType:    dns.TypeTXT,
		maxDataSize:  250, // Maximum size for TXT record data chunks
	}
}

// Connect establishes a DNS connection to the server
func (d *DNSProtocol) Connect(ctx context.Context) error {
	if d.IsConnected() {
		return ErrAlreadyConnected
	}

	if err := d.ValidateConfig(); err != nil {
		return err
	}

	// Create a context with cancel function
	ctx, d.cancel = context.WithCancel(ctx)

	// Parse the server address to extract domain and DNS server
	parts := strings.Split(d.Config.ServerAddress, "@")
	if len(parts) != 2 {
		return NewClientError(ErrTypeConfiguration, "invalid DNS server address format, expected 'domain@server'", nil)
	}

	d.domain = parts[0]
	dnsServer := parts[1]

	// Set timeout
	d.timeout = time.Duration(d.Config.ConnectTimeout) * time.Second
	d.client.Timeout = d.timeout

	// Try to connect with retry logic
	var retryCount int

	for retryCount <= d.Config.RetryCount {
		// Check if the context is done
		select {
		case <-ctx.Done():
			d.setStatus(StatusDisconnected)
			d.setLastError(ctx.Err())
			return NewClientError(ErrTypeConnection, "connection cancelled", ctx.Err())
		default:
			// Continue connecting
		}

		// Test the connection by sending a simple query
		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn(d.domain), dns.TypeA)
		m.RecursionDesired = true

		_, _, err := d.client.Exchange(m, dnsServer)
		if err == nil {
			break // Connection successful
		}

		// If this was the last retry, return the error
		if retryCount == d.Config.RetryCount {
			d.setStatus(StatusDisconnected)
			d.setLastError(err)
			return NewClientError(ErrTypeConnection, "failed to connect after retries", err)
		}

		// Wait before retrying
		retryCount++
		select {
		case <-ctx.Done():
			d.setStatus(StatusDisconnected)
			d.setLastError(ctx.Err())
			return NewClientError(ErrTypeConnection, "connection cancelled during retry", ctx.Err())
		case <-time.After(time.Duration(d.Config.RetryInterval) * time.Second):
			// Continue to next retry
		}
	}

	// Store the connection details
	d.conn = nil // DNS doesn't use the standard net.Conn
	d.setStatus(StatusConnected)
	d.setLastError(nil)

	return nil
}

// Disconnect closes the DNS connection
func (d *DNSProtocol) Disconnect() error {
	if !d.IsConnected() {
		return nil // Already disconnected
	}

	// DNS doesn't have a persistent connection to close
	// Just call the base Disconnect method to cancel the context and update status
	d.BaseProtocol.Disconnect()
	return nil
}

// Send sends data over the DNS connection
func (d *DNSProtocol) Send(data []byte) (int, error) {
	if !d.IsConnected() {
		return 0, ErrNotConnected
	}

	// Get the DNS server from the server address
	parts := strings.Split(d.Config.ServerAddress, "@")
	if len(parts) != 2 {
		return 0, NewClientError(ErrTypeConfiguration, "invalid DNS server address format", nil)
	}
	dnsServer := parts[1]

	// Set timeout if configured
	if d.Config.WriteTimeout > 0 {
		d.client.Timeout = time.Duration(d.Config.WriteTimeout) * time.Second
	}

	// Encode the data as base64
	encoded := base64.StdEncoding.EncodeToString(data)

	// Split the data into chunks if it's too large
	chunks := splitString(encoded, d.maxDataSize)
	totalSent := 0

	// Send each chunk as a separate DNS query
	for i, chunk := range chunks {
		// Create a unique subdomain for each chunk
		subdomain := d.createSubdomain(i, len(chunks))

		// Create a DNS message
		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn(subdomain+"."+d.domain), d.queryType)
		m.RecursionDesired = true

		// Add the data as a TXT record
		txt := new(dns.TXT)
		txt.Hdr = dns.RR_Header{
			Name:   dns.Fqdn(subdomain + "." + d.domain),
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    60,
		}
		txt.Txt = []string{chunk}
		m.Extra = append(m.Extra, txt)

		// Send the query
		_, _, err := d.client.Exchange(m, dnsServer)
		if err != nil {
			d.setLastError(err)
			return totalSent, NewClientError(ErrTypeSend, "failed to send data chunk", err)
		}

		totalSent += len(chunk)
	}

	// Reset timeout to default
	d.client.Timeout = d.timeout

	return len(data), nil
}

// Receive receives data from the DNS connection
func (d *DNSProtocol) Receive() ([]byte, error) {
	if !d.IsConnected() {
		return nil, ErrNotConnected
	}

	// Get the DNS server from the server address
	parts := strings.Split(d.Config.ServerAddress, "@")
	if len(parts) != 2 {
		return nil, NewClientError(ErrTypeConfiguration, "invalid DNS server address format", nil)
	}
	dnsServer := parts[1]

	// Set timeout if configured
	if d.Config.ReadTimeout > 0 {
		d.client.Timeout = time.Duration(d.Config.ReadTimeout) * time.Second
	}

	// Create a DNS message to query for response data
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn("response."+d.domain), dns.TypeTXT)
	m.RecursionDesired = true

	// Send the query
	r, _, err := d.client.Exchange(m, dnsServer)
	if err != nil {
		// Check if it's a timeout error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			d.setLastError(err)
			return nil, NewClientError(ErrTypeTimeout, "read timeout", err)
		}

		// Handle other errors
		d.setLastError(err)
		return nil, NewClientError(ErrTypeReceive, "failed to receive data", err)
	}

	// Check if we got a valid response
	if r.Rcode != dns.RcodeSuccess {
		d.setLastError(errors.New("DNS query failed"))
		return nil, NewClientError(ErrTypeReceive, "DNS query failed with code: "+dns.RcodeToString[r.Rcode], nil)
	}

	// Extract TXT records from the response
	var txtData []string
	for _, a := range r.Answer {
		if txt, ok := a.(*dns.TXT); ok {
			txtData = append(txtData, txt.Txt...)
		}
	}

	// If no TXT records were found, return an error
	if len(txtData) == 0 {
		d.setLastError(errors.New("no TXT records found"))
		return nil, NewClientError(ErrTypeReceive, "no TXT records found in response", nil)
	}

	// Join all TXT records
	joined := strings.Join(txtData, "")

	// Decode the base64 data
	decoded, err := base64.StdEncoding.DecodeString(joined)
	if err != nil {
		d.setLastError(err)
		return nil, NewClientError(ErrTypeReceive, "failed to decode base64 data", err)
	}

	// Reset timeout to default
	d.client.Timeout = d.timeout

	return decoded, nil
}

// GetConnection returns the underlying connection
// Note: This overrides the base method since DNS doesn't use net.Conn
func (d *DNSProtocol) GetConnection() net.Conn {
	return nil
}

// ValidateConfig validates the DNS protocol configuration
func (d *DNSProtocol) ValidateConfig() error {
	// Call the base ValidateConfig method
	if err := d.BaseProtocol.ValidateConfig(); err != nil {
		return err
	}

	// Additional DNS-specific validation
	parts := strings.Split(d.Config.ServerAddress, "@")
	if len(parts) != 2 {
		return NewClientError(ErrTypeConfiguration, "invalid DNS server address format, expected 'domain@server'", nil)
	}

	domain := parts[0]
	server := parts[1]

	// Check if the domain is valid
	if domain == "" {
		return NewClientError(ErrTypeConfiguration, "domain cannot be empty", nil)
	}

	// Check if the server is valid
	if server == "" {
		return NewClientError(ErrTypeConfiguration, "DNS server cannot be empty", nil)
	}

	// Try to resolve the DNS server
	_, err := net.ResolveIPAddr("ip", server)
	if err != nil {
		// If it's not an IP address, check if it's a valid hostname
		if !strings.Contains(server, ":") {
			server += ":53" // Add default DNS port
		}
		_, err = net.ResolveTCPAddr("tcp", server)
		if err != nil {
			return NewClientError(ErrTypeConfiguration, "invalid DNS server address", err)
		}
	}

	return nil
}

// IsConnected returns true if the DNS connection is connected
// Note: This overrides the base method since DNS doesn't have a persistent connection
func (d *DNSProtocol) IsConnected() bool {
	return d.GetStatus() == StatusConnected
}

// SetQueryType sets the DNS query type to use
func (d *DNSProtocol) SetQueryType(queryType uint16) {
	d.queryType = queryType
}

// SetMaxDataSize sets the maximum size for data chunks
func (d *DNSProtocol) SetMaxDataSize(size int) {
	if size > 0 {
		d.maxDataSize = size
	}
}

// createSubdomain creates a unique subdomain for a chunk
func (d *DNSProtocol) createSubdomain(index, total int) string {
	return "c" + padNumber(index, total) + "of" + padNumber(total, total)
}

// Helper functions

// splitString splits a string into chunks of the specified size
func splitString(s string, chunkSize int) []string {
	if chunkSize <= 0 {
		return []string{s}
	}

	var chunks []string
	for i := 0; i < len(s); i += chunkSize {
		end := i + chunkSize
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

// padNumber pads a number with zeros to match the width of the max number
func padNumber(num, max int) string {
	width := len(string(rune(max)))
	format := "%0" + string(rune('0'+width)) + "d"
	return string([]byte(format))[0:2] + "d"
}
