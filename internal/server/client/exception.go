package client

import (
	"sync"
	"time"
)

// ExceptionSeverity represents the severity level of an exception
type ExceptionSeverity string

const (
	// SeverityInfo represents an informational exception
	SeverityInfo ExceptionSeverity = "info"
	// SeverityWarning represents a warning exception
	SeverityWarning ExceptionSeverity = "warning"
	// SeverityError represents an error exception
	SeverityError ExceptionSeverity = "error"
	// SeverityCritical represents a critical exception
	SeverityCritical ExceptionSeverity = "critical"
)

// ExceptionReport represents a detailed exception report from a client
type ExceptionReport struct {
	// ID is the unique identifier for this exception report
	ID string `json:"id"`
	
	// ClientID is the ID of the client that reported the exception
	ClientID string `json:"client_id"`
	
	// Timestamp is when the exception was reported
	Timestamp time.Time `json:"timestamp"`
	
	// Message is the exception message
	Message string `json:"message"`
	
	// Severity is the severity level of the exception
	Severity ExceptionSeverity `json:"severity"`
	
	// Module is the module where the exception occurred (optional)
	Module string `json:"module,omitempty"`
	
	// StackTrace is the stack trace of the exception (optional)
	StackTrace string `json:"stack_trace,omitempty"`
	
	// AdditionalInfo contains any additional information about the exception
	AdditionalInfo map[string]string `json:"additional_info,omitempty"`
}

// ExceptionManager manages exception reports
type ExceptionManager struct {
	// reports is a map of exception report IDs to ExceptionReport objects
	reports map[string]*ExceptionReport
	
	// clientReports is a map of client IDs to lists of exception report IDs
	clientReports map[string][]string
	
	// mu protects concurrent access to the reports maps
	mu sync.RWMutex
}

// NewExceptionManager creates a new exception manager
func NewExceptionManager() *ExceptionManager {
	return &ExceptionManager{
		reports:       make(map[string]*ExceptionReport),
		clientReports: make(map[string][]string),
	}
}

// AddReport adds a new exception report
func (m *ExceptionManager) AddReport(report *ExceptionReport) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Add the report to the reports map
	m.reports[report.ID] = report
	
	// Add the report ID to the client's list of reports
	m.clientReports[report.ClientID] = append(m.clientReports[report.ClientID], report.ID)
}

// GetReport retrieves an exception report by ID
func (m *ExceptionManager) GetReport(reportID string) (*ExceptionReport, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	report, exists := m.reports[reportID]
	return report, exists
}

// GetClientReports retrieves all exception reports for a client
func (m *ExceptionManager) GetClientReports(clientID string) []*ExceptionReport {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	reportIDs, exists := m.clientReports[clientID]
	if !exists {
		return []*ExceptionReport{}
	}
	
	reports := make([]*ExceptionReport, 0, len(reportIDs))
	for _, id := range reportIDs {
		if report, exists := m.reports[id]; exists {
			reports = append(reports, report)
		}
	}
	
	return reports
}

// GetAllReports retrieves all exception reports
func (m *ExceptionManager) GetAllReports() []*ExceptionReport {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	reports := make([]*ExceptionReport, 0, len(m.reports))
	for _, report := range m.reports {
		reports = append(reports, report)
	}
	
	return reports
}
