package protocol

import (
    "context"
    "net"
    "time"
)

// TCPProtocol implements the Protocol interface for TCP
type TCPProtocol struct {
    BaseProtocol
    Address string
    Conn    net.Conn
}

// NewTCPProtocol creates a new TCP protocol
func NewTCPProtocol(address string) *TCPProtocol {
    return &TCPProtocol{
        BaseProtocol: BaseProtocol{
            Name:      "tcp",
            Connected: false,
            Timeout:   30 * time.Second,
        },
        Address: address,
    }
}

// Connect establishes a TCP connection to the server
func (p *TCPProtocol) Connect(ctx context.Context) error {
    dialer := net.Dialer{Timeout: p.Timeout}
    conn, err := dialer.DialContext(ctx, "tcp", p.Address)
    if err != nil {
        return err
    }
    
    p.Conn = conn
    p.Connected = true
    return nil
}

// Disconnect closes the TCP connection
func (p *TCPProtocol) Disconnect() error {
    if !p.Connected || p.Conn == nil {
        return nil
    }
    
    err := p.Conn.Close()
    p.Connected = false
    p.Conn = nil
    return err
}

// Send sends data over the TCP connection
func (p *TCPProtocol) Send(data []byte) error {
    if !p.Connected || p.Conn == nil {
        return ErrNotConnected
    }
    
    _, err := p.Conn.Write(data)
    return err
}

// Receive receives data from the TCP connection with timeout
func (p *TCPProtocol) Receive(timeout time.Duration) ([]byte, error) {
    if !p.Connected || p.Conn == nil {
        return nil, ErrNotConnected
    }
    
    buffer := make([]byte, 4096)
    if timeout > 0 {
        err := p.Conn.SetReadDeadline(time.Now().Add(timeout))
        if err != nil {
            return nil, err
        }
    }
    
    n, err := p.Conn.Read(buffer)
    if err != nil {
        return nil, err
    }
    
    return buffer[:n], nil
}
