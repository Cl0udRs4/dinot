package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	
	"github.com/Cl0udRs4/dinot/internal/server/client"
	"github.com/google/uuid"
)

// ModuleInfo represents information about a module
type ModuleInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Parameters  []string `json:"parameters,omitempty"`
}

// handleModules handles the /api/modules endpoint
func (h *APIHandler) handleModules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Get all available modules
		modules := []ModuleInfo{
			{Name: "shell", Description: "Execute shell commands", Parameters: []string{"command"}},
			{Name: "file", Description: "File operations", Parameters: []string{"path", "operation"}},
			{Name: "process", Description: "Process management", Parameters: []string{"pid", "action"}},
			{Name: "network", Description: "Network operations", Parameters: []string{"host", "port", "protocol"}},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(modules)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleModule handles the /api/modules/{name} endpoint
func (h *APIHandler) handleModule(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid module name", http.StatusBadRequest)
		return
	}
	
	moduleName := parts[3]
	
	switch r.Method {
	case http.MethodGet:
		// Get module details based on name
		var module ModuleInfo
		
		switch moduleName {
		case "shell":
			module = ModuleInfo{
				Name:        "shell",
				Description: "Execute shell commands on the client",
				Parameters:  []string{"command", "timeout"},
			}
		case "file":
			module = ModuleInfo{
				Name:        "file",
				Description: "Perform file operations on the client",
				Parameters:  []string{"path", "operation", "content"},
			}
		case "process":
			module = ModuleInfo{
				Name:        "process",
				Description: "Manage processes on the client",
				Parameters:  []string{"pid", "action", "priority"},
			}
		case "network":
			module = ModuleInfo{
				Name:        "network",
				Description: "Perform network operations on the client",
				Parameters:  []string{"host", "port", "protocol", "timeout"},
			}
		default:
			http.Error(w, "Module not found", http.StatusNotFound)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(module)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleClientModules handles client module operations
// This function handles both /api/clients/{id}/modules and /api/clients/{id}/modules/{name} endpoints
func (h *APIHandler) handleClientModules(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	
	// Check if this is a client modules request
	if len(parts) < 5 || parts[3] == "" || parts[4] != "modules" {
		return // Not a client modules request, let other handlers process it
	}
	
	clientID := parts[3]
	
	// Get client
	client, err := h.clientManager.GetClient(clientID)
	if err != nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}
	
	// Check if this is a request for a specific module
	if len(parts) >= 6 && parts[5] != "" {
		moduleName := parts[5]
		h.handleClientModule(w, r, client, moduleName)
		return
	}
	
	// Handle client modules list
	switch r.Method {
	case http.MethodGet:
		// Get client modules
		// In a real implementation, this would get the modules from the client
		// For now, we'll return a placeholder list
		modules := []string{"shell", "file"}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"client_id": clientID,
			"modules":   modules,
		})
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleClientModule handles operations on a specific client module
func (h *APIHandler) handleClientModule(w http.ResponseWriter, r *http.Request, client *client.Client, moduleName string) {
	switch r.Method {
	case http.MethodGet:
		// Get module status
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"client_id": client.ID,
			"module":    moduleName,
			"active":    client.IsModuleActive(moduleName),
		})
		
	case http.MethodPost:
		// Execute module
		var params map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		// Send command to client to execute module
		command := map[string]interface{}{
			"type":       "execute_module",
			"module":     moduleName,
			"command_id": uuid.New().String(),
			"params":     params,
		}
		
		// Send command to client (implementation depends on your communication system)
		err := h.sendCommandToClient(client.ID, command)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to send command: %v", err), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"client_id":  client.ID,
			"module":     moduleName,
			"command_id": command["command_id"],
			"status":     "command_sent",
		})
		
	case http.MethodPut:
		// Load module
		// Get module binary from module repository
		moduleBytes, err := h.getModuleBinary(moduleName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get module binary: %v", err), http.StatusInternalServerError)
			return
		}
		
		// Verify module signature
		err = h.securityManager.VerifyModuleSignature(moduleName, moduleBytes, nil) // Signature should be retrieved separately
		if err != nil {
			http.Error(w, fmt.Sprintf("Module signature verification failed: %v", err), http.StatusForbidden)
			return
		}
		
		// Send command to client to load module
		command := map[string]interface{}{
			"type":         "load_module",
			"module":       moduleName,
			"command_id":   uuid.New().String(),
			"module_bytes": moduleBytes,
		}
		
		// Send command to client
		err = h.sendCommandToClient(client.ID, command)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to send command: %v", err), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"client_id":  client.ID,
			"module":     moduleName,
			"command_id": command["command_id"],
			"status":     "load_command_sent",
		})
		
	case http.MethodDelete:
		// Unload module
		command := map[string]interface{}{
			"type":       "unload_module",
			"module":     moduleName,
			"command_id": uuid.New().String(),
		}
		
		// Send command to client
		err := h.sendCommandToClient(client.ID, command)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to send command: %v", err), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"client_id":  client.ID,
			"module":     moduleName,
			"command_id": command["command_id"],
			"status":     "unload_command_sent",
		})
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
