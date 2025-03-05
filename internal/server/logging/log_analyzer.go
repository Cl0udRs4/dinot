package logging

import (
	"sync"
	"time"
)

// LogEntry represents a log entry for analysis
type LogEntry struct {
	// Timestamp is the time of the log entry
	Timestamp time.Time
	// Level is the log level
	Level LogLevel
	// Message is the log message
	Message string
	// Fields are the log fields
	Fields map[string]interface{}
}

// LogStats represents statistics about logs
type LogStats struct {
	// TotalEntries is the total number of log entries
	TotalEntries int
	// EntriesByLevel is the number of entries by level
	EntriesByLevel map[LogLevel]int
	// TopMessages are the most frequent messages
	TopMessages []struct {
		Message string
		Count   int
	}
	// TimeRange is the time range of the logs
	TimeRange struct {
		Start time.Time
		End   time.Time
	}
}

// LogAnalyzerConfig represents configuration for log analysis
type LogAnalyzerConfig struct {
	// MaxEntries is the maximum number of entries to store
	MaxEntries int
	// TopMessageCount is the number of top messages to track
	TopMessageCount int
}

// LogAnalyzer analyzes logs
type LogAnalyzer struct {
	// config is the analyzer configuration
	config LogAnalyzerConfig
	// entries stores log entries
	entries []LogEntry
	// messageCount tracks message frequency
	messageCount map[string]int
	// mu protects concurrent access to the analyzer
	mu sync.RWMutex
}

// NewLogAnalyzer creates a new log analyzer
func NewLogAnalyzer(config LogAnalyzerConfig) *LogAnalyzer {
	return &LogAnalyzer{
		config:       config,
		entries:      make([]LogEntry, 0, config.MaxEntries),
		messageCount: make(map[string]int),
	}
}

// AddEntry adds a log entry for analysis
func (a *LogAnalyzer) AddEntry(entry LogEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Add entry to the list
	a.entries = append(a.entries, entry)
	
	// Maintain max entries
	if len(a.entries) > a.config.MaxEntries {
		// Remove oldest entry from message count
		oldestMessage := a.entries[0].Message
		a.messageCount[oldestMessage]--
		if a.messageCount[oldestMessage] <= 0 {
			delete(a.messageCount, oldestMessage)
		}
		
		// Remove oldest entry
		a.entries = a.entries[1:]
	}
	
	// Update message count
	a.messageCount[entry.Message]++
}

// GetStats returns statistics about the logs
func (a *LogAnalyzer) GetStats() LogStats {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	stats := LogStats{
		TotalEntries:    len(a.entries),
		EntriesByLevel:  make(map[LogLevel]int),
	}
	
	if len(a.entries) == 0 {
		return stats
	}
	
	// Set time range
	stats.TimeRange.Start = a.entries[0].Timestamp
	stats.TimeRange.End = a.entries[len(a.entries)-1].Timestamp
	
	// Count entries by level
	for _, entry := range a.entries {
		stats.EntriesByLevel[entry.Level]++
	}
	
	// Get top messages
	type messageFreq struct {
		message string
		count   int
	}
	
	topMessages := make([]messageFreq, 0, len(a.messageCount))
	for message, count := range a.messageCount {
		topMessages = append(topMessages, messageFreq{message, count})
	}
	
	// Sort by count (simple bubble sort for small lists)
	for i := 0; i < len(topMessages)-1; i++ {
		for j := 0; j < len(topMessages)-i-1; j++ {
			if topMessages[j].count < topMessages[j+1].count {
				topMessages[j], topMessages[j+1] = topMessages[j+1], topMessages[j]
			}
		}
	}
	
	// Get top N messages
	count := a.config.TopMessageCount
	if count > len(topMessages) {
		count = len(topMessages)
	}
	
	stats.TopMessages = make([]struct {
		Message string
		Count   int
	}, count)
	
	for i := 0; i < count; i++ {
		stats.TopMessages[i].Message = topMessages[i].message
		stats.TopMessages[i].Count = topMessages[i].count
	}
	
	return stats
}

// GetEntries returns all log entries
func (a *LogAnalyzer) GetEntries() []LogEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	entries := make([]LogEntry, len(a.entries))
	copy(entries, a.entries)
	return entries
}

// GetEntriesByLevel returns log entries filtered by level
func (a *LogAnalyzer) GetEntriesByLevel(level LogLevel) []LogEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	var filtered []LogEntry
	for _, entry := range a.entries {
		if entry.Level == level {
			filtered = append(filtered, entry)
		}
	}
	
	return filtered
}

// ClearEntries clears all log entries
func (a *LogAnalyzer) ClearEntries() {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.entries = make([]LogEntry, 0, a.config.MaxEntries)
	a.messageCount = make(map[string]int)
}
