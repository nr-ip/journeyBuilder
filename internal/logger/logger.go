package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	// DefaultLogger is the global logger instance
	DefaultLogger *log.Logger
	// LogFile is the file handle for the log file (if enabled)
	LogFile *os.File
)

// InitLogger initializes the logger with file and/or console output
// If LOG_FILE is set, logs will be written to that file
// If LOG_FILE is empty, logs only go to stderr (console)
// Logs are always written to stderr, and optionally to a file
func InitLogger() error {
	var writers []io.Writer

	// Always write to stderr (console)
	writers = append(writers, os.Stderr)

	// Check if file logging is enabled
	logFile := os.Getenv("LOG_FILE")
	if logFile != "" {
		// Create logs directory if it doesn't exist
		logDir := filepath.Dir(logFile)
		if logDir != "." && logDir != "" {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return err
			}
		}

		// Open log file in append mode
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}

		LogFile = file
		writers = append(writers, file)
	}

	// Create multi-writer that writes to all destinations
	multiWriter := io.MultiWriter(writers...)

	// Initialize logger with timestamp, date, and file location
	DefaultLogger = log.New(multiWriter, "", log.LstdFlags|log.Lshortfile)

	// Set the standard log package to use our logger
	log.SetOutput(multiWriter)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return nil
}

// Close closes the log file if it was opened
func Close() error {
	if LogFile != nil {
		return LogFile.Close()
	}
	return nil
}

// GetLogger returns the default logger instance
func GetLogger() *log.Logger {
	if DefaultLogger == nil {
		// Fallback to standard logger if not initialized
		return log.Default()
	}
	return DefaultLogger
}
