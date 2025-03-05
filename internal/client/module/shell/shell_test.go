package shell

import (
    "context"
    "encoding/json"
    "testing"
)

func TestShellModule(t *testing.T) {
    // Create a shell module
    module := NewModule()
    
    // Initialize the module
    err := module.Init()
    if err != nil {
        t.Fatalf("Failed to initialize module: %v", err)
    }
    
    // Test executing a simple command
    ctx := context.Background()
    params := json.RawMessage(`{"command": "echo test"}`)
    
    result, err := module.Execute(ctx, params)
    if err != nil {
        t.Fatalf("Failed to execute command: %v", err)
    }
    
    // Parse the result
    var shellResult ShellResult
    err = json.Unmarshal(result, &shellResult)
    if err != nil {
        t.Fatalf("Failed to parse result: %v", err)
    }
    
    // Verify the command executed successfully
    if !shellResult.Success {
        t.Errorf("Expected command to succeed, got error: %s", shellResult.Error)
    }
    
    // Verify the output contains the expected text
    if shellResult.Output != "test\n" {
        t.Errorf("Expected output to be 'test\\n', got '%s'", shellResult.Output)
    }
    
    // Test executing a command with an error
    params = json.RawMessage(`{"command": "non_existent_command"}`)
    
    result, err = module.Execute(ctx, params)
    if err != nil {
        t.Fatalf("Failed to execute command: %v", err)
    }
    
    // Parse the result
    err = json.Unmarshal(result, &shellResult)
    if err != nil {
        t.Fatalf("Failed to parse result: %v", err)
    }
    
    // Verify the command failed
    if shellResult.Success {
        t.Error("Expected command to fail")
    }
    
    // Verify the error is not empty
    if shellResult.Error == "" {
        t.Error("Expected error message, got empty string")
    }
    
    // Test with invalid JSON
    params = json.RawMessage(`{"invalid json`)
    
    _, err = module.Execute(ctx, params)
    if err == nil {
        t.Error("Expected error for invalid JSON, got nil")
    }
    
    // Test cleanup
    err = module.Cleanup()
    if err != nil {
        t.Fatalf("Failed to cleanup module: %v", err)
    }
}
