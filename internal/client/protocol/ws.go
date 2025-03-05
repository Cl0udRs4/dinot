package protocol

import (
    "context"
    "net/http"
    "time"

    "github.com/gorilla/websocket"
)

// WSProtocol implements the Protocol interface for WebSocket
type WSProtocol struct {
    BaseProtocol
    URL     string
    Headers http.Header
    Conn    *websocket.Conn
}

// NewWSProtocol creates a new WebSocket protocol
func NewWSProtocol(url string) *WSProtocol {
    return &WSProtocol{
        BaseProtocol: BaseProtocol{
            Name:      "ws",
            Connected: false,
            Timeout:   30 * time.Second,
        },
        URL:     url,
        Headers: http.Header{},
    }
}

// Connect establishes a WebSocket connection to the server
func (p *WSProtocol) Connect(ctx context.Context) error {
    dialer := websocket.Dialer{
        HandshakeTimeout: p.Timeout,
    }
    
    conn, _, err := dialer.DialContext(ctx, p.URL, p.Headers)
    if err != nil {
        return err
    }
    
    p.Conn = conn
    p.Connected = true
    return nil
}

// Disconnect closes the WebSocket connection
func (p *WSProtocol) Disconnect() error {
    if !p.Connected || p.Conn == nil {
        return nil
    }
    
    err := p.Conn.Close()
    p.Connected = false
    p.Conn = nil
    return err
}

// Send sends data over the WebSocket connection
func (p *WSProtocol) Send(data []byte) error {
    if !p.Connected || p.Conn == nil {
        return ErrNotConnected
    }
    
    return p.Conn.WriteMessage(websocket.BinaryMessage, data)
}

// Receive receives data from the WebSocket connection with timeout
func (p *WSProtocol) Receive(timeout time.Duration) ([]byte, error) {
    if !p.Connected || p.Conn == nil {
        return nil, ErrNotConnected
    }
    
    if timeout > 0 {
        err := p.Conn.SetReadDeadline(time.Now().Add(timeout))
        if err != nil {
            return nil, err
        }
    }
    
    messageType, data, err := p.Conn.ReadMessage()
    if err != nil {
        return nil, err
    }
    
    if messageType != websocket.BinaryMessage && messageType != websocket.TextMessage {
        return nil, ErrInvalidMessageType
    }
    
    return data, nil
}

// SetHeader sets a header for the WebSocket connection
func (p *WSProtocol) SetHeader(key, value string) {
    p.Headers.Set(key, value)
}
