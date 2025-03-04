// Package common provides shared utilities and types for the server
package common

import (
	"fmt"
)

// Error types for the server
const (
	ErrListenerAlreadyRunning    = "listener already running"
	ErrListenerNotRunning        = "listener not running"
	ErrListenerAlreadyRegistered = "listener already registered"
	ErrListenerNotRegistered     = "listener not registered"
	ErrListenerStartFailed       = "listener start failed"
	ErrListenerStopFailed        = "listener stop failed"
	ErrInvalidConfig             = "invalid configuration"
	ErrNotImplemented            = "not implemented"
)

// ServerError represents an error in the server
type ServerError struct {
	Type    string
	Message string
	Err     error
}

// Error returns the error message
func (e *ServerError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// NewServerError creates a new server error
func NewServerError(errType, message string, err error) *ServerError {
	return &ServerError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}
