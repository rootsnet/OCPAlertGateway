package logger

import (
	"log"
	"os"
)

// Logger provides separated loggers for informational
// and error-level logging.
//
// Info logs are written to stdout.
// Err logs are written to stderr and prefixed with "[ERROR]".
type Logger struct {
	// Info is used for normal operational logs.
	Info *log.Logger
	// Err is used for error logs.
	Err *log.Logger
}

// New creates and returns a new Logger instance.
//
// - Info logger writes to os.Stdout with standard timestamp flags.
// - Err logger writes to os.Stderr with "[ERROR]" prefix and standard timestamp flags.
//
// Standard flags (log.LstdFlags) include date and time.
func New() *Logger {
	return &Logger{
		Info: log.New(os.Stdout, "", log.LstdFlags),
		Err:  log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
	}
}
