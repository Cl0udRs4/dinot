package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/client"
	"github.com/Cl0udRs4/dinot/internal/server/listener"
)

// Server represents the C2 server with console interface
type Server struct {
	// listenerManager manages protocol listeners
	listenerManager *listener.ListenerManager
	
	// clientManager manages connected clients
	clientManager *client.ClientManager
	
	// heartbeatMonitor monitors client heartbeats
	heartbeatMonitor *client.HeartbeatMonitor
	
	// console is the command-line interface
	console *Console
}

// NewServer creates a new C2 server with console interface
func NewServer() *Server {
	clientManager := client.NewClientManager()
	heartbeatMonitor := client.NewHeartbeatMonitor(clientManager, 30*time.Second, 60*time.Second)
	
	// Create default listener config
	defaultConfig := listener.Config{
		Address:        "0.0.0.0:8080",
		BufferSize:     4096,
		MaxConnections: 100,
		Timeout:        30,
	}
	
	listenerManager := listener.NewListenerManager(defaultConfig)
	
	return &Server{
		listenerManager:  listenerManager,
		clientManager:    clientManager,
		heartbeatMonitor: heartbeatMonitor,
		console:          NewConsole(clientManager, heartbeatMonitor),
	}
}

// Start starts the C2 server and console interface
func (s *Server) Start() error {
	// Start the heartbeat monitor
	s.heartbeatMonitor.Start()
	
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		s.Stop()
		os.Exit(0)
	}()
	
	// Start the console interface
	s.console.Start()
	
	return nil
}

// Stop stops the C2 server and console interface
func (s *Server) Stop() {
	// Stop the console interface
	s.console.Stop()
	
	// Stop the heartbeat monitor
	s.heartbeatMonitor.Stop()
	
	// Stop all listeners
	s.listenerManager.HaltAll()
}

// GetClientManager returns the client manager
func (s *Server) GetClientManager() *client.ClientManager {
	return s.clientManager
}

// GetHeartbeatMonitor returns the heartbeat monitor
func (s *Server) GetHeartbeatMonitor() *client.HeartbeatMonitor {
	return s.heartbeatMonitor
}

// GetListenerManager returns the listener manager
func (s *Server) GetListenerManager() *listener.ListenerManager {
	return s.listenerManager
}
