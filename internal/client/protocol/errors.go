package protocol

import (
	"fmt"
)

// ClientError represents an error in the client protocol
type ClientError struct {
	Type    string
	Message string
	Err     error
}

// Error returns the error message
func (e *ClientError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// NewClientError creates a new client error
func NewClientError(errType, message string, err error) *ClientError {
	return &ClientError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

// Error types for the client protocol
const (
	ErrTypeConnection     = "connection error"
	ErrTypeDisconnection  = "disconnection error"
	ErrTypeSend           = "send error"
	ErrTypeReceive        = "receive error"
	ErrTypeConfiguration  = "configuration error"
	ErrTypeTimeout        = "timeout error"
	ErrTypeProtocolSwitch = "protocol switch error"
)

// IsConnectionError checks if the error is a connection error
func IsConnectionError(err error) bool {
	if clientErr, ok := err.(*ClientError); ok {
		return clientErr.Type == ErrTypeConnection
	}
	return false
}

// IsTimeoutError checks if the error is a timeout error
func IsTimeoutError(err error) bool {
	if clientErr, ok := err.(*ClientError); ok {
		return clientErr.Type == ErrTypeTimeout
	}
	return false
}

// IsProtocolSwitchError checks if the error is a protocol switch error
func IsProtocolSwitchError(err error) bool {
	if clientErr, ok := err.(*ClientError); ok {
		return clientErr.Type == ErrTypeProtocolSwitch
	}
	return false
}
