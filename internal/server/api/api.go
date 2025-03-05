package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/client"
)

// APIHandler represents the HTTP API handler
type APIHandler struct {
	// clientManager is the client manager to interact with
	clientManager *client.ClientManager

	// heartbeatMonitor is the heartbeat monitor to interact with
	heartbeatMonitor *client.HeartbeatMonitor

	// authEnabled indicates whether authentication is enabled
	authEnabled bool

	// authUser is the username for basic auth
	authUser string

	// authPassword is the password for basic auth
	authPassword string

	// jwtSecret is the secret for JWT authentication
	jwtSecret string

	// jwtEnabled indicates whether JWT authentication is enabled
	jwtEnabled bool
}

// Config represents the API configuration
type Config struct {
	// Address is the address to listen on
	Address string

	// AuthEnabled indicates whether authentication is enabled
	AuthEnabled bool

	// AuthUser is the username for basic auth
	AuthUser string

	// AuthPassword is the password for basic auth
	AuthPassword string

	// JWTSecret is the secret for JWT authentication
	JWTSecret string

	// JWTEnabled indicates whether JWT authentication is enabled
	JWTEnabled bool
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(clientManager *client.ClientManager, heartbeatMonitor *client.HeartbeatMonitor, config Config) *APIHandler {
	return &APIHandler{
		clientManager:    clientManager,
		heartbeatMonitor: heartbeatMonitor,
		authEnabled:      config.AuthEnabled,
		authUser:         config.AuthUser,
		authPassword:     config.AuthPassword,
		jwtSecret:        config.JWTSecret,
		jwtEnabled:       config.JWTEnabled,
	}
}

// Start starts the HTTP API server
func (h *APIHandler) Start(address string) error {
	// Register API routes
	http.HandleFunc("/api/clients", h.authMiddleware(h.handleClients))
	http.HandleFunc("/api/clients/", h.authMiddleware(h.handleClient))
	http.HandleFunc("/api/heartbeat", h.authMiddleware(h.handleHeartbeat))
	http.HandleFunc("/api/status", h.authMiddleware(h.handleStatus))
	http.HandleFunc("/api/exceptions", h.authMiddleware(h.handleExceptions))
	http.HandleFunc("/api/exceptions/", h.authMiddleware(h.handleException))

	// Start the HTTP server
	fmt.Printf("Starting HTTP API server on %s\n", address)
	return http.ListenAndServe(address, nil)
}

// authMiddleware is a middleware that handles authentication
func (h *APIHandler) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication if disabled
		if !h.authEnabled {
			next(w, r)
			return
		}

		// Check for JWT authentication
		if h.jwtEnabled {
			token := r.Header.Get("Authorization")
			if token != "" && strings.HasPrefix(token, "Bearer ") {
				token = strings.TrimPrefix(token, "Bearer ")
				if h.validateJWT(token) {
					next(w, r)
					return
				}
			}
		}

		// Check for basic authentication
		username, password, ok := r.BasicAuth()
		if ok && username == h.authUser && password == h.authPassword {
			next(w, r)
			return
		}

		// Authentication failed
		w.Header().Set("WWW-Authenticate", `Basic realm="C2 API"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
}

// validateJWT validates a JWT token
func (h *APIHandler) validateJWT(token string) bool {
	// TODO: Implement JWT validation
	// This is a placeholder for JWT validation
	return token == "valid-token"
}

// handleClients handles the /api/clients endpoint
func (h *APIHandler) handleClients(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Get all clients or filter by status
		status := r.URL.Query().Get("status")
		var clients []*client.Client
		if status != "" {
			clients = h.clientManager.GetClientsByStatus(client.ClientStatus(status))
		} else {
			clients = h.clientManager.GetAllClients()
		}

		// Return the clients as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clients)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleClient handles the /api/clients/{id} endpoint
func (h *APIHandler) handleClient(w http.ResponseWriter, r *http.Request) {
	// Extract the client ID from the URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}
	clientID := parts[len(parts)-1]

	switch r.Method {
	case http.MethodGet:
		// Get client details
		client, err := h.clientManager.GetClient(clientID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Return the client as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(client)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleHeartbeat handles the /api/heartbeat endpoint
func (h *APIHandler) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Get heartbeat settings
		settings := map[string]interface{}{
			"checkInterval":     h.heartbeatMonitor.GetCheckInterval(),
			"timeout":           h.heartbeatMonitor.GetTimeout(),
			"randomEnabled":     h.heartbeatMonitor.IsRandomEnabled(),
			"randomMinInterval": h.heartbeatMonitor.GetRandomMinInterval(),
			"randomMaxInterval": h.heartbeatMonitor.GetRandomMaxInterval(),
		}

		// Return the settings as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(settings)

	case http.MethodPost:
		// Update heartbeat settings
		var settings map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Apply the settings
		if checkInterval, ok := settings["checkInterval"].(float64); ok {
			h.heartbeatMonitor.SetCheckInterval(time.Duration(checkInterval) * time.Second)
		}
		if timeout, ok := settings["timeout"].(float64); ok {
			h.heartbeatMonitor.SetTimeout(time.Duration(timeout) * time.Second)
		}
		if randomEnabled, ok := settings["randomEnabled"].(bool); ok {
			if randomEnabled {
				minInterval := time.Duration(settings["randomMinInterval"].(float64)) * time.Second
				maxInterval := time.Duration(settings["randomMaxInterval"].(float64)) * time.Second
				h.heartbeatMonitor.EnableRandomIntervals(minInterval, maxInterval)
			} else {
				h.heartbeatMonitor.DisableRandomIntervals()
			}
		}

		// Return success
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Heartbeat settings updated")

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleStatus handles the /api/status endpoint
func (h *APIHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// Update client status
		var data struct {
			ClientID     string `json:"clientId"`
			Status       string `json:"status"`
			ErrorMessage string `json:"errorMessage,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate the status
		var status client.ClientStatus
		switch data.Status {
		case "online":
			status = client.StatusOnline
		case "offline":
			status = client.StatusOffline
		case "busy":
			status = client.StatusBusy
		case "error":
			status = client.StatusError
		default:
			http.Error(w, "Invalid status", http.StatusBadRequest)
			return
		}

		// Update the client status
		err := h.clientManager.UpdateClientStatus(data.ClientID, status, data.ErrorMessage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Return success
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Client status updated")

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleExceptions handles the /api/exceptions endpoint
func (h *APIHandler) handleExceptions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Get all exceptions or filter by client ID
		clientID := r.URL.Query().Get("clientId")
		var exceptions []*client.ExceptionReport
		if clientID != "" {
			var err error
			exceptions, err = h.clientManager.GetExceptionReports(clientID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
		} else {
			exceptions = h.clientManager.GetAllExceptionReports()
		}

		// Return the exceptions as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(exceptions)

	case http.MethodPost:
		// Report a new exception
		var data struct {
			ClientID       string            `json:"clientId"`
			Message        string            `json:"message"`
			Severity       string            `json:"severity"`
			Module         string            `json:"module,omitempty"`
			StackTrace     string            `json:"stackTrace,omitempty"`
			AdditionalInfo map[string]string `json:"additionalInfo,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate the severity
		var severity client.ExceptionSeverity
		switch data.Severity {
		case "info":
			severity = client.SeverityInfo
		case "warning":
			severity = client.SeverityWarning
		case "error":
			severity = client.SeverityError
		case "critical":
			severity = client.SeverityCritical
		default:
			http.Error(w, "Invalid severity", http.StatusBadRequest)
			return
		}

		// Report the exception
		report, err := h.clientManager.ReportException(
			data.ClientID,
			data.Message,
			severity,
			data.Module,
			data.StackTrace,
			data.AdditionalInfo,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Return the report as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(report)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleException handles the /api/exceptions/{id} endpoint
func (h *APIHandler) handleException(w http.ResponseWriter, r *http.Request) {
	// Extract the exception ID from the URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid exception ID", http.StatusBadRequest)
		return
	}
	exceptionID := parts[len(parts)-1]

	switch r.Method {
	case http.MethodGet:
	// Get all exceptions
		allExceptions := h.clientManager.GetAllExceptionReports()
		var report *client.ExceptionReport
		exists := false
		
		// Find the exception with the matching ID
		for _, exc := range allExceptions {
			if exc.ID == exceptionID {
				report = exc
				exists = true
				break
			}
		}
		
		if !exists {
			http.Error(w, "Exception not found", http.StatusNotFound)
			return
		}

		// Return the exception as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(report)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
