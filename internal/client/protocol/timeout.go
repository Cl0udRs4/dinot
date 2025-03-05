package protocol

import (
	"context"
	"sync"
	"time"
)

// TimeoutDetector is responsible for detecting timeouts in protocol communications
// and triggering protocol switching when necessary
type TimeoutDetector struct {
	// protocols is a map of protocol name to Protocol
	protocols map[string]Protocol
	
	// activeProtocol is the currently active protocol
	activeProtocol Protocol
	
	// timeoutThreshold is the number of consecutive timeouts before switching protocols
	timeoutThreshold int
	
	// timeoutCount tracks consecutive timeouts for the active protocol
	timeoutCount int
	
	// checkInterval is the interval between timeout checks
	checkInterval time.Duration
	
	// lastActivity is the time of the last successful communication
	lastActivity time.Time
	
	// maxInactivity is the maximum allowed inactivity period
	maxInactivity time.Duration
	
	// mu protects concurrent access to the detector
	mu sync.RWMutex
	
	// ctx is the context for the detector
	ctx context.Context
	
	// cancel is the function to cancel the detector context
	cancel context.CancelFunc
	
	// onTimeout is called when a timeout is detected
	onTimeout func(protocol Protocol)
	
	// onSwitch is called when a protocol switch is triggered
	onSwitch func(oldProtocol, newProtocol Protocol)
	
	// running indicates if the detector is running
	running bool
}

// TimeoutDetectorConfig defines the configuration for the timeout detector
type TimeoutDetectorConfig struct {
	// TimeoutThreshold is the number of consecutive timeouts before switching protocols
	TimeoutThreshold int
	
	// CheckInterval is the interval between timeout checks in seconds
	CheckInterval int
	
	// MaxInactivity is the maximum allowed inactivity period in seconds
	MaxInactivity int
}

// NewTimeoutDetector creates a new timeout detector
func NewTimeoutDetector(config TimeoutDetectorConfig) *TimeoutDetector {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Set default values if not provided
	if config.TimeoutThreshold <= 0 {
		config.TimeoutThreshold = 3
	}
	
	if config.CheckInterval <= 0 {
		config.CheckInterval = 5
	}
	
	if config.MaxInactivity <= 0 {
		config.MaxInactivity = 60
	}
	
	return &TimeoutDetector{
		protocols:        make(map[string]Protocol),
		timeoutThreshold: config.TimeoutThreshold,
		timeoutCount:     0,
		checkInterval:    time.Duration(config.CheckInterval) * time.Second,
		lastActivity:     time.Now(),
		maxInactivity:    time.Duration(config.MaxInactivity) * time.Second,
		ctx:              ctx,
		cancel:           cancel,
		running:          false,
	}
}

// RegisterProtocol registers a protocol with the detector
func (d *TimeoutDetector) RegisterProtocol(protocol Protocol) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	name := protocol.GetName()
	d.protocols[name] = protocol
	
	// If no active protocol is set, make this the active protocol
	if d.activeProtocol == nil {
		d.activeProtocol = protocol
	}
}

// SetActiveProtocol sets the active protocol
func (d *TimeoutDetector) SetActiveProtocol(protocol Protocol) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.activeProtocol = protocol
	d.timeoutCount = 0
	d.lastActivity = time.Now()
}

// SetOnTimeoutHandler sets the handler to be called when a timeout is detected
func (d *TimeoutDetector) SetOnTimeoutHandler(handler func(protocol Protocol)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.onTimeout = handler
}

// SetOnSwitchHandler sets the handler to be called when a protocol switch is triggered
func (d *TimeoutDetector) SetOnSwitchHandler(handler func(oldProtocol, newProtocol Protocol)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.onSwitch = handler
}

// Start starts the timeout detector
func (d *TimeoutDetector) Start() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if d.running {
		return
	}
	
	d.running = true
	d.lastActivity = time.Now()
	
	// Start the timeout detection loop in a goroutine
	go d.detectionLoop()
}

// Stop stops the timeout detector
func (d *TimeoutDetector) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if !d.running {
		return
	}
	
	d.running = false
	d.cancel()
}

// RecordActivity records a successful activity, resetting the timeout count
func (d *TimeoutDetector) RecordActivity() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.timeoutCount = 0
	d.lastActivity = time.Now()
}

// RecordTimeout records a timeout, incrementing the timeout count
func (d *TimeoutDetector) RecordTimeout() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.timeoutCount++
	
	// Call the timeout handler if set
	if d.onTimeout != nil && d.activeProtocol != nil {
		d.onTimeout(d.activeProtocol)
	}
	
	// Check if we need to switch protocols
	if d.timeoutCount >= d.timeoutThreshold {
		d.triggerProtocolSwitch()
	}
}

// IsRunning returns true if the detector is running
func (d *TimeoutDetector) IsRunning() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	return d.running
}

// GetTimeoutCount returns the current timeout count
func (d *TimeoutDetector) GetTimeoutCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	return d.timeoutCount
}

// GetLastActivity returns the time of the last successful activity
func (d *TimeoutDetector) GetLastActivity() time.Time {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	return d.lastActivity
}

// detectionLoop is the main loop for timeout detection
func (d *TimeoutDetector) detectionLoop() {
	ticker := time.NewTicker(d.checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.checkInactivity()
		}
	}
}

// checkInactivity checks if the active protocol has been inactive for too long
func (d *TimeoutDetector) checkInactivity() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if d.activeProtocol == nil {
		return
	}
	
	// Check if the protocol has been inactive for too long
	if time.Since(d.lastActivity) > d.maxInactivity {
		// Increment the timeout count
		d.timeoutCount++
		
		// Call the timeout handler if set
		if d.onTimeout != nil {
			d.onTimeout(d.activeProtocol)
		}
		
		// Check if we need to switch protocols
		if d.timeoutCount >= d.timeoutThreshold {
			d.triggerProtocolSwitch()
		}
	}
}

// triggerProtocolSwitch triggers a protocol switch
func (d *TimeoutDetector) triggerProtocolSwitch() {
	if d.activeProtocol == nil || len(d.protocols) <= 1 {
		return
	}
	
	// Find the next protocol to switch to
	var nextProtocol Protocol
	for _, protocol := range d.protocols {
		if protocol != d.activeProtocol {
			nextProtocol = protocol
			break
		}
	}
	
	if nextProtocol == nil {
		return
	}
	
	// Store the old protocol for the handler
	oldProtocol := d.activeProtocol
	
	// Switch to the new protocol
	d.activeProtocol = nextProtocol
	d.timeoutCount = 0
	
	// Call the switch handler if set
	if d.onSwitch != nil {
		d.onSwitch(oldProtocol, nextProtocol)
	}
}

// SetTimeoutThreshold sets the timeout threshold
func (d *TimeoutDetector) SetTimeoutThreshold(threshold int) {
	if threshold <= 0 {
		return
	}
	
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.timeoutThreshold = threshold
}

// SetCheckInterval sets the check interval
func (d *TimeoutDetector) SetCheckInterval(interval int) {
	if interval <= 0 {
		return
	}
	
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.checkInterval = time.Duration(interval) * time.Second
}

// SetMaxInactivity sets the maximum allowed inactivity period
func (d *TimeoutDetector) SetMaxInactivity(maxInactivity int) {
	if maxInactivity <= 0 {
		return
	}
	
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.maxInactivity = time.Duration(maxInactivity) * time.Second
}
