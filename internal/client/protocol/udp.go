package protocol

import (
    "context"
    "net"
    "time"
)

// UDPProtocol implements the Protocol interface for UDP
type UDPProtocol struct {
    BaseProtocol
    Address     string
    Conn        *net.UDPConn
    RemoteAddr  *net.UDPAddr
}

// NewUDPProtocol creates a new UDP protocol
func NewUDPProtocol(address string) *UDPProtocol {
    return &UDPProtocol{
        BaseProtocol: BaseProtocol{
            Name:      "udp",
            Connected: false,
            Timeout:   30 * time.Second,
        },
        Address: address,
    }
}

// Connect establishes a UDP connection to the server
func (p *UDPProtocol) Connect(ctx context.Context) error {
    remoteAddr, err := net.ResolveUDPAddr("udp", p.Address)
    if err != nil {
        return err
    }
    
    conn, err := net.DialUDP("udp", nil, remoteAddr)
    if err != nil {
        return err
    }
    
    p.Conn = conn
    p.RemoteAddr = remoteAddr
    p.Connected = true
    return nil
}

// Disconnect closes the UDP connection
func (p *UDPProtocol) Disconnect() error {
    if !p.Connected || p.Conn == nil {
        return nil
    }
    
    err := p.Conn.Close()
    p.Connected = false
    p.Conn = nil
    p.RemoteAddr = nil
    return err
}

// Send sends data over the UDP connection
func (p *UDPProtocol) Send(data []byte) error {
    if !p.Connected || p.Conn == nil {
        return ErrNotConnected
    }
    
    _, err := p.Conn.Write(data)
    return err
}

// Receive receives data from the UDP connection with timeout
func (p *UDPProtocol) Receive(timeout time.Duration) ([]byte, error) {
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
    
    n, _, err := p.Conn.ReadFromUDP(buffer)
    if err != nil {
        return nil, err
    }
    
    return buffer[:n], nil
}
