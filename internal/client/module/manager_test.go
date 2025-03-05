package module

import (
    "context"
    "encoding/json"
    "testing"
)

// MockModule is a mock implementation of the Module interface for testing
type MockModule struct {
    BaseModule
    initFunc    func() error
    executeFunc func(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
    cleanupFunc func() error
}

func (m *MockModule) Init() error {
    if m.initFunc != nil {
        return m.initFunc()
    }
    return nil
}

func (m *MockModule) Execute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
    if m.executeFunc != nil {
        return m.executeFunc(ctx, params)
    }
    return nil, nil
}

func (m *MockModule) Cleanup() error {
    if m.cleanupFunc != nil {
        return m.cleanupFunc()
    }
    return nil
}

func TestModuleManager(t *testing.T) {
    // Create a module manager
    manager := NewModuleManager()
    
    // Create a mock module
    module := &MockModule{
        BaseModule: BaseModule{
            Name: "test-module",
        },
    }
    
    // Load the module
    err := manager.LoadModule(module)
    if err != nil {
        t.Fatalf("Failed to load module: %v", err)
    }
    
    // Verify the module was loaded
    modules := manager.GetModules()
    if len(modules) != 1 || modules[0] != "test-module" {
        t.Errorf("Expected modules to contain 'test-module', got %v", modules)
    }
    
    // Try to load the same module again
    err = manager.LoadModule(module)
    if err != ErrModuleAlreadyLoaded {
        t.Errorf("Expected ErrModuleAlreadyLoaded, got %v", err)
    }
    
    // Execute the module
    ctx := context.Background()
    params := json.RawMessage(`{"param1": "value1"}`)
    
    // Set up the execute function
    expectedResult := json.RawMessage(`{"result": "success"}`)
    module.executeFunc = func(ctx context.Context, p json.RawMessage) (json.RawMessage, error) {
        return expectedResult, nil
    }
    
    result, err := manager.ExecuteModule(ctx, "test-module", params)
    if err != nil {
        t.Fatalf("Failed to execute module: %v", err)
    }
    
    // Verify the result
    if string(result) != string(expectedResult) {
        t.Errorf("Expected result to be %s, got %s", expectedResult, result)
    }
    
    // Try to execute a non-existent module
    _, err = manager.ExecuteModule(ctx, "non-existent", params)
    if err != ErrModuleNotFound {
        t.Errorf("Expected ErrModuleNotFound, got %v", err)
    }
    
    // Unload the module
    err = manager.UnloadModule("test-module")
    if err != nil {
        t.Fatalf("Failed to unload module: %v", err)
    }
    
    // Verify the module was unloaded
    modules = manager.GetModules()
    if len(modules) != 0 {
        t.Errorf("Expected modules to be empty, got %v", modules)
    }
    
    // Try to unload a non-existent module
    err = manager.UnloadModule("non-existent")
    if err != ErrModuleNotFound {
        t.Errorf("Expected ErrModuleNotFound, got %v", err)
    }
}

func TestModuleManagerDynamicLoading(t *testing.T) {
    // This test is a placeholder for testing dynamic module loading
    // In a real implementation, you would need to create a test plugin
    // and load it using the LoadModuleFromBytes method
    
    // For now, we'll just verify that the IsModuleLoaded method works
    manager := NewModuleManager()
    
    // Check if a non-existent module is loaded
    if manager.IsModuleLoaded("non-existent") {
        t.Error("Expected IsModuleLoaded to return false for non-existent module")
    }
    
    // Load a module
    module := &MockModule{
        BaseModule: BaseModule{
            Name: "test-module",
        },
    }
    
    err := manager.LoadModule(module)
    if err != nil {
        t.Fatalf("Failed to load module: %v", err)
    }
    
    // Check if the module is loaded
    if !manager.IsModuleLoaded("test-module") {
        t.Error("Expected IsModuleLoaded to return true for loaded module")
    }
}
