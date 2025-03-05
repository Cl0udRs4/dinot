package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFileLogging tests the file logging functionality
func TestFileLogging(t *testing.T) {
	// Create a temporary directory for logs
	tempDir, err := os.MkdirTemp("", "log-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create logger
	logger := NewLogrusLogger()
	
	// Enable file logging
	config := FileLogConfig{
		Directory:  tempDir,
		MaxSize:    1,
		MaxAge:     1,
		MaxBackups: 1,
		Compress:   false,
	}
	
	err = logger.EnableFileLogging(config)
	if err != nil {
		t.Fatalf("Failed to enable file logging: %v", err)
	}
	
	// Log some messages
	logger.Info("Test info message", nil)
	logger.Error("Test error message", nil)
	
	// Verify log file exists
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}
	
	if len(files) == 0 {
		t.Errorf("No log files found in temp directory")
	}
	
	// Verify log file contains the messages
	logFilePath := filepath.Join(tempDir, "dinot.log")
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if !strings.Contains(string(content), "Test info message") ||
	   !strings.Contains(string(content), "Test error message") {
		t.Errorf("Log file does not contain expected messages")
	}
	
	// Test WithField with file logging
	fieldLogger := logger.WithField("test_key", "test_value")
	fieldLogger.Info("Test field message", nil)
	
	// Read updated content
	content, err = os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	// Just check if the message is in the log file, as the field format might vary
	if !strings.Contains(string(content), "Test field message") {
		t.Errorf("Log file does not contain field message")
	}
	
	// Disable file logging
	logger.DisableFileLogging()
	
	// Log another message after disabling file logging
	logger.Info("After disable message", nil)
	
	// Read content again
	afterContent, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	// Content should be the same as before (no new message)
	if len(afterContent) != len(content) {
		t.Errorf("Log file was modified after disabling file logging")
	}
}
