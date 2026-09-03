package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger appends timestamped lines to a local file. Nothing is sent over the network.
type Logger struct {
	mu   sync.Mutex
	path string
}

// New creates a logger that writes to path, creating parent directories as needed.
func New(path string) (*Logger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	return &Logger{path: path}, nil
}

// Path returns the log file location.
func (l *Logger) Path() string {
	if l == nil {
		return ""
	}
	return l.path
}

// Info writes an informational line.
func (l *Logger) Info(format string, args ...any) {
	l.write("INFO", format, args...)
}

// Error writes an error line.
func (l *Logger) Error(format string, args ...any) {
	l.write("ERROR", format, args...)
}

func (l *Logger) write(level, format string, args ...any) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()

	line := fmt.Sprintf("%s [%s] %s\n", time.Now().Format(time.RFC3339), level, fmt.Sprintf(format, args...))
	_, _ = f.WriteString(line)
}
