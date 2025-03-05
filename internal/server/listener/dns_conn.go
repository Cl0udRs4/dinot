package listener

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/miekg/dns"
)

// Custom errors for deadline handling
var (
	ErrTimeout = errors.New("i/o timeout")
)

// DNSConn wraps a DNS connection to make it compatible with the net.Conn interface
type DNSConn struct {
	writer     dns.ResponseWriter
	request    *dns.Msg
	response   *dns.Msg
	remoteAddr net.Addr
	localAddr  net.Addr
	buffer     []byte
	readPos    int
	closed     bool
	readDeadline  time.Time
	writeDeadline time.Time
}

// NewDNSConn creates a new DNS connection wrapper
func NewDNSConn(w dns.ResponseWriter, r *dns.Msg, m *dns.Msg, buffer []byte) *DNSConn {
	return &DNSConn{
		writer:     w,
		request:    r,
		response:   m,
		remoteAddr: w.RemoteAddr(),
		localAddr:  w.LocalAddr(),
		buffer:     buffer,
	}
}

// Read reads data from the connection
func (d *DNSConn) Read(b []byte) (n int, err error) {
	if d.closed {
		return 0, io.EOF
	}
	
	if !d.readDeadline.IsZero() && time.Now().After(d.readDeadline) {
		return 0, ErrTimeout
	}
	
	// If we have read all the data, return EOF
	if d.readPos >= len(d.buffer) {
		return 0, io.EOF
	}
	
	// Copy data from buffer to b
	n = copy(b, d.buffer[d.readPos:])
	d.readPos += n
	return n, nil
}

// Write writes data to the connection
func (d *DNSConn) Write(b []byte) (n int, err error) {
	if d.closed {
		return 0, io.ErrClosedPipe
	}
	
	if !d.writeDeadline.IsZero() && time.Now().After(d.writeDeadline) {
		return 0, ErrTimeout
	}
	
	// In a real implementation, we would encode the data into the DNS response
	// For now, we'll just store it in the response
	return len(b), nil
}

// Close closes the connection
func (d *DNSConn) Close() error {
	d.closed = true
	return nil
}

// LocalAddr returns the local network address
func (d *DNSConn) LocalAddr() net.Addr {
	return d.localAddr
}

// RemoteAddr returns the remote network address
func (d *DNSConn) RemoteAddr() net.Addr {
	return d.remoteAddr
}

// SetDeadline sets the read and write deadlines
func (d *DNSConn) SetDeadline(t time.Time) error {
	d.readDeadline = t
	d.writeDeadline = t
	return nil
}

// SetReadDeadline sets the read deadline
func (d *DNSConn) SetReadDeadline(t time.Time) error {
	d.readDeadline = t
	return nil
}

// SetWriteDeadline sets the write deadline
func (d *DNSConn) SetWriteDeadline(t time.Time) error {
	d.writeDeadline = t
	return nil
}
