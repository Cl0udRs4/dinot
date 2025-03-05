package integration

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "testing"
    "time"
)

// testConfig holds configuration for dynamic module deployment tests
type testConfig struct {
    ServerHost     string
    ServerPort     int
    Protocol       string
    NumClients     int
}

// TestDynamicModuleDeployment tests the dynamic module deployment functionality
func TestDynamicModuleDeployment(t *testing.T) {
    // Skip the test for now as it requires a running server
    t.Skip("Skipping dynamic module deployment test as it requires a running server")
    
    // Define server configuration
    serverConfig := testConfig{
        ServerHost:     "localhost",
        ServerPort:     8080,
        Protocol:       "tcp",
        NumClients:     1,
    }
    
    // Start the client
    client, err := startTestClient(serverConfig)
    if err != nil {
        t.Fatalf("Failed to start test client: %v", err)
    }
    defer stopTestClient(client)
    
    // Wait for client to register
    time.Sleep(2 * time.Second)
    
    // Get client ID
    clientID, err := getClientID()
    if err != nil {
        t.Fatalf("Failed to get client ID: %v", err)
    }
    
    // Test 1: List client modules
    t.Run("ListClientModules", func(t *testing.T) {
        resp, err := http.Get(fmt.Sprintf("http://%s:%d/api/clients/%s/modules", serverConfig.ServerHost, serverConfig.ServerPort+10, clientID))
        if err != nil {
            t.Fatalf("Failed to list client modules: %v", err)
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusOK {
            t.Fatalf("Expected status OK, got %s", resp.Status)
        }
        
        var result map[string]interface{}
        if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
            t.Fatalf("Failed to decode response: %v", err)
        }
        
        modules, ok := result["modules"].([]interface{})
        if !ok {
            t.Fatalf("Expected modules to be an array")
        }
        
        t.Logf("Initial modules: %v", modules)
    })
    
    // Test 2: Load a module
    t.Run("LoadModule", func(t *testing.T) {
        // Prepare request body
        body := map[string]interface{}{
            "module": "shell",
        }
        
        bodyBytes, err := json.Marshal(body)
        if err != nil {
            t.Fatalf("Failed to marshal request body: %v", err)
        }
        
        // Send request
        req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s:%d/api/clients/%s/modules/shell", serverConfig.ServerHost, serverConfig.ServerPort+10, clientID), bytes.NewBuffer(bodyBytes))
        if err != nil {
            t.Fatalf("Failed to create request: %v", err)
        }
        
        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            t.Fatalf("Failed to load module: %v", err)
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusOK {
            t.Fatalf("Expected status OK, got %s", resp.Status)
        }
        
        var result map[string]interface{}
        if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
            t.Fatalf("Failed to decode response: %v", err)
        }
        
        status, ok := result["status"].(string)
        if !ok || status != "load_command_sent" {
            t.Fatalf("Expected status 'load_command_sent', got %v", status)
        }
        
        // Wait for module to load
        time.Sleep(2 * time.Second)
    })
    
    // Test 3: Execute the loaded module
    t.Run("ExecuteModule", func(t *testing.T) {
        // Prepare request body
        body := map[string]interface{}{
            "command": "echo test",
        }
        
        bodyBytes, err := json.Marshal(body)
        if err != nil {
            t.Fatalf("Failed to marshal request body: %v", err)
        }
        
        // Send request
        resp, err := http.Post(fmt.Sprintf("http://%s:%d/api/clients/%s/modules/shell", serverConfig.ServerHost, serverConfig.ServerPort+10, clientID), "application/json", bytes.NewBuffer(bodyBytes))
        if err != nil {
            t.Fatalf("Failed to execute module: %v", err)
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusOK {
            t.Fatalf("Expected status OK, got %s", resp.Status)
        }
        
        var result map[string]interface{}
        if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
            t.Fatalf("Failed to decode response: %v", err)
        }
        
        status, ok := result["status"].(string)
        if !ok || status != "command_sent" {
            t.Fatalf("Expected status 'command_sent', got %v", status)
        }
    })
    
    // Test 4: Unload the module
    t.Run("UnloadModule", func(t *testing.T) {
        // Send request
        req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s:%d/api/clients/%s/modules/shell", serverConfig.ServerHost, serverConfig.ServerPort+10, clientID), nil)
        if err != nil {
            t.Fatalf("Failed to create request: %v", err)
        }
        
        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            t.Fatalf("Failed to unload module: %v", err)
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusOK {
            t.Fatalf("Expected status OK, got %s", resp.Status)
        }
        
        var result map[string]interface{}
        if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
            t.Fatalf("Failed to decode response: %v", err)
        }
        
        status, ok := result["status"].(string)
        if !ok || status != "unload_command_sent" {
            t.Fatalf("Expected status 'unload_command_sent', got %v", status)
        }
    })
}

// Helper functions for testing

// startTestClient starts a test client
func startTestClient(config testConfig) (interface{}, error) {
    // In a real implementation, this would start a client
    // For now, we'll return a placeholder
    return struct{}{}, nil
}

// stopTestClient stops a test client
func stopTestClient(client interface{}) {
    // In a real implementation, this would stop the client
    // For now, we'll do nothing
}

// getClientID gets the ID of the first registered client
func getClientID() (string, error) {
    // In a real implementation, this would get the client ID from the server
    // For now, we'll return a placeholder
    return "test-client-id", nil
}
