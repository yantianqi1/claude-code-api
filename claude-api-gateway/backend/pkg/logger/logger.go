package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
)

// Initialize initializes the logger
func Initialize(debug bool) {
	InfoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)

	if debug {
		DebugLogger = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		DebugLogger = log.New(io.Discard, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	}
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	if DebugLogger != nil {
		DebugLogger.Output(2, fmt.Sprintf(format, v...))
	}
}
