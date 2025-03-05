package logging

import (
	"testing"
	"time"
)

// TestLogAnalyzer tests the log analyzer functionality
func TestLogAnalyzer(t *testing.T) {
	// Create a log analyzer config
	config := LogAnalyzerConfig{
		MaxEntries:      100,
		TopMessageCount: 5,
	}
	
	// Create a log analyzer
	analyzer := NewLogAnalyzer(config)
	
	// Add some log entries
	for i := 0; i < 10; i++ {
		entry := LogEntry{
			Timestamp: time.Now(),
			Level:     InfoLevel,
			Message:   "Test info message",
			Fields:    nil,
		}
		analyzer.AddEntry(entry)
	}
	
	for i := 0; i < 5; i++ {
		entry := LogEntry{
			Timestamp: time.Now(),
			Level:     ErrorLevel,
			Message:   "Test error message",
			Fields:    nil,
		}
		analyzer.AddEntry(entry)
	}
	
	// Get stats
	stats := analyzer.GetStats()
	
	// Check if stats were collected
	if stats.TotalEntries != 15 {
		t.Errorf("Expected 15 entries, got %d", stats.TotalEntries)
	}
	
	if stats.EntriesByLevel[InfoLevel] != 10 {
		t.Errorf("Expected 10 info entries, got %d", stats.EntriesByLevel[InfoLevel])
	}
	
	if stats.EntriesByLevel[ErrorLevel] != 5 {
		t.Errorf("Expected 5 error entries, got %d", stats.EntriesByLevel[ErrorLevel])
	}
	
	// Check top messages
	if len(stats.TopMessages) == 0 {
		t.Errorf("Expected top messages")
	} else {
		// The most frequent message should be "Test info message"
		if stats.TopMessages[0].Message != "Test info message" || stats.TopMessages[0].Count != 10 {
			t.Errorf("Expected top message to be 'Test info message' with count 10, got '%s' with count %d",
				stats.TopMessages[0].Message, stats.TopMessages[0].Count)
		}
	}
	
	// Get entries by level
	infoEntries := analyzer.GetEntriesByLevel(InfoLevel)
	if len(infoEntries) != 10 {
		t.Errorf("Expected 10 info entries, got %d", len(infoEntries))
	}
	
	errorEntries := analyzer.GetEntriesByLevel(ErrorLevel)
	if len(errorEntries) != 5 {
		t.Errorf("Expected 5 error entries, got %d", len(errorEntries))
	}
	
	// Get all entries
	allEntries := analyzer.GetEntries()
	if len(allEntries) != 15 {
		t.Errorf("Expected 15 entries, got %d", len(allEntries))
	}
	
	// Test max entries limit
	smallConfig := LogAnalyzerConfig{
		MaxEntries:      5,
		TopMessageCount: 2,
	}
	
	smallAnalyzer := NewLogAnalyzer(smallConfig)
	
	// Add more entries than the max
	for i := 0; i < 10; i++ {
		entry := LogEntry{
			Timestamp: time.Now(),
			Level:     InfoLevel,
			Message:   "Test message " + string(rune('A'+i)),
			Fields:    nil,
		}
		smallAnalyzer.AddEntry(entry)
	}
	
	// Check that only the most recent entries are kept
	smallStats := smallAnalyzer.GetStats()
	if smallStats.TotalEntries != 5 {
		t.Errorf("Expected 5 entries (max limit), got %d", smallStats.TotalEntries)
	}
	
	// Check that only the top 2 messages are returned
	if len(smallStats.TopMessages) != 2 {
		t.Errorf("Expected 2 top messages, got %d", len(smallStats.TopMessages))
	}
	
	// Clear entries
	analyzer.ClearEntries()
	
	// Check if entries were cleared
	entries := analyzer.GetEntries()
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", len(entries))
	}
	
	// Check stats after clearing
	statsAfterClear := analyzer.GetStats()
	if statsAfterClear.TotalEntries != 0 {
		t.Errorf("Expected 0 entries in stats after clear, got %d", statsAfterClear.TotalEntries)
	}
}
