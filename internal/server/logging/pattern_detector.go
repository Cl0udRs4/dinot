package logging

import (
	"sync"
	"time"

	"github.com/Cl0udRs4/dinot/internal/server/client"
)

// ExceptionPattern represents a pattern of exceptions
type ExceptionPattern struct {
	// PatternID is the unique identifier for the pattern
	PatternID string
	// ClientID is the client ID associated with the pattern
	ClientID string
	// Module is the module associated with the pattern
	Module string
	// MessagePattern is the common message pattern
	MessagePattern string
	// Severity is the severity of the pattern
	Severity client.ExceptionSeverity
	// Frequency is the number of occurrences
	Frequency int
	// FirstSeen is when the pattern was first seen
	FirstSeen time.Time
	// LastSeen is when the pattern was last seen
	LastSeen time.Time
	// TimeWindow is the time window for the pattern in seconds
	TimeWindow int64
}

// PatternDetectorConfig represents configuration for pattern detection
type PatternDetectorConfig struct {
	// TimeWindow is the time window for pattern detection in seconds
	TimeWindow int64
	// MinFrequency is the minimum frequency to consider a pattern
	MinFrequency int
	// SimilarityThreshold is the threshold for message similarity (0-1)
	SimilarityThreshold float64
}

// PatternDetector detects patterns in exceptions
type PatternDetector struct {
	// logger is the logger instance
	logger Logger
	// clientManager is the client manager instance
	clientManager *client.ClientManager
	// config is the detector configuration
	config PatternDetectorConfig
	// patterns stores detected patterns
	patterns map[string]*ExceptionPattern
	// mu protects concurrent access to the detector
	mu sync.RWMutex
}

// NewPatternDetector creates a new pattern detector
func NewPatternDetector(logger Logger, clientManager *client.ClientManager, config PatternDetectorConfig) *PatternDetector {
	return &PatternDetector{
		logger:        logger,
		clientManager: clientManager,
		config:        config,
		patterns:      make(map[string]*ExceptionPattern),
	}
}

// DetectPatterns detects patterns in exceptions
func (d *PatternDetector) DetectPatterns() []*ExceptionPattern {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	// Get all exception reports
	exceptions := d.clientManager.GetAllExceptionReports()
	
	// Group exceptions by client and module
	groupedExceptions := make(map[string][]*client.ExceptionReport)
	for _, exception := range exceptions {
		key := exception.ClientID + ":" + exception.Module
		groupedExceptions[key] = append(groupedExceptions[key], exception)
	}
	
	// Clear old patterns
	d.patterns = make(map[string]*ExceptionPattern)
	
	// Detect patterns in each group
	var detectedPatterns []*ExceptionPattern
	for key, groupExceptions := range groupedExceptions {
		patterns := d.detectPatternsInGroup(key, groupExceptions)
		for _, pattern := range patterns {
			d.patterns[pattern.PatternID] = pattern
			detectedPatterns = append(detectedPatterns, pattern)
		}
	}
	
	return detectedPatterns
}

// detectPatternsInGroup detects patterns in a group of exceptions
func (d *PatternDetector) detectPatternsInGroup(groupKey string, exceptions []*client.ExceptionReport) []*ExceptionPattern {
	if len(exceptions) == 0 {
		return nil
	}
	
	// Simple implementation: group by message and count frequency
	messageGroups := make(map[string][]*client.ExceptionReport)
	for _, exception := range exceptions {
		messageGroups[exception.Message] = append(messageGroups[exception.Message], exception)
	}
	
	var patterns []*ExceptionPattern
	for message, group := range messageGroups {
		if len(group) < d.config.MinFrequency {
			continue
		}
		
		// Check time window
		firstSeen := group[0].Timestamp
		lastSeen := group[len(group)-1].Timestamp
		timeWindow := lastSeen.Sub(firstSeen).Seconds()
		
		if timeWindow > float64(d.config.TimeWindow) {
			continue
		}
		
		// Create pattern
		clientID := group[0].ClientID
		module := group[0].Module
		severity := group[0].Severity
		
		pattern := &ExceptionPattern{
			PatternID:      groupKey + ":" + message,
			ClientID:       clientID,
			Module:         module,
			MessagePattern: message,
			Severity:       severity,
			Frequency:      len(group),
			FirstSeen:      firstSeen,
			LastSeen:       lastSeen,
			TimeWindow:     int64(timeWindow),
		}
		
		patterns = append(patterns, pattern)
	}
	
	return patterns
}

// GetDetectedPatterns returns all detected patterns
func (d *PatternDetector) GetDetectedPatterns() []*ExceptionPattern {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	patterns := make([]*ExceptionPattern, 0, len(d.patterns))
	for _, pattern := range d.patterns {
		patterns = append(patterns, pattern)
	}
	
	return patterns
}

// GetPatternsByClient returns patterns for a specific client
func (d *PatternDetector) GetPatternsByClient(clientID string) []*ExceptionPattern {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	var clientPatterns []*ExceptionPattern
	for _, pattern := range d.patterns {
		if pattern.ClientID == clientID {
			clientPatterns = append(clientPatterns, pattern)
		}
	}
	
	return clientPatterns
}
