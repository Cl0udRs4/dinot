package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/api"
	"github.com/Cl0udRs4/dinot/internal/server/client"
	"github.com/Cl0udRs4/dinot/internal/server/listener"
	"github.com/Cl0udRs4/dinot/internal/server/logging"
)

func main() {
	// Parse command line flags
	logLevelStr := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	_ = flag.Bool("console", true, "Enable console mode")
	enableAPI := flag.Bool("api", false, "Enable API mode")
	apiPort := flag.Int("api-port", 8090, "API server port")
	tcpPort := flag.Int("tcp-port", 8080, "TCP listener port")
	udpPort := flag.Int("udp-port", 8081, "UDP listener port")
	wsPort := flag.Int("ws-port", 8082, "WebSocket listener port")
	dnsPort := flag.Int("dns-port", 8053, "DNS listener port")
	flag.Parse()

	// Initialize logger
	logger := logging.GetLogger()
	logLevel := logging.LogLevel(*logLevelStr)
	logger.SetLevel(logLevel)

	// Initialize client manager
	clientManager := client.NewClientManager()
	
	// Initialize heartbeat monitor
	checkInterval := 10 * time.Second
	timeout := 60 * time.Second
	heartbeatMonitor := client.NewHeartbeatMonitor(clientManager, checkInterval, timeout)

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize TCP listener
	tcpConfig := listener.Config{
		Address:        fmt.Sprintf("0.0.0.0:%d", *tcpPort),
		BufferSize:     4096,
		MaxConnections: 100,
		Timeout:        60,
	}
	tcpListener := listener.NewTCPListener(tcpConfig)

	// Initialize UDP listener
	udpConfig := listener.Config{
		Address:        fmt.Sprintf("0.0.0.0:%d", *udpPort),
		BufferSize:     4096,
		MaxConnections: 100,
		Timeout:        60,
	}
	udpListener := listener.NewUDPListener(udpConfig)

	// Initialize WebSocket listener
	wsConfig := listener.Config{
		Address:        fmt.Sprintf("0.0.0.0:%d", *wsPort),
		BufferSize:     4096,
		MaxConnections: 100,
		Timeout:        60,
	}
	wsListener := listener.NewWSListener(wsConfig)

	// Initialize DNS listener
	dnsConfig := listener.DNSConfig{
		Config: listener.Config{
			Address:        fmt.Sprintf("0.0.0.0:%d", *dnsPort),
			BufferSize:     4096,
			MaxConnections: 100,
			Timeout:        60,
		},
		Domain:      "example.com",
		TTL:         300,
		RecordTypes: []string{"A", "TXT"},
	}
	dnsListener := listener.NewDNSListener(dnsConfig)

	// Create a connection handler function
	connectionHandler := func(conn net.Conn) {
		clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
		fmt.Printf("New connection from %s (ID: %s)\n", conn.RemoteAddr().String(), clientID)
		
		// Create a new client
		c := client.NewClient(
			clientID,
			"Client-"+clientID,
			conn.RemoteAddr().String(),
			"unknown",
			"unknown",
			[]string{},
			"tcp",
		)
		
		// Register the client
		clientManager.RegisterClient(c)
		
		// Handle client communication
		go handleClient(conn, clientID, clientManager)
	}
	
	// Initialize and start API server if enabled
	if *enableAPI {
		apiConfig := api.Config{
			Address:      fmt.Sprintf("0.0.0.0:%d", *apiPort),
			AuthEnabled:  false,
			AuthUser:     "",
			AuthPassword: "",
			JWTSecret:    "",
			JWTEnabled:   false,
		}
		apiHandler := api.NewAPIHandler(clientManager, heartbeatMonitor, apiConfig)
		go func() {
			fmt.Printf("Starting API server on %s\n", apiConfig.Address)
			if err := apiHandler.Start(apiConfig.Address); err != nil {
				fmt.Printf("Error starting API server: %v\n", err)
			}
		}()
	}
	
	// Start the TCP listener
	err := tcpListener.Start(ctx, connectionHandler)
	if err != nil {
		fmt.Printf("Error starting TCP listener: %v\n", err)
		os.Exit(1)
	}
	
	// Start the UDP listener
	err = udpListener.Start(ctx, connectionHandler)
	if err != nil {
		fmt.Printf("Error starting UDP listener: %v\n", err)
	} else {
		fmt.Printf("UDP listener started on %s\n", udpConfig.Address)
	}
	
	// Start the WebSocket listener
	err = wsListener.Start(ctx, connectionHandler)
	if err != nil {
		fmt.Printf("Error starting WebSocket listener: %v\n", err)
	} else {
		fmt.Printf("WebSocket listener started on %s\n", wsConfig.Address)
	}
	
	// Start the DNS listener
	err = dnsListener.Start(ctx, connectionHandler)
	if err != nil {
		fmt.Printf("Error starting DNS listener: %v\n", err)
	} else {
		fmt.Printf("DNS listener started on %s\n", dnsConfig.Address)
	}

	fmt.Println("Server started on 0.0.0.0:8080")
	fmt.Println("Press Ctrl+C to stop the server")

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	fmt.Println("Shutting down server...")
	tcpListener.Stop()
	udpListener.Stop()
	wsListener.Stop()
	dnsListener.Stop()
	
	// Stop the heartbeat monitor
	heartbeatMonitor.Stop()
	
	fmt.Println("Server shutdown complete")
}

// handleClient handles communication with a client
func handleClient(conn net.Conn, clientID string, clientManager *client.ClientManager) {
	defer func() {
		conn.Close()
		clientManager.UnregisterClient(clientID)
		fmt.Printf("Client disconnected: %s\n", clientID)
	}()

	buffer := make([]byte, 4096)

	for {
		// Set read deadline
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		// Read data from the connection
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Error reading from client %s: %v\n", clientID, err)
			break
		}

		// Process the data
		data := buffer[:n]
		fmt.Printf("Received %d bytes from client %s\n", n, clientID)

		// Update client's last seen time
		client, err := clientManager.GetClient(clientID)
		if err == nil {
			client.UpdateLastSeen()
		}

		// Echo the data back to the client
		_, err = conn.Write(data)
		if err != nil {
			fmt.Printf("Error writing to client %s: %v\n", clientID, err)
			break
		}
	}
}
