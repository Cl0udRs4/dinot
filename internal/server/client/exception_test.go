package client

import (
	"testing"
	"time"
)

// TestExceptionManager tests the exception manager functionality
func TestExceptionManager(t *testing.T) {
	// Create a new exception manager
	manager := NewExceptionManager()

	// Create a test exception report
	report := &ExceptionReport{
		ID:        "test-report-id",
		ClientID:  "test-client-id",
		Timestamp: time.Now(),
		Message:   "Test exception message",
		Severity:  SeverityError,
		Module:    "test-module",
	}

	// Add the report to the manager
	manager.AddReport(report)

	// Test GetReport
	retrievedReport, exists := manager.GetReport("test-report-id")
	if !exists {
		t.Error("Expected report to exist, but it doesn't")
	}
	if retrievedReport.ID != "test-report-id" {
		t.Errorf("Expected report ID test-report-id, got %s", retrievedReport.ID)
	}
	if retrievedReport.Message != "Test exception message" {
		t.Errorf("Expected message 'Test exception message', got %s", retrievedReport.Message)
	}

	// Test GetClientReports
	clientReports := manager.GetClientReports("test-client-id")
	if len(clientReports) != 1 {
		t.Errorf("Expected 1 client report, got %d", len(clientReports))
	}
	if clientReports[0].ID != "test-report-id" {
		t.Errorf("Expected report ID test-report-id, got %s", clientReports[0].ID)
	}

	// Test GetAllReports
	allReports := manager.GetAllReports()
	if len(allReports) != 1 {
		t.Errorf("Expected 1 report, got %d", len(allReports))
	}
	if allReports[0].ID != "test-report-id" {
		t.Errorf("Expected report ID test-report-id, got %s", allReports[0].ID)
	}

	// Test non-existent report
	_, exists = manager.GetReport("non-existent-id")
	if exists {
		t.Error("Expected non-existent report to not exist, but it does")
	}

	// Test non-existent client
	clientReports = manager.GetClientReports("non-existent-client")
	if len(clientReports) != 0 {
		t.Errorf("Expected 0 client reports, got %d", len(clientReports))
	}
}

// TestClientManagerExceptionReporting tests the client manager's exception reporting functionality
func TestClientManagerExceptionReporting(t *testing.T) {
	// Create a new client manager
	manager := NewClientManager()

	// Create and register a test client
	client := NewClient(
		"test-client-id",
		"Test Client",
		"192.168.1.100",
		"Linux",
		"x86_64",
		[]string{"shell", "file", "process"},
		"tcp",
	)
	err := manager.RegisterClient(client)
	if err != nil {
		t.Fatal(err)
	}

	// Report an exception
	report, err := manager.ReportException(
		"test-client-id",
		"Test exception message",
		SeverityError,
		"test-module",
		"test stack trace",
		map[string]string{"key": "value"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if report.ClientID != "test-client-id" {
		t.Errorf("Expected client ID test-client-id, got %s", report.ClientID)
	}
	if report.Message != "Test exception message" {
		t.Errorf("Expected message 'Test exception message', got %s", report.Message)
	}
	if report.Severity != SeverityError {
		t.Errorf("Expected severity error, got %s", report.Severity)
	}
	if report.Module != "test-module" {
		t.Errorf("Expected module test-module, got %s", report.Module)
	}
	if report.StackTrace != "test stack trace" {
		t.Errorf("Expected stack trace 'test stack trace', got %s", report.StackTrace)
	}
	if report.AdditionalInfo["key"] != "value" {
		t.Errorf("Expected additional info key 'value', got %s", report.AdditionalInfo["key"])
	}

	// Check that the client's status was updated
	client, err = manager.GetClient("test-client-id")
	if err != nil {
		t.Fatal(err)
	}
	if client.Status != StatusError {
		t.Errorf("Expected client status error, got %s", client.Status)
	}
	if client.ErrorMessage != "Test exception message" {
		t.Errorf("Expected error message 'Test exception message', got %s", client.ErrorMessage)
	}

	// Get client exception reports
	reports, err := manager.GetExceptionReports("test-client-id")
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 {
		t.Errorf("Expected 1 exception report, got %d", len(reports))
	}
	if reports[0].Message != "Test exception message" {
		t.Errorf("Expected message 'Test exception message', got %s", reports[0].Message)
	}
}
