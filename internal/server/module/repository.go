package module

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "sync"
)

// Repository manages module storage and retrieval
type Repository struct {
    baseDir string
    mu      sync.RWMutex
}

// NewRepository creates a new module repository
func NewRepository(baseDir string) (*Repository, error) {
    // Create the base directory if it doesn't exist
    if err := os.MkdirAll(baseDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create module directory: %w", err)
    }
    
    return &Repository{
        baseDir: baseDir,
    }, nil
}

// StoreModule stores a module in the repository
func (r *Repository) StoreModule(name string, data []byte) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Create the module file path
    filePath := filepath.Join(r.baseDir, name+".so")
    
    // Write the module data to the file
    if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
        return fmt.Errorf("failed to write module file: %w", err)
    }
    
    return nil
}

// GetModule retrieves a module from the repository
func (r *Repository) GetModule(name string) ([]byte, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    // Create the module file path
    filePath := filepath.Join(r.baseDir, name+".so")
    
    // Read the module data from the file
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read module file: %w", err)
    }
    
    return data, nil
}

// ListModules lists all modules in the repository
func (r *Repository) ListModules() ([]string, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    // Get all module files
    files, err := filepath.Glob(filepath.Join(r.baseDir, "*.so"))
    if err != nil {
        return nil, fmt.Errorf("failed to list module files: %w", err)
    }
    
    // Extract module names
    modules := make([]string, 0, len(files))
    for _, file := range files {
        name := filepath.Base(file)
        name = name[:len(name)-3] // Remove .so extension
        modules = append(modules, name)
    }
    
    return modules, nil
}

// DeleteModule deletes a module from the repository
func (r *Repository) DeleteModule(name string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Create the module file path
    filePath := filepath.Join(r.baseDir, name+".so")
    
    // Delete the module file
    if err := os.Remove(filePath); err != nil {
        return fmt.Errorf("failed to delete module file: %w", err)
    }
    
    return nil
}
