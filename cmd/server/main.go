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

	"github.com/Cl0udRs4/dinot/internal/server/client"
	"github.com/Cl0udRs4/dinot/internal/server/listener"
	"github.com/Cl0udRs4/dinot/internal/server/logging"
)

func main() {
	// Parse command line flags
	logLevelStr := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	// Unused flags for future implementation
	_ = flag.Bool("console", true, "Enable console mode")
	_ = flag.Bool("api", false, "Enable API mode")
	flag.Parse()

	// Initialize logger
	logger := logging.GetLogger()
	logLevel := logging.LogLevel(*logLevelStr)
	logger.SetLevel(logLevel)

	// Initialize client manager
	clientManager := client.NewClientManager()

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize TCP listener
	tcpConfig := listener.Config{
		Address:        "0.0.0.0:8080",
		BufferSize:     4096,
		MaxConnections: 100,
		Timeout:        60,
	}

	// Create TCP listener
	tcpListener := listener.NewTCPListener(tcpConfig)

	// Start the TCP listener
	err := tcpListener.Start(ctx, func(conn net.Conn) {
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
	})

	if err != nil {
		fmt.Printf("Error starting TCP listener: %v\n", err)
		os.Exit(1)
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
