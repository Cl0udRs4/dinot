package listener

import (
	"context"
	"net"
	"sync"

	"github.com/Cl0udRs4/dinot/internal/server/common"
	"github.com/miekg/dns"
)

// DNSListener implements the Listener interface for DNS protocol
type DNSListener struct {
	*BaseListener
	server     *dns.Server
	clientsMtx sync.RWMutex
	clients    map[string]net.Addr
	handler    ConnectionHandler
}

// DNSConfig extends the base Config with DNS-specific settings
type DNSConfig struct {
	Config
	// Domain is the base domain for DNS queries
	Domain string
	
	// TTL is the time-to-live for DNS records
	TTL uint32
	
	// RecordTypes is a list of supported DNS record types (A, TXT, etc.)
	RecordTypes []string
}

// NewDNSListener creates a new DNS listener
func NewDNSListener(config DNSConfig) *DNSListener {
	return &DNSListener{
		BaseListener: NewBaseListener("dns", config.Config),
		clients:      make(map[string]net.Addr),
	}
}

// Start starts the DNS listener
func (d *DNSListener) Start(ctx context.Context, handler ConnectionHandler) error {
	if d.GetStatus() == StatusRunning {
		return common.NewServerError(common.ErrListenerAlreadyRunning, "DNS listener is already running", nil)
	}

	if err := d.ValidateConfig(); err != nil {
		return err
	}

	// Store the connection handler
	d.handler = handler

	// Create a new DNS server
	d.server = &dns.Server{
		Addr:    d.Config.Address,
		Net:     "udp",
		Handler: dns.HandlerFunc(d.handleDNSRequest),
	}

	ctx, d.cancel = context.WithCancel(ctx)
	d.setStatus(StatusRunning)

	// Start the DNS server in a separate goroutine
	go func() {
		err := d.server.ListenAndServe()
		if err != nil {
			d.setStatus(StatusError)
		}
	}()

	// Create a goroutine to handle the context cancellation
	go func() {
		<-ctx.Done()
		if d.server != nil {
			// Shutdown the server
			d.server.Shutdown()
		}
	}()

	return nil
}

// Stop stops the DNS listener
func (d *DNSListener) Stop() error {
	if d.GetStatus() != StatusRunning {
		return common.NewServerError(common.ErrListenerNotRunning, "DNS listener is not running", nil)
	}

	if d.server != nil {
		// Shutdown the server
		d.server.Shutdown()
	}

	// Clear clients map
	d.clientsMtx.Lock()
	d.clients = make(map[string]net.Addr)
	d.clientsMtx.Unlock()

	// Call the base Stop method to cancel the context and update status
	return d.BaseListener.Stop()
}

// handleDNSRequest handles incoming DNS requests
func (d *DNSListener) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	// Create a response message
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = true

	// Get client address
	clientAddr := w.RemoteAddr()
	clientKey := clientAddr.String()

	// Store client address
	d.clientsMtx.Lock()
	d.clients[clientKey] = clientAddr
	d.clientsMtx.Unlock()

	// Check if this is a query we should handle
	if len(r.Question) > 0 {
		// Extract data from the DNS query
		question := r.Question[0]
		
		// Create a DNS connection wrapper
		conn := NewDNSConn(w, r, m, []byte(question.Name))

		// Handle the connection in a separate goroutine
		go func() {
			// Call the connection handler
			d.handler(conn)
			
			// Send the response
			w.WriteMsg(m)
		}()
	} else {
		// Send an empty response for queries we don't handle
		w.WriteMsg(m)
	}
}
