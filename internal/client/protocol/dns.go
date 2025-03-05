package protocol

import (
    "context"
    "encoding/base64"
    "fmt"
    "strings"
    "time"

    "github.com/miekg/dns"
)

// DNSProtocol implements the Protocol interface for DNS
type DNSProtocol struct {
    BaseProtocol
    Domain      string
    Server      string
    Client      *dns.Client
    MessageSize int
}

// NewDNSProtocol creates a new DNS protocol
func NewDNSProtocol(domain, server string) *DNSProtocol {
    return &DNSProtocol{
        BaseProtocol: BaseProtocol{
            Name:      "dns",
            Connected: false,
            Timeout:   30 * time.Second,
        },
        Domain:      domain,
        Server:      server,
        Client:      &dns.Client{Timeout: 30 * time.Second},
        MessageSize: 200, // Maximum size of data in a single DNS message
    }
}

// Connect establishes a DNS connection
func (p *DNSProtocol) Connect(ctx context.Context) error {
    // DNS is connectionless, so we just set connected to true
    p.Connected = true
    return nil
}

// Disconnect closes the DNS connection
func (p *DNSProtocol) Disconnect() error {
    p.Connected = false
    return nil
}

// Send sends data over DNS
func (p *DNSProtocol) Send(data []byte) error {
    if !p.Connected {
        return ErrNotConnected
    }
    
    // Encode data as base64
    encoded := base64.StdEncoding.EncodeToString(data)
    
    // Split data into chunks if necessary
    chunks := splitString(encoded, p.MessageSize)
    
    for i, chunk := range chunks {
        // Create a DNS message
        m := new(dns.Msg)
        m.SetQuestion(fmt.Sprintf("%d.%s.%s", i, chunk, p.Domain), dns.TypeTXT)
        
        // Send the DNS query
        _, _, err := p.Client.Exchange(m, p.Server)
        if err != nil {
            return err
        }
    }
    
    return nil
}

// Receive receives data from DNS with timeout
func (p *DNSProtocol) Receive(timeout time.Duration) ([]byte, error) {
    if !p.Connected {
        return nil, ErrNotConnected
    }
    
    // Create a DNS message
    m := new(dns.Msg)
    m.SetQuestion(fmt.Sprintf("recv.%s", p.Domain), dns.TypeTXT)
    
    // Set timeout
    client := &dns.Client{Timeout: timeout}
    
    // Send the DNS query
    resp, _, err := client.Exchange(m, p.Server)
    if err != nil {
        return nil, err
    }
    
    if len(resp.Answer) == 0 {
        return nil, ErrTimeout
    }
    
    // Extract data from TXT records
    var builder strings.Builder
    for _, ans := range resp.Answer {
        if txt, ok := ans.(*dns.TXT); ok {
            for _, t := range txt.Txt {
                builder.WriteString(t)
            }
        }
    }
    
    // Decode base64 data
    decoded, err := base64.StdEncoding.DecodeString(builder.String())
    if err != nil {
        return nil, err
    }
    
    return decoded, nil
}

// splitString splits a string into chunks of the specified size
func splitString(s string, chunkSize int) []string {
    if len(s) == 0 {
        return []string{}
    }
    
    if chunkSize >= len(s) {
        return []string{s}
    }
    
    var chunks []string
    currentLen := 0
    currentStart := 0
    
    for i := range s {
        currentLen++
        if currentLen == chunkSize {
            chunks = append(chunks, s[currentStart:i+1])
            currentLen = 0
            currentStart = i + 1
        }
    }
    
    if currentStart < len(s) {
        chunks = append(chunks, s[currentStart:])
    }
    
    return chunks
}
