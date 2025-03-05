package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/client"
)

// setupTestAPI creates a test API with mock data
func setupTestAPI() (*APIHandler, *client.ClientManager, *client.HeartbeatMonitor) {
	clientManager := client.NewClientManager()
	heartbeatMonitor := client.NewHeartbeatMonitor(clientManager, 30*time.Second, 60*time.Second)
	
	// Create a test client
	testClient := client.NewClient(
		"test-client-id",
		"Test Client",
		"192.168.1.100",
		"Linux",
		"x86_64",
		[]string{"shell", "file", "process"},
		"tcp",
	)
	
	// Register the test client
	_ = clientManager.RegisterClient(testClient)
	
	config := Config{
		Address:      "127.0.0.1:8080",
		AuthEnabled:  false,
		AuthUser:     "",
		AuthPassword: "",
		JWTSecret:    "",
		JWTEnabled:   false,
	}
	
	apiHandler := NewAPIHandler(clientManager, heartbeatMonitor, config)
	
	return apiHandler, clientManager, heartbeatMonitor
}

// TestGetClients tests the GET /api/clients endpoint
func TestGetClients(t *testing.T) {
	apiHandler, _, _ := setupTestAPI()
	
	// Create a request to get all clients
	req, err := http.NewRequest("GET", "/api/clients", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create a response recorder
	rr := httptest.NewRecorder()
	
	// Call the handler
	handler := http.HandlerFunc(apiHandler.handleClients)
	handler.ServeHTTP(rr, req)
	
	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	// Check the response body
	var clients []*client.Client
	err = json.Unmarshal(rr.Body.Bytes(), &clients)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(clients) != 1 {
		t.Errorf("expected 1 client, got %d", len(clients))
	}
	
	if clients[0].ID != "test-client-id" {
		t.Errorf("expected client ID test-client-id, got %s", clients[0].ID)
	}
}

// TestGetClientByID tests the GET /api/clients/{id} endpoint
func TestGetClientByID(t *testing.T) {
	apiHandler, _, _ := setupTestAPI()
	
	// Create a request to get a client by ID
	req, err := http.NewRequest("GET", "/api/clients/test-client-id", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create a response recorder
	rr := httptest.NewRecorder()
	
	// Call the handler
	handler := http.HandlerFunc(apiHandler.handleClient)
	handler.ServeHTTP(rr, req)
	
	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	// Check the response body
	var client client.Client
	err = json.Unmarshal(rr.Body.Bytes(), &client)
	if err != nil {
		t.Fatal(err)
	}
	
	if client.ID != "test-client-id" {
		t.Errorf("expected client ID test-client-id, got %s", client.ID)
	}
}

// TestUpdateClientStatus tests the POST /api/status endpoint
func TestUpdateClientStatus(t *testing.T) {
	apiHandler, _, _ := setupTestAPI()
	
	// Create a request to update a client's status
	data := map[string]string{
		"clientId": "test-client-id",
		"status":   "busy",
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	
	req, err := http.NewRequest("POST", "/api/status", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Create a response recorder
	rr := httptest.NewRecorder()
	
	// Call the handler
	handler := http.HandlerFunc(apiHandler.handleStatus)
	handler.ServeHTTP(rr, req)
	
	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	// Verify the client's status was updated
	client, err := apiHandler.clientManager.GetClient("test-client-id")
	if err != nil {
		t.Fatal(err)
	}
	
	if client.Status != client.StatusBusy {
		t.Errorf("expected client status busy, got %s", client.Status)
	}
}

// TestGetHeartbeatSettings tests the GET /api/heartbeat endpoint
func TestGetHeartbeatSettings(t *testing.T) {
	apiHandler, _, _ := setupTestAPI()
	
	// Create a request to get heartbeat settings
	req, err := http.NewRequest("GET", "/api/heartbeat", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create a response recorder
	rr := httptest.NewRecorder()
	
	// Call the handler
	handler := http.HandlerFunc(apiHandler.handleHeartbeat)
	handler.ServeHTTP(rr, req)
	
	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	// Check the response body
	var settings map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &settings)
	if err != nil {
		t.Fatal(err)
	}
	
	if settings["checkInterval"] == nil {
		t.Errorf("expected checkInterval to be set")
	}
	
	if settings["timeout"] == nil {
		t.Errorf("expected timeout to be set")
	}
}

// TestUpdateHeartbeatSettings tests the POST /api/heartbeat endpoint
func TestUpdateHeartbeatSettings(t *testing.T) {
	apiHandler, _, heartbeatMonitor := setupTestAPI()
	
	// Create a request to update heartbeat settings
	data := map[string]interface{}{
		"checkInterval": 45.0,
		"timeout":       90.0,
		"randomEnabled": true,
		"randomMinInterval": 5.0,
		"randomMaxInterval": 300.0,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	
	req, err := http.NewRequest("POST", "/api/heartbeat", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Create a response recorder
	rr := httptest.NewRecorder()
	
	// Call the handler
	handler := http.HandlerFunc(apiHandler.handleHeartbeat)
	handler.ServeHTTP(rr, req)
	
	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	// Verify the heartbeat settings were updated
	if heartbeatMonitor.GetCheckInterval() != 45*time.Second {
		t.Errorf("expected check interval 45s, got %s", heartbeatMonitor.GetCheckInterval())
	}
	
	if heartbeatMonitor.GetTimeout() != 90*time.Second {
		t.Errorf("expected timeout 90s, got %s", heartbeatMonitor.GetTimeout())
	}
	
	if !heartbeatMonitor.IsRandomEnabled() {
		t.Errorf("expected random intervals to be enabled")
	}
	
	if heartbeatMonitor.GetRandomMinInterval() != 5*time.Second {
		t.Errorf("expected random min interval 5s, got %s", heartbeatMonitor.GetRandomMinInterval())
	}
	
	if heartbeatMonitor.GetRandomMaxInterval() != 300*time.Second {
		t.Errorf("expected random max interval 300s, got %s", heartbeatMonitor.GetRandomMaxInterval())
	}
}

// TestAuthMiddleware tests the authentication middleware
func TestAuthMiddleware(t *testing.T) {
	clientManager := client.NewClientManager()
	heartbeatMonitor := client.NewHeartbeatMonitor(clientManager, 30*time.Second, 60*time.Second)
	
	config := Config{
		Address:      "127.0.0.1:8080",
		AuthEnabled:  true,
		AuthUser:     "admin",
		AuthPassword: "password",
		JWTSecret:    "secret",
		JWTEnabled:   true,
	}
	
	apiHandler := NewAPIHandler(clientManager, heartbeatMonitor, config)
	
	// Create a test handler
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}
	
	// Wrap the test handler with the auth middleware
	handler := apiHandler.authMiddleware(testHandler)
	
	// Test with no authentication
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
	
	// Test with basic authentication
	req, _ = http.NewRequest("GET", "/", nil)
	req.SetBasicAuth("admin", "password")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	// Test with JWT authentication
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
