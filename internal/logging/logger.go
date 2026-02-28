// Package logging provides production logging for hookflow.
// Logs are written to a known location (~/.hookflow/logs/) with automatic rotation.
// Enable verbose logging with HOOKFLOW_DEBUG=1 or --verbose flag.
package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents log severity
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger is the main logging interface
type Logger struct {
	mu       sync.Mutex
	level    Level
	file     *os.File
	filePath string
	session  string // Unique session ID for correlating logs
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// logDir returns the hookflow log directory
func logDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to temp directory
		return filepath.Join(os.TempDir(), "hookflow", "logs")
	}
	return filepath.Join(home, ".hookflow", "logs")
}

// Init initializes the default logger. Call once at startup.
func Init() error {
	var initErr error
	once.Do(func() {
		dir := logDir()
		if err := os.MkdirAll(dir, 0755); err != nil {
			initErr = fmt.Errorf("failed to create log directory: %w", err)
			return
		}

		// Log file named by date for easy rotation
		today := time.Now().Format("2006-01-02")
		logFile := filepath.Join(dir, fmt.Sprintf("hookflow-%s.log", today))

		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			initErr = fmt.Errorf("failed to open log file: %w", err)
			return
		}

		// Generate session ID for correlating logs from same invocation
		sessionID := fmt.Sprintf("%d-%d", os.Getpid(), time.Now().UnixNano()%100000)

		// Determine log level from environment
		level := LevelInfo
		if os.Getenv("HOOKFLOW_DEBUG") == "1" || os.Getenv("HOOKFLOW_VERBOSE") == "1" {
			level = LevelDebug
		}

		defaultLogger = &Logger{
			level:    level,
			file:     f,
			filePath: logFile,
			session:  sessionID,
		}

		// Clean up old logs (keep last 7 days)
		go cleanOldLogs(dir, 7)
	})
	return initErr
}

// SetLevel sets the minimum log level
func SetLevel(level Level) {
	if defaultLogger != nil {
		defaultLogger.mu.Lock()
		defaultLogger.level = level
		defaultLogger.mu.Unlock()
	}
}

// EnableDebug enables debug-level logging
func EnableDebug() {
	SetLevel(LevelDebug)
}

// Close closes the log file
func Close() {
	if defaultLogger != nil && defaultLogger.file != nil {
		defaultLogger.file.Close()
	}
}

// LogPath returns the current log file path
func LogPath() string {
	if defaultLogger != nil {
		return defaultLogger.filePath
	}
	return ""
}

// LogDir returns the log directory path
func LogDir() string {
	return logDir()
}

// log writes a log entry
func log(level Level, format string, args ...interface{}) {
	if defaultLogger == nil {
		return
	}

	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()

	if level < defaultLogger.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)

	// Get caller info for debug logs
	caller := ""
	if level == LevelDebug {
		if _, file, line, ok := runtime.Caller(2); ok {
			caller = fmt.Sprintf(" [%s:%d]", filepath.Base(file), line)
		}
	}

	entry := fmt.Sprintf("[%s] [%s] [%s]%s %s\n",
		timestamp,
		level.String(),
		defaultLogger.session,
		caller,
		message,
	)

	defaultLogger.file.WriteString(entry)
}

// Debug logs at debug level
func Debug(format string, args ...interface{}) {
	log(LevelDebug, format, args...)
}

// Info logs at info level
func Info(format string, args ...interface{}) {
	log(LevelInfo, format, args...)
}

// Warn logs at warn level
func Warn(format string, args ...interface{}) {
	log(LevelWarn, format, args...)
}

// Error logs at error level
func Error(format string, args ...interface{}) {
	log(LevelError, format, args...)
}

// WithContext returns a contextual logger that prefixes all messages
type ContextLogger struct {
	prefix string
}

// Context creates a new contextual logger
func Context(prefix string) *ContextLogger {
	return &ContextLogger{prefix: prefix}
}

func (c *ContextLogger) Debug(format string, args ...interface{}) {
	Debug("[%s] "+format, append([]interface{}{c.prefix}, args...)...)
}

func (c *ContextLogger) Info(format string, args ...interface{}) {
	Info("[%s] "+format, append([]interface{}{c.prefix}, args...)...)
}

func (c *ContextLogger) Warn(format string, args ...interface{}) {
	Warn("[%s] "+format, append([]interface{}{c.prefix}, args...)...)
}

func (c *ContextLogger) Error(format string, args ...interface{}) {
	Error("[%s] "+format, append([]interface{}{c.prefix}, args...)...)
}

// cleanOldLogs removes log files older than maxDays
func cleanOldLogs(dir string, maxDays int) {
	cutoff := time.Now().AddDate(0, 0, -maxDays)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, "hookflow-") || !strings.HasSuffix(name, ".log") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(dir, name))
		}
	}
}

// Tee returns a writer that writes to both the log file and the provided writer
func Tee(w io.Writer) io.Writer {
	if defaultLogger == nil || defaultLogger.file == nil {
		return w
	}
	return io.MultiWriter(w, defaultLogger.file)
}

// StartOperation logs the start of an operation and returns a function to log completion
func StartOperation(name string, details ...string) func(error) {
	start := time.Now()
	detail := ""
	if len(details) > 0 {
		detail = " " + strings.Join(details, " ")
	}
	Debug("START %s%s", name, detail)

	return func(err error) {
		duration := time.Since(start)
		if err != nil {
			Error("FAIL %s%s (took %v): %v", name, detail, duration, err)
		} else {
			Debug("DONE %s%s (took %v)", name, detail, duration)
		}
	}
}
