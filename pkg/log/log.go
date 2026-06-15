// Package log provides simple file-based debug logging for Shifter.
// Logs are written to ~/.shifter/logs/shifter-YYYYMMDD.log
package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/moximxxx/shifter/pkg/paths"
)

var (
	mu       sync.Mutex
	enabled  bool
	logFile  *os.File
)

// Init starts logging to ~/.shifter/logs/.
func Init() error {
	mu.Lock()
	defer mu.Unlock()

	home := paths.MustHomeDir()
	logDir := filepath.Join(home, ".shifter", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	filename := "shifter-" + time.Now().Format("2006-01-02") + ".log"
	path := filepath.Join(logDir, filename)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	logFile = f
	enabled = true
	// Use write directly to avoid double-lock (Init holds mu)
	ts := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(f, "%s [INFO] log: session started\n", ts)
	return nil
}

// Enabled returns true if logging is active.
func Enabled() bool { return enabled }

// Info logs an informational message.
func Info(category, format string, args ...interface{}) {
	if !enabled {
		return
	}
	write("INFO", category, format, args...)
}

// Error logs an error message.
func Error(category, format string, args ...interface{}) {
	if !enabled {
		return
	}
	write("ERROR", category, format, args...)
}

func write(level, category, format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if logFile == nil {
		return
	}
	ts := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(logFile, "%s [%s] %s: %s\n", ts, level, category, msg)
}

// Close flushes and closes the log file.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
	enabled = false
}
