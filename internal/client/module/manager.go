package module

import (
    "context"
    "encoding/json"
    "errors"
    "sync"
)

var (
    // ErrModuleNotFound is returned when a module is not found
    ErrModuleNotFound = errors.New("module not found")
    
    // ErrModuleAlreadyLoaded is returned when a module is already loaded
    ErrModuleAlreadyLoaded = errors.New("module already loaded")
)

// ModuleManager manages client modules
type ModuleManager struct {
    modules map[string]Module
    mu      sync.RWMutex
}

// NewModuleManager creates a new module manager
func NewModuleManager() *ModuleManager {
    return &ModuleManager{
        modules: make(map[string]Module),
    }
}

// LoadModule loads a module
func (m *ModuleManager) LoadModule(module Module) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    name := module.GetName()
    if _, exists := m.modules[name]; exists {
        return ErrModuleAlreadyLoaded
    }
    
    err := module.Init()
    if err != nil {
        return err
    }
    
    m.modules[name] = module
    return nil
}

// UnloadModule unloads a module
func (m *ModuleManager) UnloadModule(name string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    module, exists := m.modules[name]
    if !exists {
        return ErrModuleNotFound
    }
    
    err := module.Cleanup()
    if err != nil {
        return err
    }
    
    delete(m.modules, name)
    return nil
}

// ExecuteModule executes a module
func (m *ModuleManager) ExecuteModule(ctx context.Context, name string, params json.RawMessage) (json.RawMessage, error) {
    m.mu.RLock()
    module, exists := m.modules[name]
    m.mu.RUnlock()
    
    if !exists {
        return nil, ErrModuleNotFound
    }
    
    return module.Execute(ctx, params)
}

// GetModules returns a list of loaded modules
func (m *ModuleManager) GetModules() []string {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    modules := make([]string, 0, len(m.modules))
    for name := range m.modules {
        modules = append(modules, name)
    }
    
    return modules
}

// GetModule returns a module by name
func (m *ModuleManager) GetModule(name string) (Module, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    module, exists := m.modules[name]
    if !exists {
        return nil, ErrModuleNotFound
    }
    
    return module, nil
}
