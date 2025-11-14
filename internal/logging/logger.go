// internal/logging/logger.go - Structured logging with context
package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// LogLevel represents logging severity
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

// Logger provides structured logging
type Logger struct {
	serviceName string
	environment string
	minLevel    LogLevel
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time      `json:"timestamp"`
	Level       LogLevel       `json:"level"`
	Service     string         `json:"service"`
	Environment string         `json:"environment"`
	Message     string         `json:"message"`
	RequestID   string         `json:"request_id,omitempty"`
	TraceID     string         `json:"trace_id,omitempty"`
	UserID      string         `json:"user_id,omitempty"`
	Method      string         `json:"method,omitempty"`
	Path        string         `json:"path,omitempty"`
	StatusCode  int            `json:"status_code,omitempty"`
	Duration    int64          `json:"duration_ms,omitempty"`
	Error       string         `json:"error,omitempty"`
	Fields      map[string]any `json:"fields,omitempty"`
}

// NewLogger creates a new structured logger
func NewLogger(serviceName, environment string) *Logger {
	minLevel := LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		minLevel = LevelDebug
	}

	return &Logger{
		serviceName: serviceName,
		environment: environment,
		minLevel:    minLevel,
	}
}

// log writes a log entry
func (l *Logger) log(entry LogEntry) {
	entry.Service = l.serviceName
	entry.Environment = l.environment
	entry.Timestamp = time.Now()

	// Skip if below minimum level
	if !l.shouldLog(entry.Level) {
		return
	}

	// JSON output
	if os.Getenv("LOG_FORMAT") == "json" {
		data, _ := json.Marshal(entry)
		fmt.Println(string(data))
	} else {
		// Human-readable format
		l.printReadable(entry)
	}
}

// shouldLog checks if level should be logged
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
		LevelFatal: 4,
	}
	return levels[level] >= levels[l.minLevel]
}

// printReadable outputs human-readable logs
func (l *Logger) printReadable(entry LogEntry) {
	color := ""
	reset := "\033[0m"

	switch entry.Level {
	case LevelDebug:
		color = "\033[36m" // Cyan
	case LevelInfo:
		color = "\033[32m" // Green
	case LevelWarn:
		color = "\033[33m" // Yellow
	case LevelError:
		color = "\033[31m" // Red
	case LevelFatal:
		color = "\033[35m" // Magenta
	}

	fmt.Printf("%s[%s]%s %s | %s",
		color, entry.Level, reset,
		entry.Timestamp.Format("2006-01-02 15:04:05"),
		entry.Message)

	if entry.RequestID != "" {
		fmt.Printf(" | req_id=%s", entry.RequestID)
	}
	if entry.UserID != "" {
		fmt.Printf(" | user_id=%s", entry.UserID)
	}
	if entry.Method != "" && entry.Path != "" {
		fmt.Printf(" | %s %s", entry.Method, entry.Path)
	}
	if entry.StatusCode > 0 {
		fmt.Printf(" | status=%d", entry.StatusCode)
	}
	if entry.Duration > 0 {
		fmt.Printf(" | duration=%dms", entry.Duration)
	}
	if entry.Error != "" {
		fmt.Printf(" | error=%s", entry.Error)
	}

	if len(entry.Fields) > 0 {
		fmt.Printf(" | fields=%v", entry.Fields)
	}

	fmt.Println()
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields map[string]any) {
	l.log(LogEntry{
		Level:   LevelDebug,
		Message: msg,
		Fields:  fields,
	})
}

// Info logs an info message
func (l *Logger) Info(msg string, fields map[string]any) {
	l.log(LogEntry{
		Level:   LevelInfo,
		Message: msg,
		Fields:  fields,
	})
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields map[string]any) {
	l.log(LogEntry{
		Level:   LevelWarn,
		Message: msg,
		Fields:  fields,
	})
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, fields map[string]any) {
	entry := LogEntry{
		Level:   LevelError,
		Message: msg,
		Fields:  fields,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	l.log(entry)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, err error, fields map[string]any) {
	entry := LogEntry{
		Level:   LevelFatal,
		Message: msg,
		Fields:  fields,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	l.log(entry)
	os.Exit(1)
}

// FromContext extracts request context for logging
func (l *Logger) FromContext(c echo.Context) *ContextLogger {
	requestID := c.Response().Header().Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}

	traceID := c.Response().Header().Get("X-Trace-ID")
	userID := ""
	if uid, ok := c.Get("user_id").(uuid.UUID); ok {
		userID = uid.String()
	}

	return &ContextLogger{
		logger:    l,
		requestID: requestID,
		traceID:   traceID,
		userID:    userID,
		method:    c.Request().Method,
		path:      c.Path(),
	}
}

// ContextLogger logs with request context
type ContextLogger struct {
	logger    *Logger
	requestID string
	traceID   string
	userID    string
	method    string
	path      string
}

// Info logs with context
func (cl *ContextLogger) Info(msg string, fields map[string]any) {
	cl.logger.log(LogEntry{
		Level:     LevelInfo,
		Message:   msg,
		RequestID: cl.requestID,
		TraceID:   cl.traceID,
		UserID:    cl.userID,
		Method:    cl.method,
		Path:      cl.path,
		Fields:    fields,
	})
}

// Error logs error with context
func (cl *ContextLogger) Error(msg string, err error, fields map[string]any) {
	entry := LogEntry{
		Level:     LevelError,
		Message:   msg,
		RequestID: cl.requestID,
		TraceID:   cl.traceID,
		UserID:    cl.userID,
		Method:    cl.method,
		Path:      cl.path,
		Fields:    fields,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	cl.logger.log(entry)
}

// LogRequest logs an HTTP request with full context
func (cl *ContextLogger) LogRequest(statusCode int, duration time.Duration) {
	level := LevelInfo
	if statusCode >= 500 {
		level = LevelError
	} else if statusCode >= 400 {
		level = LevelWarn
	}

	cl.logger.log(LogEntry{
		Level:      level,
		Message:    "HTTP Request",
		RequestID:  cl.requestID,
		TraceID:    cl.traceID,
		UserID:     cl.userID,
		Method:     cl.method,
		Path:       cl.path,
		StatusCode: statusCode,
		Duration:   duration.Milliseconds(),
	})
}

// LoggingMiddleware creates middleware with structured logging
func LoggingMiddleware(logger *Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			contextLogger := logger.FromContext(c)

			// Store logger in context
			c.Set("logger", contextLogger)

			// Execute request
			err := next(c)

			// Log request
			statusCode := c.Response().Status
			duration := time.Since(start)

			contextLogger.LogRequest(statusCode, duration)

			return err
		}
	}
}

// GetLogger retrieves logger from echo context
func GetLogger(c echo.Context) *ContextLogger {
	if logger, ok := c.Get("logger").(*ContextLogger); ok {
		return logger
	}
	// Fallback to creating new logger
	return NewLogger("digiorder", "production").FromContext(c)
}
