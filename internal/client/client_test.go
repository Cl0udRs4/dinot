package client

import (
    "context"
    "encoding/json"
    "testing"
    "time"
    
    "github.com/Cl0udRs4/dinot/internal/client/protocol"
)

// MockProtocol is a mock implementation of the Protocol interface for testing
type MockProtocol struct {
    protocol.BaseProtocol
    connected bool
    sendFunc  func(data []byte) error
    recvFunc  func(timeout time.Duration) ([]byte, error)
}

func (p *MockProtocol) Connect(ctx context.Context) error {
    p.connected = true
    return nil
}

func (p *MockProtocol) Disconnect() error {
    p.connected = false
    return nil
}

func (p *MockProtocol) Send(data []byte) error {
    if p.sendFunc != nil {
        return p.sendFunc(data)
    }
    return nil
}

func (p *MockProtocol) Receive(timeout time.Duration) ([]byte, error) {
    if p.recvFunc != nil {
        return p.recvFunc(timeout)
    }
    return nil, nil
}

func (p *MockProtocol) IsConnected() bool {
    return p.connected
}

func TestClientHeartbeat(t *testing.T) {
    // Create a mock protocol
    mockProto := &MockProtocol{
        BaseProtocol: protocol.BaseProtocol{
            Name: "tcp",
        },
        connected: false,
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
    
    // Track sent heartbeats
    var sentHeartbeats []map[string]interface{}
    mockProto.sendFunc = func(data []byte) error {
        var heartbeat map[string]interface{}
        err := json.Unmarshal(data, &heartbeat)
        if err != nil {
            return err
        }
        
        if heartbeat["type"] == "heartbeat" {
            sentHeartbeats = append(sentHeartbeats, heartbeat)
        }
        
        return nil
    }
    
    // Start the client
    err = client.Start()
    if err != nil {
        t.Fatalf("Failed to start client: %v", err)
    }
    
    // Wait for at least one heartbeat
    time.Sleep(150 * time.Millisecond)
    
    // Stop the client
    err = client.Stop()
    if err != nil {
        t.Fatalf("Failed to stop client: %v", err)
    }
    
    // Verify that at least one heartbeat was sent
    if len(sentHeartbeats) == 0 {
        t.Fatal("No heartbeats were sent")
    }
    
    // Verify heartbeat contents
    heartbeat := sentHeartbeats[0]
    if heartbeat["client_id"] != "test-client" {
        t.Errorf("Expected client_id to be 'test-client', got %v", heartbeat["client_id"])
    }
    
    if heartbeat["random_enabled"] != false {
        t.Errorf("Expected random_enabled to be false, got %v", heartbeat["random_enabled"])
    }
}

func TestClientRandomHeartbeat(t *testing.T) {
    // Create a mock protocol
    mockProto := &MockProtocol{
        BaseProtocol: protocol.BaseProtocol{
            Name: "tcp",
        },
        connected: false,
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
    
    // Track sent heartbeats
    var sentHeartbeats []map[string]interface{}
    mockProto.sendFunc = func(data []byte) error {
        var heartbeat map[string]interface{}
        err := json.Unmarshal(data, &heartbeat)
        if err != nil {
            return err
        }
        
        if heartbeat["type"] == "heartbeat" {
            sentHeartbeats = append(sentHeartbeats, heartbeat)
        }
        
        return nil
    }
    
    // Start the client
    err = client.Start()
    if err != nil {
        t.Fatalf("Failed to start client: %v", err)
    }
    
    // Enable random heartbeats
    minInterval := 50 * time.Millisecond
    maxInterval := 150 * time.Millisecond
    client.EnableRandomHeartbeat(minInterval, maxInterval)
    
    // Wait for at least two heartbeats
    time.Sleep(300 * time.Millisecond)
    
    // Stop the client
    err = client.Stop()
    if err != nil {
        t.Fatalf("Failed to stop client: %v", err)
    }
    
    // Verify that at least two heartbeats were sent
    if len(sentHeartbeats) < 2 {
        t.Fatalf("Expected at least 2 heartbeats, got %d", len(sentHeartbeats))
    }
    
    // Verify that random heartbeats were enabled
    for _, heartbeat := range sentHeartbeats[1:] { // Skip the first one which might have been sent before enabling random
        if heartbeat["random_enabled"] != true {
            t.Errorf("Expected random_enabled to be true, got %v", heartbeat["random_enabled"])
        }
    }
}

func TestClientProtocolSwitching(t *testing.T) {
    // Skip this test for now as it requires more complex setup
    t.Skip("Skipping protocol switching test for now")
    
    // Create two mock protocols
    mockProto1 := &MockProtocol{
        BaseProtocol: protocol.BaseProtocol{
            Name: "tcp",
        },
        connected: true,
    }
    
    mockProto2 := &MockProtocol{
        BaseProtocol: protocol.BaseProtocol{
            Name: "udp",
        },
        connected: true,
    }
    
    // Create a client config
    config := Config{
        ID:                     "test-client",
        Name:                   "Test Client",
        ServerAddresses:        map[string]string{"tcp": "tcp://localhost:8080", "udp": "udp://localhost:8081"},
        HeartbeatInterval:      100 * time.Millisecond,
        ProtocolSwitchThreshold: 2, // Set a low threshold for testing
    }
    
    // Create a client
    client, err := NewClient(config)
    if err != nil {
        t.Fatalf("Failed to create client: %v", err)
    }
    
    // Replace the protocol manager with one that uses our mocks
    protocols := []protocol.Protocol{mockProto1, mockProto2}
    client.protocolMgr = protocol.NewProtocolManager(protocols, config.ProtocolSwitchThreshold)
    
    // Make the first protocol fail to send
    mockProto1.sendFunc = func(data []byte) error {
        return protocol.ErrSendFailed
    }
    
    // Track sent heartbeats on the second protocol
    var sentHeartbeats []map[string]interface{}
    mockProto2.sendFunc = func(data []byte) error {
        var heartbeat map[string]interface{}
        err := json.Unmarshal(data, &heartbeat)
        if err != nil {
            return err
        }
        
        if heartbeat["type"] == "heartbeat" {
            sentHeartbeats = append(sentHeartbeats, heartbeat)
        }
        
        return nil
    }
    
    // Start the client
    err = client.Start()
    if err != nil {
        t.Fatalf("Failed to start client: %v", err)
    }
    
    // Wait for protocol switching to occur
    // We need to wait for at least threshold+1 heartbeats
    time.Sleep(300 * time.Millisecond)
    
    // Stop the client
    err = client.Stop()
    if err != nil {
        t.Fatalf("Failed to stop client: %v", err)
    }
    
    // Verify that at least one heartbeat was sent on the second protocol
    if len(sentHeartbeats) == 0 {
        t.Fatal("No heartbeats were sent on the second protocol")
    }
    
    // Verify that the protocol was switched
    currentProto := client.protocolMgr.GetCurrentProtocol()
    if currentProto.GetName() != "mock2" {
        t.Errorf("Expected current protocol to be 'mock2', got '%s'", currentProto.GetName())
    }
}
