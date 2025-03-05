package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/client"
)

// Command represents a console command
type Command struct {
	// Name is the command name
	Name string
	
	// Description is a brief description of what the command does
	Description string
	
	// Usage is a usage example
	Usage string
	
	// Execute is the function that executes the command
	Execute func(args []string) error
}

// Console represents the command-line console interface
type Console struct {
	// clientManager is the client manager to interact with
	clientManager *client.ClientManager
	
	// heartbeatMonitor is the heartbeat monitor to interact with
	heartbeatMonitor *client.HeartbeatMonitor
	
	// commands is a map of command names to Command objects
	commands map[string]*Command
	
	// running indicates whether the console is running
	running bool
	
	// reader is the input reader
	reader *bufio.Reader
}

// NewConsole creates a new console interface
func NewConsole(clientManager *client.ClientManager, heartbeatMonitor *client.HeartbeatMonitor) *Console {
	console := &Console{
		clientManager:    clientManager,
		heartbeatMonitor: heartbeatMonitor,
		commands:         make(map[string]*Command),
		reader:           bufio.NewReader(os.Stdin),
	}
	
	// Register commands
	console.registerCommands()
	
	return console
}

// registerCommands registers all available commands
func (c *Console) registerCommands() {
	// Help command
	c.commands["help"] = &Command{
		Name:        "help",
		Description: "Display available commands",
		Usage:       "help",
		Execute:     c.cmdHelp,
	}
	
	// List clients command
	c.commands["list"] = &Command{
		Name:        "list",
		Description: "List all clients or filter by status",
		Usage:       "list [status]",
		Execute:     c.cmdList,
	}
	
	// Show client details command
	c.commands["info"] = &Command{
		Name:        "info",
		Description: "Show detailed information about a client",
		Usage:       "info <client_id>",
		Execute:     c.cmdInfo,
	}
	
	// Set client status command
	c.commands["status"] = &Command{
		Name:        "status",
		Description: "Set a client's status",
		Usage:       "status <client_id> <online|offline|busy|error> [error_message]",
		Execute:     c.cmdStatus,
	}
	
	// Heartbeat settings command
	c.commands["heartbeat"] = &Command{
		Name:        "heartbeat",
		Description: "Configure heartbeat settings",
		Usage:       "heartbeat <check|timeout|random> [args...]",
		Execute:     c.cmdHeartbeat,
	}
	
	// Exception management command
	c.commands["exception"] = &Command{
		Name:        "exception",
		Description: "Manage exception reports",
		Usage:       "exception <list|report> [args...]",
		Execute:     c.cmdException,
	}
	
	// Exit command
	c.commands["exit"] = &Command{
		Name:        "exit",
		Description: "Exit the console",
		Usage:       "exit",
		Execute:     c.cmdExit,
	}
}

// Start starts the console interface
func (c *Console) Start() {
	c.running = true
	
	fmt.Println("C2 Console Interface")
	fmt.Println("Type 'help' for available commands")
	
	for c.running {
		fmt.Print("> ")
		input, err := c.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// Handle EOF by exiting the loop
				fmt.Println("Input stream closed, exiting console")
				c.Stop()
				break
			}
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}
		
		// Trim whitespace and split into command and args
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		
		parts := strings.Fields(input)
		cmdName := parts[0]
		args := parts[1:]
		
		// Find and execute the command
		cmd, exists := c.commands[cmdName]
		if !exists {
			fmt.Printf("Unknown command: %s\n", cmdName)
			fmt.Println("Type 'help' for available commands")
			continue
		}
		
		if err := cmd.Execute(args); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

// Stop stops the console interface
func (c *Console) Stop() {
	c.running = false
}

// cmdHelp implements the help command
func (c *Console) cmdHelp(args []string) error {
	fmt.Println("Available commands:")
	
	// Get a sorted list of commands
	var cmdNames []string
	for name := range c.commands {
		cmdNames = append(cmdNames, name)
	}
	
	// Print each command with its description and usage
	for _, name := range cmdNames {
		cmd := c.commands[name]
		fmt.Printf("  %-10s - %s\n", cmd.Name, cmd.Description)
		fmt.Printf("    Usage: %s\n", cmd.Usage)
	}
	
	return nil
}

// cmdList implements the list command
func (c *Console) cmdList(args []string) error {
	var clients []*client.Client
	
	if len(args) > 0 {
		// Filter by status
		status := client.ClientStatus(args[0])
		clients = c.clientManager.GetClientsByStatus(status)
		fmt.Printf("Clients with status '%s':\n", status)
	} else {
		// List all clients
		clients = c.clientManager.GetAllClients()
		fmt.Println("All clients:")
	}
	
	if len(clients) == 0 {
		fmt.Println("No clients found")
		return nil
	}
	
	// Print client information
	fmt.Printf("%-36s %-15s %-10s %-15s\n", "ID", "IP Address", "Status", "Last Seen")
	fmt.Println(strings.Repeat("-", 80))
	
	for _, client := range clients {
		lastSeen := time.Since(client.LastSeen).Round(time.Second)
		fmt.Printf("%-36s %-15s %-10s %s ago\n", 
			client.ID, 
			client.IPAddress, 
			client.Status, 
			lastSeen)
	}
	
	fmt.Printf("\nTotal: %d clients\n", len(clients))
	return nil
}

// cmdInfo implements the info command
func (c *Console) cmdInfo(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing client ID")
	}
	
	clientID := args[0]
	client, err := c.clientManager.GetClient(clientID)
	if err != nil {
		return err
	}
	
	fmt.Printf("Client Information:\n")
	fmt.Printf("  ID:              %s\n", client.ID)
	fmt.Printf("  Name:            %s\n", client.Name)
	fmt.Printf("  IP Address:      %s\n", client.IPAddress)
	fmt.Printf("  OS:              %s\n", client.OS)
	fmt.Printf("  Architecture:    %s\n", client.Architecture)
	fmt.Printf("  Status:          %s\n", client.Status)
	if client.Status == "error" {
		fmt.Printf("  Error Message:   %s\n", client.ErrorMessage)
	}
	fmt.Printf("  Protocol:        %s\n", client.Protocol)
	fmt.Printf("  Registered At:   %s\n", client.RegisteredAt.Format(time.RFC3339))
	fmt.Printf("  Last Seen:       %s (%s ago)\n", 
		client.LastSeen.Format(time.RFC3339),
		time.Since(client.LastSeen).Round(time.Second))
	fmt.Printf("  Heartbeat:       %s\n", client.HeartbeatInterval)
	
	fmt.Printf("  Supported Modules:\n")
	if len(client.SupportedModules) == 0 {
		fmt.Printf("    None\n")
	} else {
		for _, module := range client.SupportedModules {
			fmt.Printf("    - %s\n", module)
		}
	}
	
	fmt.Printf("  Active Modules:\n")
	if len(client.ActiveModules) == 0 {
		fmt.Printf("    None\n")
	} else {
		for _, module := range client.ActiveModules {
			fmt.Printf("    - %s\n", module)
		}
	}
	
	return nil
}

// cmdStatus implements the status command
func (c *Console) cmdStatus(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: status <client_id> <online|offline|busy|error> [error_message]")
	}
	
	clientID := args[0]
	statusStr := args[1]
	
	var status client.ClientStatus
	switch strings.ToLower(statusStr) {
	case "online":
		status = client.StatusOnline
	case "offline":
		status = client.StatusOffline
	case "busy":
		status = client.StatusBusy
	case "error":
		status = client.StatusError
	default:
		return fmt.Errorf("invalid status: %s", statusStr)
	}
	
	errorMsg := ""
	if status == client.StatusError && len(args) > 2 {
		errorMsg = strings.Join(args[2:], " ")
	}
	
	err := c.clientManager.UpdateClientStatus(clientID, status, errorMsg)
	if err != nil {
		return err
	}
	
	fmt.Printf("Updated client %s status to %s\n", clientID, status)
	return nil
}

// cmdHeartbeat implements the heartbeat command
func (c *Console) cmdHeartbeat(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: heartbeat <check|timeout|random> [args...]")
	}
	
	subcommand := args[0]
	
	switch subcommand {
	case "check":
		// Set check interval
		if len(args) < 2 {
			return fmt.Errorf("usage: heartbeat check <interval_seconds>")
		}
		
		var interval time.Duration
		_, err := fmt.Sscanf(args[1], "%d", &interval)
		if err != nil {
			return fmt.Errorf("invalid interval: %v", err)
		}
		
		c.heartbeatMonitor.SetCheckInterval(interval * time.Second)
		fmt.Printf("Set heartbeat check interval to %s\n", interval*time.Second)
		
	case "timeout":
		// Set timeout
		if len(args) < 2 {
			return fmt.Errorf("usage: heartbeat timeout <timeout_seconds>")
		}
		
		var timeout time.Duration
		_, err := fmt.Sscanf(args[1], "%d", &timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout: %v", err)
		}
		
		c.heartbeatMonitor.SetTimeout(timeout * time.Second)
		fmt.Printf("Set heartbeat timeout to %s\n", timeout*time.Second)
		
	case "random":
		// Enable/disable random intervals
		if len(args) < 2 {
			return fmt.Errorf("usage: heartbeat random <enable|disable> [min_seconds max_seconds]")
		}
		
		switch args[1] {
		case "enable":
			if len(args) < 4 {
				return fmt.Errorf("usage: heartbeat random enable <min_seconds> <max_seconds>")
			}
			
			var min, max time.Duration
			_, err := fmt.Sscanf(args[2], "%d", &min)
			if err != nil {
				return fmt.Errorf("invalid min interval: %v", err)
			}
			
			_, err = fmt.Sscanf(args[3], "%d", &max)
			if err != nil {
				return fmt.Errorf("invalid max interval: %v", err)
			}
			
			c.heartbeatMonitor.EnableRandomIntervals(min*time.Second, max*time.Second)
			fmt.Printf("Enabled random heartbeat intervals (%s - %s)\n", min*time.Second, max*time.Second)
			
		case "disable":
			c.heartbeatMonitor.DisableRandomIntervals()
			fmt.Println("Disabled random heartbeat intervals")
			
		default:
			return fmt.Errorf("usage: heartbeat random <enable|disable> [min_seconds max_seconds]")
		}
		
	default:
		return fmt.Errorf("unknown heartbeat subcommand: %s", subcommand)
	}
	
	return nil
}

// cmdException implements the exception command
func (c *Console) cmdException(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: exception <list|report> [args...]")
	}

	switch args[0] {
	case "list":
		// List exceptions
		if len(args) < 2 {
			// List all exceptions
			exceptions := c.clientManager.GetAllExceptionReports()
			if len(exceptions) == 0 {
				fmt.Println("No exceptions reported")
				return nil
			}

			fmt.Println("All exceptions:")
			fmt.Println("ID                                   Client ID                 Severity   Timestamp           Message")
			fmt.Println("--------------------------------------------------------------------------------")
			for _, exception := range exceptions {
				fmt.Printf("%-36s %-24s %-10s %-20s %s\n",
					exception.ID,
					exception.ClientID,
					exception.Severity,
					exception.Timestamp.Format("2006-01-02 15:04:05"),
					exception.Message,
				)
			}
			fmt.Printf("\nTotal: %d exceptions\n", len(exceptions))
		} else {
			// List exceptions for a specific client
			clientID := args[1]
			exceptions, err := c.clientManager.GetExceptionReports(clientID)
			if err != nil {
				return err
			}

			if len(exceptions) == 0 {
				fmt.Printf("No exceptions reported for client %s\n", clientID)
				return nil
			}

			fmt.Printf("Exceptions for client %s:\n", clientID)
			fmt.Println("ID                                   Severity   Timestamp           Message")
			fmt.Println("--------------------------------------------------------------------------------")
			for _, exception := range exceptions {
				fmt.Printf("%-36s %-10s %-20s %s\n",
					exception.ID,
					exception.Severity,
					exception.Timestamp.Format("2006-01-02 15:04:05"),
					exception.Message,
				)
			}
			fmt.Printf("\nTotal: %d exceptions\n", len(exceptions))
		}

	case "report":
		// Report a new exception
		if len(args) < 4 {
			return fmt.Errorf("usage: exception report <client_id> <info|warning|error|critical> <message> [module] [stack_trace]")
		}

		clientID := args[1]
		severityStr := args[2]
		message := args[3]

		// Validate the severity
		var severity client.ExceptionSeverity
		switch severityStr {
		case "info":
			severity = client.SeverityInfo
		case "warning":
			severity = client.SeverityWarning
		case "error":
			severity = client.SeverityError
		case "critical":
			severity = client.SeverityCritical
		default:
			return fmt.Errorf("invalid severity. Must be one of: info, warning, error, critical")
		}

		// Get optional arguments
		module := ""
		stackTrace := ""
		if len(args) > 4 {
			module = args[4]
		}
		if len(args) > 5 {
			stackTrace = args[5]
		}

		// Report the exception
		report, err := c.clientManager.ReportException(
			clientID,
			message,
			severity,
			module,
			stackTrace,
			nil, // No additional info from console
		)
		if err != nil {
			return err
		}

		fmt.Printf("Exception reported with ID: %s\n", report.ID)

	default:
		return fmt.Errorf("unknown subcommand. Available subcommands: list, report")
	}

	return nil
}

// cmdExit implements the exit command
func (c *Console) cmdExit(args []string) error {
	c.Stop()
	return nil
}
