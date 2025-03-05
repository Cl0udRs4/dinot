package client

import (
    "encoding/json"
    "testing"
    "time"
    
    "github.com/Cl0udRs4/dinot/internal/client/protocol"
)

// TestFeedbackMechanism tests the feedback mechanism
func TestFeedbackMechanism(t *testing.T) {
    // Create a mock protocol
    mockProto := &MockProtocol{
        BaseProtocol: protocol.BaseProtocol{
            Name: "tcp",
        },
        connected: true,
    }
    
    // Create a client config
    config := Config{
        ID:                     "test-client",
        Name:                   "Test Client",
        ServerAddresses:        map[string]string{"tcp": "tcp://localhost:8080"},
        HeartbeatInterval:      100 * time.Millisecond,
        ProtocolSwitchThreshold: 3,
    }
    
    // Create a client with the mock protocol
    client, err := NewClient(config)
    if err != nil {
        t.Fatalf("Failed to create client: %v", err)
    }
    
    // Replace the protocol manager with one that uses our mock
    protocols := []protocol.Protocol{mockProto}
    client.protocolMgr = protocol.NewProtocolManager(protocols, config.ProtocolSwitchThreshold)
    
    // Track sent feedback messages
    var sentFeedback []FeedbackResponse
    mockProto.sendFunc = func(data []byte) error {
        var feedback map[string]interface{}
        err := json.Unmarshal(data, &feedback)
        if err != nil {
            return err
        }
        
        if feedback["type"] == "module_result" || 
           feedback["type"] == "module_load_result" || 
           feedback["type"] == "module_unload_result" {
            var response FeedbackResponse
            err = json.Unmarshal(data, &response)
            if err != nil {
                return err
            }
            sentFeedback = append(sentFeedback, response)
        }
        
        return nil
    }
    
    // Start the client
    err = client.Start()
    if err != nil {
        t.Fatalf("Failed to start client: %v", err)
    }
    
    // Test handleExecuteModule
    commandParams := json.RawMessage(`{"command_id": "test-command-1", "command": "echo test"}`)
    client.handleExecuteModule("shell", commandParams)
    
    // Wait for feedback to be sent
    time.Sleep(100 * time.Millisecond)
    
    // Verify feedback
    if len(sentFeedback) < 2 {
        t.Fatalf("Expected at least 2 feedback messages, got %d", len(sentFeedback))
    }
    
    // Verify initial feedback
    initialFeedback := sentFeedback[0]
    if initialFeedback.Type != "module_result" {
        t.Errorf("Expected type to be 'module_result', got '%s'", initialFeedback.Type)
    }
    if initialFeedback.Status != "processing" {
        t.Errorf("Expected status to be 'processing', got '%s'", initialFeedback.Status)
    }
    
    // Verify final feedback
    finalFeedback := sentFeedback[len(sentFeedback)-1]
    if finalFeedback.Type != "module_result" {
        t.Errorf("Expected type to be 'module_result', got '%s'", finalFeedback.Type)
    }
    if finalFeedback.CommandID != "test-command-1" {
        t.Errorf("Expected command_id to be 'test-command-1', got '%s'", finalFeedback.CommandID)
    }
    
    // Stop the client
    err = client.Stop()
    if err != nil {
        t.Fatalf("Failed to stop client: %v", err)
    }
}

// TestFeedbackRetryLogic tests the retry logic in the feedback mechanism
func TestFeedbackRetryLogic(t *testing.T) {
    // Create a mock protocol
    mockProto := &MockProtocol{
        BaseProtocol: protocol.BaseProtocol{
            Name: "tcp",
        },
        connected: true,
    }
    
    // Create a client config
    config := Config{
        ID:                     "test-client",
        Name:                   "Test Client",
        ServerAddresses:        map[string]string{"tcp": "tcp://localhost:8080"},
        HeartbeatInterval:      100 * time.Millisecond,
        ProtocolSwitchThreshold: 3,
    }
    
    // Create a client with the mock protocol
    client, err := NewClient(config)
    if err != nil {
        t.Fatalf("Failed to create client: %v", err)
    }
    
    // Configure shorter retry intervals for testing
    client.feedbackConfig.RetryInterval = 50 * time.Millisecond
    client.feedbackConfig.MaxRetryInterval = 200 * time.Millisecond
    
    // Replace the protocol manager with one that uses our mock
    protocols := []protocol.Protocol{mockProto}
    client.protocolMgr = protocol.NewProtocolManager(protocols, config.ProtocolSwitchThreshold)
    
    // Track sent feedback messages
    var sentFeedback []FeedbackResponse
    failCount := 0
    maxFails := 2
    
    mockProto.sendFunc = func(data []byte) error {
        if failCount < maxFails {
            failCount++
            return protocol.ErrSendFailed
        }
        
        var feedback map[string]interface{}
        err := json.Unmarshal(data, &feedback)
        if err != nil {
            return err
        }
        
        if feedback["type"] == "module_result" {
            var response FeedbackResponse
            err = json.Unmarshal(data, &response)
            if err != nil {
                return err
            }
            sentFeedback = append(sentFeedback, response)
        }
        
        return nil
    }
    
    // Start the client
    err = client.Start()
    if err != nil {
        t.Fatalf("Failed to start client: %v", err)
    }
    
    // Test sendFeedback with retry
    response := FeedbackResponse{
        Type:      "module_result",
        ClientID:  client.config.ID,
        CommandID: "test-retry",
        Module:    "test",
        Success:   true,
        Status:    "completed",
        Timestamp: time.Now().Unix(),
    }
    
    err = client.sendFeedback(response)
    
    // Verify that the feedback was eventually sent
    if err != nil {
        t.Errorf("Expected sendFeedback to succeed after retries, got error: %v", err)
    }
    
    // Verify that the correct number of retries were attempted
    if failCount != maxFails {
        t.Errorf("Expected %d failed attempts, got %d", maxFails, failCount)
    }
    
    // Stop the client
    err = client.Stop()
    if err != nil {
        t.Fatalf("Failed to stop client: %v", err)
    }
}
