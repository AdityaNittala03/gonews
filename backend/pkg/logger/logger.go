package logger

import (
	"log/slog"
	"os"
)

var log *slog.Logger

// Init initializes the structured logger
func Init() {
	// Create a new logger with JSON output for production
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	// Use JSON handler for production, text handler for development
	env := os.Getenv("ENVIRONMENT")
	if env == "development" {
		log = slog.New(slog.NewTextHandler(os.Stdout, opts))
	} else {
		log = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	// Set as default logger
	slog.SetDefault(log)
}

// Info logs an info message with optional key-value pairs
func Info(msg string, args ...interface{}) {
	log.Info(msg, args...)
}

// Error logs an error message with optional key-value pairs
func Error(msg string, args ...interface{}) {
	log.Error(msg, args...)
}

// Warn logs a warning message with optional key-value pairs
func Warn(msg string, args ...interface{}) {
	log.Warn(msg, args...)
}

// Debug logs a debug message with optional key-value pairs
func Debug(msg string, args ...interface{}) {
	log.Debug(msg, args...)
}

// Fatal logs a fatal error and exits the program
func Fatal(msg string, args ...interface{}) {
	log.Error(msg, args...)
	os.Exit(1)
}

// With returns a new logger with the given key-value pairs added as context
func With(args ...interface{}) *slog.Logger {
	return log.With(args...)
}
