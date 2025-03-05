package cli

import (
	"bufio"
	"io"
	"os"
	"testing"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/client"
)

// mockReader is a mock for bufio.Reader that returns predefined inputs
type mockReader struct {
	*bufio.Reader
	inputs []string
	index  int
}

func newMockReader(inputs []string) *bufio.Reader {
	// Create a pipe to simulate input
	r, w := io.Pipe()
	
	// Write the inputs to the pipe in a goroutine
	go func() {
		for _, input := range inputs {
			w.Write([]byte(input + "\n"))
		}
		w.Close()
	}()
	
	return bufio.NewReader(r)
}

// setupTestConsole creates a test console with mock data
func setupTestConsole() (*Console, *client.ClientManager, *client.HeartbeatMonitor) {
	clientManager := client.NewClientManager()
	heartbeatMonitor := client.NewHeartbeatMonitor(clientManager, 30*time.Second, 60*time.Second)
	
	// Create a test client
	testClient := client.NewClient(
		"test-client-id",
		"Test Client",
		"192.168.1.100",
		"Linux",
		"x86_64",
		[]string{"shell", "file", "process"},
		"tcp",
	)
	
	// Register the test client
	_ = clientManager.RegisterClient(testClient)
	
	console := NewConsole(clientManager, heartbeatMonitor)
	
	return console, clientManager, heartbeatMonitor
}

// TestConsoleCommands tests the console commands
func TestConsoleCommands(t *testing.T) {
	console, _, _ := setupTestConsole()
	
	// Test help command
	err := console.cmdHelp([]string{})
	if err != nil {
		t.Errorf("Help command failed: %v", err)
	}
	
	// Test list command
	err = console.cmdList([]string{})
	if err != nil {
		t.Errorf("List command failed: %v", err)
	}
	
	// Test list with status filter
	err = console.cmdList([]string{"online"})
	if err != nil {
		t.Errorf("List command with status filter failed: %v", err)
	}
	
	// Test info command
	err = console.cmdInfo([]string{"test-client-id"})
	if err != nil {
		t.Errorf("Info command failed: %v", err)
	}
	
	// Test info command with invalid client ID
	err = console.cmdInfo([]string{"invalid-id"})
	if err == nil {
		t.Errorf("Info command with invalid client ID should fail")
	}
	
	// Test status command
	err = console.cmdStatus([]string{"test-client-id", "busy"})
	if err != nil {
		t.Errorf("Status command failed: %v", err)
	}
	
	// Test status command with error status and message
	err = console.cmdStatus([]string{"test-client-id", "error", "Test", "error", "message"})
	if err != nil {
		t.Errorf("Status command with error message failed: %v", err)
	}
	
	// Test heartbeat check command
	err = console.cmdHeartbeat([]string{"check", "45"})
	if err != nil {
		t.Errorf("Heartbeat check command failed: %v", err)
	}
	
	// Test heartbeat timeout command
	err = console.cmdHeartbeat([]string{"timeout", "90"})
	if err != nil {
		t.Errorf("Heartbeat timeout command failed: %v", err)
	}
	
	// Test heartbeat random enable command
	err = console.cmdHeartbeat([]string{"random", "enable", "5", "300"})
	if err != nil {
		t.Errorf("Heartbeat random enable command failed: %v", err)
	}
	
	// Test heartbeat random disable command
	err = console.cmdHeartbeat([]string{"random", "disable"})
	if err != nil {
		t.Errorf("Heartbeat random disable command failed: %v", err)
	}
	
	// Test exit command
	err = console.cmdExit([]string{})
	if err != nil {
		t.Errorf("Exit command failed: %v", err)
	}
	
	if console.running {
		t.Errorf("Console should not be running after exit command")
	}
}

// TestConsoleStartStop tests the console start and stop functions
func TestConsoleStartStop(t *testing.T) {
	console, _, _ := setupTestConsole()
	
	// Replace stdin with a pipe
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
	}()
	
	// Replace stdout with a pipe
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()
	
	// Set up mock reader with exit command
	console.reader = newMockReader([]string{"exit"})
	
	// Start the console in a goroutine
	go console.Start()
	
	// Wait for the console to process the exit command
	time.Sleep(100 * time.Millisecond)
	
	// Check that the console is not running
	if console.running {
		t.Errorf("Console should not be running after exit command")
	}
	
	// Close the pipe
	w.Close()
	r.Close()
}

// TestConsoleInvalidCommands tests handling of invalid commands
func TestConsoleInvalidCommands(t *testing.T) {
	console, _, _ := setupTestConsole()
	
	// Test list command with invalid status
	err := console.cmdList([]string{"invalid-status"})
	if err != nil {
		t.Errorf("List command with invalid status should not fail: %v", err)
	}
	
	// Test status command with missing arguments
	err = console.cmdStatus([]string{})
	if err == nil {
		t.Errorf("Status command with missing arguments should fail")
	}
	
	// Test status command with invalid status
	err = console.cmdStatus([]string{"test-client-id", "invalid-status"})
	if err == nil {
		t.Errorf("Status command with invalid status should fail")
	}
	
	// Test heartbeat command with missing subcommand
	err = console.cmdHeartbeat([]string{})
	if err == nil {
		t.Errorf("Heartbeat command with missing subcommand should fail")
	}
	
	// Test heartbeat check command with missing interval
	err = console.cmdHeartbeat([]string{"check"})
	if err == nil {
		t.Errorf("Heartbeat check command with missing interval should fail")
	}
	
	// Test heartbeat timeout command with missing timeout
	err = console.cmdHeartbeat([]string{"timeout"})
	if err == nil {
		t.Errorf("Heartbeat timeout command with missing timeout should fail")
	}
	
	// Test heartbeat random command with missing enable/disable
	err = console.cmdHeartbeat([]string{"random"})
	if err == nil {
		t.Errorf("Heartbeat random command with missing enable/disable should fail")
	}
	
	// Test heartbeat random enable command with missing min/max
	err = console.cmdHeartbeat([]string{"random", "enable"})
	if err == nil {
		t.Errorf("Heartbeat random enable command with missing min/max should fail")
	}
	
	// Test heartbeat command with invalid subcommand
	err = console.cmdHeartbeat([]string{"invalid-subcommand"})
	if err == nil {
		t.Errorf("Heartbeat command with invalid subcommand should fail")
	}
}
