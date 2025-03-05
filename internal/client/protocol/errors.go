package protocol

import "errors"

var (
    // ErrNotConnected is returned when an operation is attempted on a disconnected protocol
    ErrNotConnected = errors.New("protocol not connected")
    
    // ErrTimeout is returned when an operation times out
    ErrTimeout = errors.New("operation timed out")
    
    // ErrProtocolSwitch is returned when a protocol switch is needed
    ErrProtocolSwitch = errors.New("protocol switch needed")
    
    // ErrInvalidMessageType is returned when an invalid message type is received
    ErrInvalidMessageType = errors.New("invalid message type")
    
    // ErrNoProtocolAvailable is returned when no protocol is available
    ErrNoProtocolAvailable = errors.New("no protocol available")
    
    // ErrSendFailed is returned when sending data fails
    ErrSendFailed = errors.New("failed to send data")
)
