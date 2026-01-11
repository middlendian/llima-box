// Package log provides colored logging utilities for llima-box CLI.
package log

import (
	"fmt"
	"io"
	"os"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// Logger provides structured, colored logging to stderr.
type Logger struct {
	output io.Writer
	colors bool
}

// New creates a new Logger that writes to stderr.
func New() *Logger {
	return &Logger{
		output: os.Stderr,
		colors: isTerminal(os.Stderr),
	}
}

// Info prints an informational message to stderr.
func (l *Logger) Info(format string, args ...interface{}) {
	l.print(colorCyan, "INFO", format, args...)
}

// Success prints a success message to stderr.
func (l *Logger) Success(format string, args ...interface{}) {
	l.print(colorGreen, "SUCCESS", format, args...)
}

// Warning prints a warning message to stderr.
func (l *Logger) Warning(format string, args ...interface{}) {
	l.print(colorYellow, "WARNING", format, args...)
}

// Error prints an error message to stderr.
func (l *Logger) Error(format string, args ...interface{}) {
	l.print(colorRed, "ERROR", format, args...)
}

// Debug prints a debug message to stderr (gray color).
func (l *Logger) Debug(format string, args ...interface{}) {
	l.print(colorGray, "DEBUG", format, args...)
}

// Plain prints a plain message to stderr without a prefix or color.
func (l *Logger) Plain(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintln(l.output, msg)
}

// print formats and prints a colored log message.
func (l *Logger) print(color, level, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if l.colors {
		_, _ = fmt.Fprintf(l.output, "%s%s%s: %s\n", color, level, colorReset, msg)
	} else {
		_, _ = fmt.Fprintf(l.output, "%s: %s\n", level, msg)
	}
}

// isTerminal returns true if the writer is a terminal.
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		stat, err := f.Stat()
		if err != nil {
			return false
		}
		return (stat.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// Default logger instance
var defaultLogger = New()

// Info prints an informational message using the default logger.
func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// Success prints a success message using the default logger.
func Success(format string, args ...interface{}) {
	defaultLogger.Success(format, args...)
}

// Warning prints a warning message using the default logger.
func Warning(format string, args ...interface{}) {
	defaultLogger.Warning(format, args...)
}

// Error prints an error message using the default logger.
func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

// Debug prints a debug message using the default logger.
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

// Plain prints a plain message using the default logger.
func Plain(format string, args ...interface{}) {
	defaultLogger.Plain(format, args...)
}
