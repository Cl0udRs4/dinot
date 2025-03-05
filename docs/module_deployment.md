# Dynamic Module Deployment

The C2 system supports dynamic module deployment, allowing modules to be deployed to clients after they have been generated, even if those modules were not initially selected during client generation.

## Module Architecture

The C2 system uses a modular architecture where modules are implemented as Go plugins. Each module implements the `Module` interface defined in `internal/client/module/module.go`:

```go
// Module defines the interface for client modules
type Module interface {
    // GetName returns the module name
    GetName() string
    
    // Execute executes the module with the given parameters
    Execute(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
    
    // Init initializes the module
    Init() error
    
    // Cleanup performs cleanup when the module is unloaded
    Cleanup() error
}
```

## Module Deployment Process

1. **Module Development**: Modules are developed as Go plugins that implement the `Module` interface.
2. **Module Compilation**: Modules are compiled into shared object files (.so) using the Go plugin package.
3. **Module Storage**: Compiled modules are stored in the module repository on the server.
4. **Module Deployment**: Modules are deployed to clients using the API or console interface.
5. **Module Loading**: Clients receive the module binary and load it using the `ModuleManager.LoadModuleFromBytes` method.
6. **Module Execution**: Once loaded, modules can be executed using the `ModuleManager.ExecuteModule` method.
7. **Module Unloading**: Modules can be unloaded using the `ModuleManager.UnloadModule` method.

## Cross-Platform Module Loading

The C2 system supports cross-platform module loading, with platform-specific implementations for different operating systems:

- **Linux/macOS**: Uses Go's plugin package to dynamically load modules from shared object files.
- **Windows**: Uses a registry-based approach to load pre-compiled modules, as Go's plugin package is not supported on Windows.

The `ModuleManager.LoadModuleFromBytes` method handles platform detection and uses the appropriate loading mechanism based on the client's operating system.

## API Endpoints

The following API endpoints are available for module management:

- `GET /api/modules`: List all available modules
- `GET /api/modules/{name}`: Get details about a specific module
- `GET /api/clients/{id}/modules`: List modules loaded on a client
- `GET /api/clients/{id}/modules/{name}`: Get details about a specific module on a client
- `PUT /api/clients/{id}/modules/{name}`: Load a module on a client
- `POST /api/clients/{id}/modules/{name}`: Execute a module on a client
- `DELETE /api/clients/{id}/modules/{name}`: Unload a module from a client

## Example: Loading and Executing a Shell Module

```bash
# Load the shell module on a client
curl -X PUT http://localhost:8090/api/clients/client-123/modules/shell

# Execute the shell module to run a command
curl -X POST http://localhost:8090/api/clients/client-123/modules/shell \
    -H "Content-Type: application/json" \
    -d '{"command": "echo hello world"}'
```

## Module Repository

The module repository is responsible for storing and retrieving module binaries. It provides the following functionality:

- **Store Module**: Stores a module binary in the repository
- **Get Module**: Retrieves a module binary from the repository
- **List Modules**: Lists all modules in the repository
- **Delete Module**: Deletes a module from the repository

The module repository is implemented in `internal/server/module/repository.go`.

## Security Considerations

- **Module Signing**: Modules are signed by the server to prevent unauthorized modules from being loaded
- **Signature Verification**: Module signatures are verified before loading to ensure integrity
- **Sandboxed Execution**: Module execution is sandboxed to prevent unauthorized access to the client system
- **Authentication and Authorization**: Module loading and execution require proper authentication and authorization
- **Secure Communication**: Module binaries are transmitted over encrypted channels to prevent interception

## Error Handling

The module deployment system includes comprehensive error handling to ensure reliability:

- **Module Not Found**: Proper error handling when a requested module is not found in the repository
- **Module Already Loaded**: Prevents loading a module that is already loaded on a client
- **Loading Failures**: Handles failures during module loading, with appropriate error messages
- **Execution Failures**: Handles failures during module execution, with appropriate error messages
- **Unloading Failures**: Handles failures during module unloading, with appropriate error messages

## Integration Testing

The module deployment system includes integration tests to verify functionality:

- **Module Loading**: Tests loading modules on clients
- **Module Execution**: Tests executing loaded modules on clients
- **Module Unloading**: Tests unloading modules from clients

The integration tests are implemented in `tests/integration/dynamic_module_test.go`.

## Future Enhancements

- **Module Versioning**: Support for module versioning to manage different versions of the same module
- **Module Dependencies**: Support for module dependencies to manage modules that depend on other modules
- **Module Updates**: Support for updating modules without unloading and reloading
- **Module Marketplace**: A marketplace for sharing and discovering modules
- **Module Telemetry**: Telemetry for monitoring module usage and performance
