package logging

import (
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

// LogLevel represents the severity level of a log message
type LogLevel string

const (
	// DebugLevel is used for development and debugging information
	DebugLevel LogLevel = "debug"
	// InfoLevel is used for general operational information
	InfoLevel LogLevel = "info"
	// WarnLevel is used for warning messages
	WarnLevel LogLevel = "warn"
	// ErrorLevel is used for error messages
	ErrorLevel LogLevel = "error"
	// FatalLevel is used for critical errors that lead to termination
	FatalLevel LogLevel = "fatal"
)

// Logger defines the interface for logging operations
type Logger interface {
	// Debug logs a message at the debug level
	Debug(msg string, fields map[string]interface{})
	// Info logs a message at the info level
	Info(msg string, fields map[string]interface{})
	// Warn logs a message at the warn level
	Warn(msg string, fields map[string]interface{})
	// Error logs a message at the error level
	Error(msg string, fields map[string]interface{})
	// Fatal logs a message at the fatal level and then exits
	Fatal(msg string, fields map[string]interface{})
	// WithField adds a field to the logger
	WithField(key string, value interface{}) Logger
	// WithFields adds multiple fields to the logger
	WithFields(fields map[string]interface{}) Logger
	// SetOutput sets the output destination for the logger
	SetOutput(out io.Writer)
	// SetLevel sets the minimum log level
	SetLevel(level LogLevel)
	// GetLevel returns the current log level
	GetLevel() LogLevel
}

// LogrusLogger implements the Logger interface using logrus
type LogrusLogger struct {
	logger *logrus.Logger
	mu     sync.RWMutex
}

// NewLogrusLogger creates a new LogrusLogger
func NewLogrusLogger() *LogrusLogger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	return &LogrusLogger{
		logger: logger,
	}
}

// Debug logs a message at the debug level
func (l *LogrusLogger) Debug(msg string, fields map[string]interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if fields == nil {
		l.logger.Debug(msg)
	} else {
		l.logger.WithFields(logrus.Fields(fields)).Debug(msg)
	}
}

// Info logs a message at the info level
func (l *LogrusLogger) Info(msg string, fields map[string]interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if fields == nil {
		l.logger.Info(msg)
	} else {
		l.logger.WithFields(logrus.Fields(fields)).Info(msg)
	}
}

// Warn logs a message at the warn level
func (l *LogrusLogger) Warn(msg string, fields map[string]interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if fields == nil {
		l.logger.Warn(msg)
	} else {
		l.logger.WithFields(logrus.Fields(fields)).Warn(msg)
	}
}

// Error logs a message at the error level
func (l *LogrusLogger) Error(msg string, fields map[string]interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if fields == nil {
		l.logger.Error(msg)
	} else {
		l.logger.WithFields(logrus.Fields(fields)).Error(msg)
	}
}

// Fatal logs a message at the fatal level and then exits
func (l *LogrusLogger) Fatal(msg string, fields map[string]interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if fields == nil {
		l.logger.Fatal(msg)
	} else {
		l.logger.WithFields(logrus.Fields(fields)).Fatal(msg)
	}
}

// WithField adds a field to the logger
func (l *LogrusLogger) WithField(key string, value interface{}) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	// Create a new logger with the field
	newLogrusLogger := logrus.New()
	newLogrusLogger.SetFormatter(l.logger.Formatter)
	newLogrusLogger.SetOutput(l.logger.Out)
	newLogrusLogger.SetLevel(l.logger.Level)
	
	// Create a new LogrusLogger with the field
	newLogger := &LogrusLogger{
		logger: newLogrusLogger,
	}
	
	// Add the field to every log entry
	entry := newLogger.logger.WithField(key, value)
	newLogger.logger = entry.Logger
	
	return newLogger
}

// WithFields adds multiple fields to the logger
func (l *LogrusLogger) WithFields(fields map[string]interface{}) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	// Create a new logger with the fields
	newLogrusLogger := logrus.New()
	newLogrusLogger.SetFormatter(l.logger.Formatter)
	newLogrusLogger.SetOutput(l.logger.Out)
	newLogrusLogger.SetLevel(l.logger.Level)
	
	// Create a new LogrusLogger with the fields
	newLogger := &LogrusLogger{
		logger: newLogrusLogger,
	}
	
	// Add the fields to every log entry
	entry := newLogger.logger.WithFields(logrus.Fields(fields))
	newLogger.logger = entry.Logger
	
	return newLogger
}

// SetOutput sets the output destination for the logger
func (l *LogrusLogger) SetOutput(out io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.SetOutput(out)
}

// SetLevel sets the minimum log level
func (l *LogrusLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	switch level {
	case DebugLevel:
		l.logger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		l.logger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		l.logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		l.logger.SetLevel(logrus.ErrorLevel)
	case FatalLevel:
		l.logger.SetLevel(logrus.FatalLevel)
	default:
		l.logger.SetLevel(logrus.InfoLevel)
	}
}

// GetLevel returns the current log level
func (l *LogrusLogger) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	switch l.logger.GetLevel() {
	case logrus.DebugLevel:
		return DebugLevel
	case logrus.InfoLevel:
		return InfoLevel
	case logrus.WarnLevel:
		return WarnLevel
	case logrus.ErrorLevel:
		return ErrorLevel
	case logrus.FatalLevel:
		return FatalLevel
	default:
		return InfoLevel
	}
}

// Global logger instance
var (
	globalLogger Logger
	once         sync.Once
)

// GetLogger returns the global logger instance
func GetLogger() Logger {
	once.Do(func() {
		globalLogger = NewLogrusLogger()
	})
	return globalLogger
}

// SetLogger sets the global logger instance
func SetLogger(logger Logger) {
	globalLogger = logger
}
