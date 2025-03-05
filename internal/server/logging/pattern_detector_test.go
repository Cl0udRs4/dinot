package logging

import (
	"testing"

	"github.com/Cl0udRs4/dinot/internal/server/client"
)

// TestPatternDetector tests the pattern detection functionality
func TestPatternDetector(t *testing.T) {
	// Create a client manager
	clientManager := client.NewClientManager()
	
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
	
	// Register the client
	err := clientManager.RegisterClient(testClient)
	if err != nil {
		t.Fatalf("Failed to register client: %v", err)
	}
	
	// Create a logger
	logger := NewLogrusLogger()
	
	// Create a pattern detector config
	config := PatternDetectorConfig{
		TimeWindow:         60,
		MinFrequency:       2,
		SimilarityThreshold: 0.8,
	}
	
	// Create a pattern detector
	detector := NewPatternDetector(logger, clientManager, config)
	
	// Report some exceptions with the same message
	for i := 0; i < 3; i++ {
		_, err = clientManager.ReportException(
			"test-client-id",
			"Connection timeout",
			client.SeverityError,
			"network",
			"test stack trace",
			nil,
		)
		if err != nil {
			t.Fatalf("Failed to report exception: %v", err)
		}
	}
	
	// Report some exceptions with a different message
	for i := 0; i < 2; i++ {
		_, err = clientManager.ReportException(
			"test-client-id",
			"Authentication failed",
			client.SeverityWarning,
			"auth",
			"test stack trace",
			nil,
		)
		if err != nil {
			t.Fatalf("Failed to report exception: %v", err)
		}
	}
	
	// Detect patterns
	patterns := detector.DetectPatterns()
	
	// Verify patterns were detected
	if len(patterns) == 0 {
		t.Errorf("Expected patterns to be detected")
	}
	
	// Verify pattern details
	foundConnectionPattern := false
	foundAuthPattern := false
	
	for _, pattern := range patterns {
		if pattern.ClientID == "test-client-id" && 
		   pattern.Module == "network" && 
		   pattern.MessagePattern == "Connection timeout" && 
		   pattern.Frequency == 3 {
			foundConnectionPattern = true
		}
		
		if pattern.ClientID == "test-client-id" && 
		   pattern.Module == "auth" && 
		   pattern.MessagePattern == "Authentication failed" && 
		   pattern.Frequency == 2 {
			foundAuthPattern = true
		}
	}
	
	if !foundConnectionPattern {
		t.Errorf("Expected to find connection timeout pattern")
	}
	
	if !foundAuthPattern {
		t.Errorf("Expected to find authentication failed pattern")
	}
	
	// Get patterns by client
	clientPatterns := detector.GetPatternsByClient("test-client-id")
	if len(clientPatterns) == 0 {
		t.Errorf("Expected patterns for client")
	}
	
	// Get patterns for non-existent client
	nonExistentPatterns := detector.GetPatternsByClient("non-existent-client")
	if len(nonExistentPatterns) != 0 {
		t.Errorf("Expected no patterns for non-existent client")
	}
	
	// Get all detected patterns
	allPatterns := detector.GetDetectedPatterns()
	if len(allPatterns) != len(patterns) {
		t.Errorf("Expected same number of patterns from GetDetectedPatterns")
	}
}
