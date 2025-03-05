package logging

import (
	"testing"
	"time"
)

// TestResourceMonitor tests the resource monitoring functionality
func TestResourceMonitor(t *testing.T) {
	// Create a logger
	logger := NewLogrusLogger()
	
	// Create a resource monitor config
	config := ResourceMonitorConfig{
		Interval:      100 * time.Millisecond,
		EnableCPU:     true,
		EnableMemory:  true,
		EnableNetwork: true,
	}
	
	// Create a resource monitor
	monitor := NewResourceMonitor(logger, config)
	
	// Start the monitor
	monitor.Start()
	
	// Wait for some stats to be collected
	time.Sleep(300 * time.Millisecond)
	
	// Get the latest stats
	stats := monitor.GetLatestStats()
	
	// Check if stats were collected
	if stats.Timestamp.IsZero() {
		t.Errorf("Expected stats to be collected")
	}
	
	// Get stats history
	history := monitor.GetStatsHistory()
	if len(history) == 0 {
		t.Errorf("Expected stats history to be collected")
	}
	
	// Stop the monitor
	monitor.Stop()
	
	// Test with different config
	configMinimal := ResourceMonitorConfig{
		Interval:      50 * time.Millisecond,
		EnableCPU:     true,
		EnableMemory:  false,
		EnableNetwork: false,
	}
	
	monitorMinimal := NewResourceMonitor(logger, configMinimal)
	monitorMinimal.Start()
	
	// Wait for stats to be collected - increased wait time to ensure CPU stats are collected
	time.Sleep(300 * time.Millisecond)
	
	// Get latest stats
	statsMinimal := monitorMinimal.GetLatestStats()
	
	// Log the CPU stats for debugging
	t.Logf("CPU stats: %f", statsMinimal.CPUUsage)
	
	// Verify only CPU stats were collected
	if statsMinimal.CPUUsage == 0 {
		t.Errorf("Expected CPU stats to be collected")
	}
	
	if statsMinimal.MemoryUsage != 0 || statsMinimal.MemoryTotal != 0 {
		t.Errorf("Expected memory stats to be zero when disabled")
	}
	
	if statsMinimal.NetworkBytesSent != 0 || statsMinimal.NetworkBytesReceived != 0 {
		t.Errorf("Expected network stats to be zero when disabled")
	}
	
	// Stop the monitor
	monitorMinimal.Stop()
}
