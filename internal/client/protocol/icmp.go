package protocol

import (
    "context"
    "net"
    "os"
    "time"
    "golang.org/x/net/icmp"
    "golang.org/x/net/ipv4"
)

// ICMPProtocol implements the Protocol interface for ICMP
type ICMPProtocol struct {
    BaseProtocol
    Address     string
    Conn        *icmp.PacketConn
    SequenceNum int
    ID          int
}

// NewICMPProtocol creates a new ICMP protocol
func NewICMPProtocol(address string) *ICMPProtocol {
    return &ICMPProtocol{
        BaseProtocol: BaseProtocol{
            Name:      "icmp",
            Connected: false,
            Timeout:   30 * time.Second,
        },
        Address:     address,
        SequenceNum: 1,
        ID:          os.Getpid() & 0xffff,
    }
}

// Connect establishes an ICMP connection
func (p *ICMPProtocol) Connect(ctx context.Context) error {
    conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
    if err != nil {
        return err
    }
    
    p.Conn = conn
    p.Connected = true
    return nil
}

// Disconnect closes the ICMP connection
func (p *ICMPProtocol) Disconnect() error {
    if !p.Connected || p.Conn == nil {
        return nil
    }
    
    err := p.Conn.Close()
    p.Connected = false
    p.Conn = nil
    return err
}

// Send sends data over the ICMP connection
func (p *ICMPProtocol) Send(data []byte) error {
    if !p.Connected || p.Conn == nil {
        return ErrNotConnected
    }
    
    dst, err := net.ResolveIPAddr("ip4", p.Address)
    if err != nil {
        return err
    }
    
    msg := icmp.Message{
        Type: ipv4.ICMPTypeEcho,
        Code: 0,
        Body: &icmp.Echo{
            ID:   p.ID,
            Seq:  p.SequenceNum,
            Data: data,
        },
    }
    
    msgBytes, err := msg.Marshal(nil)
    if err != nil {
        return err
    }
    
    _, err = p.Conn.WriteTo(msgBytes, dst)
    if err != nil {
        return err
    }
    
    p.SequenceNum++
    return nil
}

// Receive receives data from the ICMP connection with timeout
func (p *ICMPProtocol) Receive(timeout time.Duration) ([]byte, error) {
    if !p.Connected || p.Conn == nil {
        return nil, ErrNotConnected
    }
    
    if timeout > 0 {
        err := p.Conn.SetReadDeadline(time.Now().Add(timeout))
        if err != nil {
            return nil, err
        }
    }
    
    buffer := make([]byte, 1500)
    n, _, err := p.Conn.ReadFrom(buffer)
    if err != nil {
        return nil, err
    }
    
    msg, err := icmp.ParseMessage(ipv4.ICMPTypeEcho.Protocol(), buffer[:n])
    if err != nil {
        return nil, err
    }
    
    if msg.Type != ipv4.ICMPTypeEchoReply {
        return nil, ErrInvalidMessageType
    }
    
    echo, ok := msg.Body.(*icmp.Echo)
    if !ok {
        return nil, ErrInvalidMessageType
    }
    
    return echo.Data, nil
}
