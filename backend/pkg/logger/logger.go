package logger

import (
	"log/slog"
	"os"
)

// Logger wraps slog.Logger for structured logging
type Logger struct {
	logger *slog.Logger
}

// NewLogger creates a new structured logger instance
func NewLogger() *Logger {
	// Create a new logger with JSON output for production
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	var handler slog.Handler
	// Use JSON handler for production, text handler for development
	env := os.Getenv("ENVIRONMENT")
	if env == "development" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return &Logger{
		logger: slog.New(handler),
	}
}

// Info logs an info message with optional key-value pairs
func (l *Logger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

// Error logs an error message with optional key-value pairs
func (l *Logger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

// Warn logs a warning message with optional key-value pairs
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

// Debug logs a debug message with optional key-value pairs
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// Fatal logs a fatal error and exits the program
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
	os.Exit(1)
}

// With returns a new logger with the given key-value pairs added as context
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{
		logger: l.logger.With(args...),
	}
}

// Package-level functions for backward compatibility
var defaultLogger *Logger

// Init initializes the default logger (for backward compatibility)
func Init() {
	defaultLogger = NewLogger()
	slog.SetDefault(defaultLogger.logger)
}

// Info logs an info message using the default logger
func Info(msg string, args ...interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Info(msg, args...)
}

// Error logs an error message using the default logger
func Error(msg string, args ...interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Error(msg, args...)
}

// Warn logs a warning message using the default logger
func Warn(msg string, args ...interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Warn(msg, args...)
}

// Debug logs a debug message using the default logger
func Debug(msg string, args ...interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Debug(msg, args...)
}

// Fatal logs a fatal error and exits the program using the default logger
func Fatal(msg string, args ...interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Fatal(msg, args...)
}

// With returns a new logger with the given key-value pairs added as context
func With(args ...interface{}) *Logger {
	if defaultLogger == nil {
		Init()
	}
	return defaultLogger.With(args...)
}
