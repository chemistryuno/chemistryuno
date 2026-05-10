package utils

import (
	"net/http"
	"testing"
)

func withTestLogger(t *testing.T, maxSize int) {
	t.Helper()
	previous := globalLogger
	globalLogger = &LogBuffer{
		entries:          make([]LogEntry, 0, maxSize),
		maxSize:          maxSize,
		subscribers:      make(map[int]chan LogEntry),
		nextSubscriberID: 1,
		nextSequence:     1,
	}
	t.Cleanup(func() {
		globalLogger = previous
	})
}

func TestParseLogLinePreservesPlainEntry(t *testing.T) {
	entry, ok := parseLogLine("2026/05/06 17:22:19 [WARNING] something happened")
	if !ok {
		t.Fatal("expected log line to parse")
	}
	if entry.Timestamp != "2026-05-06 17:22:19" {
		t.Fatalf("unexpected timestamp: %s", entry.Timestamp)
	}
	if entry.Level != "WARNING" || entry.Message != "something happened" {
		t.Fatalf("unexpected parsed entry: %+v", entry)
	}
}

func TestLogStructuredAddsBoundedStructuredEntry(t *testing.T) {
	withTestLogger(t, 1)

	uid := 100000101
	LogStructured(LogEntry{
		Level:    "INFO",
		Category: "request",
		Message:  "GET /api/rooms",
		UID:      &uid,
		Source: &LogSource{
			ClientIP:  "127.0.0.1",
			UserAgent: "test-agent",
		},
		Request: &LogRequest{
			Method:      "GET",
			Path:        "/api/rooms",
			Status:      200,
			StatusClass: "2xx",
		},
	})
	LogStructured(LogEntry{Level: "ERROR", Message: "latest"})

	logs := GetLogs(10)
	if len(logs) != 1 {
		t.Fatalf("expected bounded buffer to keep 1 entry, got %d", len(logs))
	}
	if logs[0].Message != "latest" || logs[0].Level != "ERROR" {
		t.Fatalf("unexpected latest entry: %+v", logs[0])
	}
	if logs[0].Sequence == 0 {
		t.Fatalf("expected stored log to include sequence")
	}
}

func TestLogFilteringByAttribution(t *testing.T) {
	withTestLogger(t, 10)

	uid := 100000101
	otherUID := 100000102
	LogStructured(LogEntry{
		Level:    "INFO",
		Category: "request",
		Message:  "GET /api/rooms",
		UID:      &uid,
		Source:   &LogSource{ClientIP: "10.0.0.7", ForwardedFor: "203.0.113.10"},
		Request:  &LogRequest{Method: "GET", Path: "/api/rooms", Status: 404, StatusClass: "4xx"},
	})
	LogStructured(LogEntry{
		Level:    "INFO",
		Category: "request",
		Message:  "POST /api/chat",
		UID:      &otherUID,
		Source:   &LogSource{ClientIP: "10.0.0.8"},
		Request:  &LogRequest{Method: "POST", Path: "/api/chat", Status: 200, StatusClass: "2xx"},
	})

	logs := GetLogsFiltered(LogFilter{UID: &uid, SourceIP: "203.0.113", StatusClass: "4xx", Keyword: "rooms"}, 10)
	if len(logs) != 1 {
		t.Fatalf("expected one filtered log, got %d", len(logs))
	}
	if logs[0].UID == nil || *logs[0].UID != uid {
		t.Fatalf("expected uid %d, got %+v", uid, logs[0].UID)
	}
}

func TestRedactionHelpersMaskSensitiveValues(t *testing.T) {
	sanitized := SanitizeURLPath("/api/rooms?token=abc123&key=secret-room&page=1")
	if sanitized == "/api/rooms?token=abc123&key=secret-room&page=1" {
		t.Fatalf("expected url to be sanitized")
	}
	if containsAny(sanitized, "abc123", "secret-room") {
		t.Fatalf("sanitized url leaked sensitive value: %s", sanitized)
	}

	if got := SanitizeHeaderValue("Authorization", "Bearer abc123"); got != redactedLogValue {
		t.Fatalf("expected authorization header to be redacted, got %q", got)
	}
	if got := RedactSensitiveString("token=abc123 code:xyz"); containsAny(got, "abc123", "xyz") {
		t.Fatalf("redacted string leaked sensitive values: %q", got)
	}
}

func TestSourceFromRequestNormalizesSource(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/api/test", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Origin", "http://localhost:5000")
	req.Header.Set("Referer", "http://localhost:5000/admin?token=abc")
	req.Header.Set("User-Agent", "Mozilla/5.0    Test")
	req.Header.Set("X-Forwarded-For", "203.0.113.10")

	source := SourceFromRequest(req, "127.0.0.1")
	if source.ClientIP != "127.0.0.1" || source.ForwardedFor != "203.0.113.10" {
		t.Fatalf("unexpected source IP metadata: %+v", source)
	}
	if source.UserAgent != "Mozilla/5.0 Test" {
		t.Fatalf("expected normalized user agent, got %q", source.UserAgent)
	}
	if containsAny(source.Referer, "abc") {
		t.Fatalf("referer leaked token: %q", source.Referer)
	}
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if needle != "" && contains(value, needle) {
			return true
		}
	}
	return false
}

func contains(value string, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
