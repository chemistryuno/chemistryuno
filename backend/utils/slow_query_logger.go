package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// SlowQueryLog represents a structured log entry for slow database queries
type SlowQueryLog struct {
	Timestamp  string  `json:"timestamp"`
	Duration   float64 `json:"duration_ms"`
	Table      string  `json:"table"`
	Operation  string  `json:"operation"`
	Query      string  `json:"query,omitempty"`
	Caller     string  `json:"caller"`
	Level      string  `json:"level"`
	Message    string  `json:"message"`
}

var (
	slowQueryThresholdMs = 100.0
	slowQueryLogLevel    = "warn"
)

func init() {
	// Read threshold from environment variable
	if threshold := os.Getenv("SLOW_QUERY_THRESHOLD_MS"); threshold != "" {
		var t float64
		if _, err := fmt.Sscanf(threshold, "%f", &t); err == nil && t > 0 {
			slowQueryThresholdMs = t
		}
	}

	// Read log level from environment variable
	if level := os.Getenv("SLOW_QUERY_LOG_LEVEL"); level != "" {
		slowQueryLogLevel = strings.ToLower(level)
	}
}

// LogSlowQuery logs a slow database query as structured JSON
func LogSlowQuery(duration time.Duration, table, operation, query string) {
	durationMs := float64(duration.Microseconds()) / 1000.0

	// Only log if duration exceeds threshold
	if durationMs < slowQueryThresholdMs {
		return
	}

	// Get caller information (skip 1 frame for this function)
	_, file, line, ok := runtime.Caller(1)
	caller := "unknown"
	if ok {
		// Simplify file path to relative from project root
		if idx := strings.LastIndex(file, "chemistryuno/"); idx >= 0 {
			caller = file[idx+len("chemistryuno/"):]
		} else {
			caller = file
		}
		caller = fmt.Sprintf("%s:%d", caller, line)
	}

	logEntry := SlowQueryLog{
		Timestamp: time.Now().Format(time.RFC3339),
		Duration:  durationMs,
		Table:     table,
		Operation: operation,
		Query:     sanitizeQuery(query),
		Caller:    caller,
		Level:     slowQueryLogLevel,
		Message:   fmt.Sprintf("Slow query detected: %s %s (%.2fms)", operation, table, durationMs),
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("Failed to marshal slow query log: %v", err)
		return
	}

	// Log based on configured level
	switch slowQueryLogLevel {
	case "debug":
		log.Printf("[DEBUG] %s", string(jsonBytes))
	case "info":
		log.Printf("[INFO] %s", string(jsonBytes))
	case "warn":
		log.Printf("[WARN] %s", string(jsonBytes))
	case "error":
		log.Printf("[ERROR] %s", string(jsonBytes))
	default:
		log.Printf("[WARN] %s", string(jsonBytes))
	}
}

// sanitizeQuery removes potential PII from query strings
func sanitizeQuery(query string) string {
	// Don't log full query if it's too long or contains sensitive data
	if len(query) > 200 {
		return query[:200] + "..."
	}
	return query
}
