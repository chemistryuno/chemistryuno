package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LogEntry 日志条目结构
type LogEntry struct {
	Sequence     uint64                 `json:"sequence,omitempty"`
	Timestamp    string                 `json:"timestamp"`
	Level        string                 `json:"level"` // info, warning, error, debug
	Message      string                 `json:"message"`
	Category     string                 `json:"category,omitempty"`
	UID          *int                   `json:"uid,omitempty"`
	AttemptedUID *int                   `json:"attempted_uid,omitempty"`
	Role         string                 `json:"role,omitempty"`
	AuthState    string                 `json:"auth_state,omitempty"`
	Source       *LogSource             `json:"source,omitempty"`
	Request      *LogRequest            `json:"request,omitempty"`
	WebSocket    *LogWebSocket          `json:"websocket,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

type LogSource struct {
	ClientIP       string `json:"client_ip,omitempty"`
	ForwardedFor   string `json:"forwarded_for,omitempty"`
	RealIP         string `json:"real_ip,omitempty"`
	ForwardedProto string `json:"forwarded_proto,omitempty"`
	ForwardedHost  string `json:"forwarded_host,omitempty"`
	Origin         string `json:"origin,omitempty"`
	Referer        string `json:"referer,omitempty"`
	UserAgent      string `json:"user_agent,omitempty"`
}

type LogRequest struct {
	Method      string `json:"method,omitempty"`
	Path        string `json:"path,omitempty"`
	Route       string `json:"route,omitempty"`
	Status      int    `json:"status,omitempty"`
	StatusClass string `json:"status_class,omitempty"`
	LatencyMs   int64  `json:"latency_ms,omitempty"`
	BytesIn     int64  `json:"bytes_in,omitempty"`
	BytesOut    int    `json:"bytes_out,omitempty"`
}

type LogWebSocket struct {
	Event  string `json:"event,omitempty"`
	Type   string `json:"type,omitempty"`
	RoomID string `json:"room_id,omitempty"`
}

type LogFilter struct {
	Level        string
	Category     string
	UID          *int
	AttemptedUID *int
	SourceIP     string
	StatusClass  string
	Keyword      string
}

// LogBuffer 日志缓冲区管理器
type LogBuffer struct {
	entries          []LogEntry
	maxSize          int
	mutex            sync.RWMutex
	logFile          *os.File
	subscribers      map[int]chan LogEntry
	nextSubscriberID int
	nextSequence     uint64
}

type logCaptureWriter struct {
	buffer *LogBuffer
}

var stdLogPrefixRe = regexp.MustCompile(`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})\s+(.*)$`)
var sensitiveLogKeys = map[string]bool{
	"access_token":  true,
	"authorization": true,
	"code":          true,
	"cookie":        true,
	"key":           true,
	"password":      true,
	"refresh_token": true,
	"secret":        true,
	"sid":           true,
	"state":         true,
	"token":         true,
}

const redactedLogValue = "[REDACTED]"

func parseLogLine(rawLine string) (LogEntry, bool) {
	line := strings.TrimSpace(rawLine)
	if line == "" {
		return LogEntry{}, false
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := line

	if matches := stdLogPrefixRe.FindStringSubmatch(line); len(matches) == 3 {
		if parsed, err := time.Parse("2006/01/02 15:04:05", matches[1]); err == nil {
			timestamp = parsed.Format("2006-01-02 15:04:05")
		}
		message = strings.TrimSpace(matches[2])
	}

	level := "INFO"
	for _, candidate := range []string{"ERROR", "WARNING", "INFO", "DEBUG"} {
		if strings.HasPrefix(message, "["+candidate+"]") {
			level = candidate
			message = strings.TrimSpace(strings.TrimPrefix(message, "["+candidate+"]"))
			break
		}
	}

	return LogEntry{
		Timestamp: timestamp,
		Level:     level,
		Message:   message,
	}, true
}

func (w *logCaptureWriter) Write(p []byte) (int, error) {
	if w == nil || w.buffer == nil {
		return len(p), nil
	}

	lines := strings.Split(string(p), "\n")
	w.buffer.mutex.Lock()
	defer w.buffer.mutex.Unlock()

	for _, line := range lines {
		entry, ok := parseLogLine(line)
		if !ok {
			continue
		}
		w.buffer.appendEntryLocked(entry)
	}

	return len(p), nil
}

// SubscribeLogs 订阅实时日志流。
func SubscribeLogs() (int, <-chan LogEntry) {
	if globalLogger == nil {
		ch := make(chan LogEntry)
		close(ch)
		return 0, ch
	}

	globalLogger.mutex.Lock()
	defer globalLogger.mutex.Unlock()

	id := globalLogger.nextSubscriberID
	globalLogger.nextSubscriberID++

	ch := make(chan LogEntry, 256)
	globalLogger.subscribers[id] = ch

	return id, ch
}

// UnsubscribeLogs 取消实时日志订阅。
func UnsubscribeLogs(id int) {
	if globalLogger == nil {
		return
	}

	globalLogger.mutex.Lock()
	defer globalLogger.mutex.Unlock()

	if ch, ok := globalLogger.subscribers[id]; ok {
		delete(globalLogger.subscribers, id)
		close(ch)
	}
}

var globalLogger *LogBuffer

// InitLogger 初始化全局日志管理器
func InitLogger(maxSize int) error {
	// 确保logs目录存在
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	logFile, err := os.OpenFile("logs/game.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	globalLogger = &LogBuffer{
		entries:          make([]LogEntry, 0, maxSize),
		maxSize:          maxSize,
		logFile:          logFile,
		subscribers:      make(map[int]chan LogEntry),
		nextSubscriberID: 1,
		nextSequence:     1,
	}

	// 将标准日志输出同时写入控制台、文件，并实时捕获到内存缓冲供 /admin/logs 使用。
	log.SetOutput(io.MultiWriter(os.Stdout, logFile, &logCaptureWriter{buffer: globalLogger}))

	return nil
}

// Log 记录日志
func Log(level, message string) {
	if globalLogger == nil {
		log.Printf("[%s] %s", level, message)
		return
	}

	globalLogger.log(level, message)
}

func LogStructured(entry LogEntry) {
	entry.Level = strings.ToUpper(strings.TrimSpace(entry.Level))
	if entry.Level == "" {
		entry.Level = "INFO"
	}
	if strings.TrimSpace(entry.Timestamp) == "" {
		entry.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	}
	if strings.TrimSpace(entry.Message) == "" {
		entry.Message = entry.Category
	}
	entry.Message = RedactSensitiveString(entry.Message)

	if globalLogger == nil {
		log.Printf("[%s] %s", entry.Level, entry.Message)
		return
	}

	globalLogger.mutex.Lock()
	globalLogger.appendEntryLocked(entry)
	globalLogger.mutex.Unlock()
	globalLogger.writePlainLine(entry)
}

// LogInfo 记录信息日志
func LogInfo(message string) {
	Log("INFO", message)
}

// LogWarning 记录警告日志
func LogWarning(message string) {
	Log("WARNING", message)
}

// LogError 记录错误日志
func LogError(message string) {
	Log("ERROR", message)
}

// LogDebug 记录调试日志
func LogDebug(message string) {
	Log("DEBUG", message)
}

// log 内部日志方法
func (lb *LogBuffer) log(level, message string) {
	log.Printf("[%s] %s", strings.ToUpper(level), message)
}

func (lb *LogBuffer) appendEntryLocked(entry LogEntry) {
	if entry.Sequence == 0 {
		entry.Sequence = lb.nextSequence
		lb.nextSequence++
	}
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	}
	entry.Level = strings.ToUpper(strings.TrimSpace(entry.Level))
	if entry.Level == "" {
		entry.Level = "INFO"
	}
	entry.Message = RedactSensitiveString(entry.Message)
	lb.entries = append(lb.entries, entry)
	if len(lb.entries) > lb.maxSize {
		lb.entries = lb.entries[1:]
	}

	for _, subscriber := range lb.subscribers {
		select {
		case subscriber <- entry:
		default:
			// 丢弃慢订阅者当前消息，避免阻塞日志主链路。
		}
	}
}

func (lb *LogBuffer) writePlainLine(entry LogEntry) {
	line := fmt.Sprintf("%s [%s] %s\n", entry.Timestamp, entry.Level, entry.Message)
	_, _ = fmt.Fprint(os.Stdout, line)
	if lb.logFile != nil {
		_, _ = lb.logFile.WriteString(line)
	}
}

// GetLogs 获取最近的日志（按时间倒序）
func GetLogs(count int) []LogEntry {
	if globalLogger == nil {
		return []LogEntry{}
	}

	globalLogger.mutex.RLock()
	defer globalLogger.mutex.RUnlock()

	total := len(globalLogger.entries)
	if count <= 0 || count > total {
		count = total
	}

	// 返回最后count条日志（倒序）
	result := make([]LogEntry, count)
	for i := 0; i < count; i++ {
		result[i] = globalLogger.entries[total-count+i]
	}

	// 反转以得到倒序
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// GetLogsByLevel 获取指定级别的日志
func GetLogsByLevel(level string, count int) []LogEntry {
	return GetLogsFiltered(LogFilter{Level: level}, count)
}

func GetLogsFiltered(filter LogFilter, count int) []LogEntry {
	if globalLogger == nil {
		return []LogEntry{}
	}

	globalLogger.mutex.RLock()
	defer globalLogger.mutex.RUnlock()

	var filtered []LogEntry
	filter.Level = strings.ToUpper(strings.TrimSpace(filter.Level))
	filter.Category = strings.ToLower(strings.TrimSpace(filter.Category))
	filter.StatusClass = strings.ToLower(strings.TrimSpace(filter.StatusClass))
	filter.SourceIP = strings.TrimSpace(filter.SourceIP)
	filter.Keyword = strings.ToLower(strings.TrimSpace(filter.Keyword))
	for _, entry := range globalLogger.entries {
		if logEntryMatchesFilter(entry, filter) {
			filtered = append(filtered, entry)
		}
	}

	if count <= 0 || count > len(filtered) {
		count = len(filtered)
	}

	// 返回最后count条日志（倒序）
	result := make([]LogEntry, count)
	total := len(filtered)
	for i := 0; i < count; i++ {
		result[i] = filtered[total-count+i]
	}

	// 反转以得到倒序
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

func LogEntryMatchesFilter(entry LogEntry, filter LogFilter) bool {
	filter.Level = strings.ToUpper(strings.TrimSpace(filter.Level))
	filter.Category = strings.ToLower(strings.TrimSpace(filter.Category))
	filter.StatusClass = strings.ToLower(strings.TrimSpace(filter.StatusClass))
	filter.SourceIP = strings.TrimSpace(filter.SourceIP)
	filter.Keyword = strings.ToLower(strings.TrimSpace(filter.Keyword))
	return logEntryMatchesFilter(entry, filter)
}

func logEntryMatchesFilter(entry LogEntry, filter LogFilter) bool {
	if filter.Level != "" && strings.ToUpper(entry.Level) != filter.Level {
		return false
	}
	if filter.Category != "" && strings.ToLower(entry.Category) != filter.Category {
		return false
	}
	if filter.UID != nil {
		if entry.UID == nil || *entry.UID != *filter.UID {
			return false
		}
	}
	if filter.AttemptedUID != nil {
		if entry.AttemptedUID == nil || *entry.AttemptedUID != *filter.AttemptedUID {
			return false
		}
	}
	if filter.SourceIP != "" && !entrySourceMatches(entry.Source, filter.SourceIP) {
		return false
	}
	if filter.StatusClass != "" {
		if entry.Request == nil || strings.ToLower(entry.Request.StatusClass) != filter.StatusClass {
			return false
		}
	}
	if filter.Keyword != "" && !strings.Contains(strings.ToLower(entrySearchText(entry)), filter.Keyword) {
		return false
	}
	return true
}

func entrySourceMatches(source *LogSource, ip string) bool {
	if source == nil {
		return false
	}
	return strings.Contains(source.ClientIP, ip) ||
		strings.Contains(source.ForwardedFor, ip) ||
		strings.Contains(source.RealIP, ip)
}

func entrySearchText(entry LogEntry) string {
	parts := []string{entry.Message, entry.Category, entry.Level, entry.Role, entry.AuthState}
	if entry.Source != nil {
		parts = append(parts,
			entry.Source.ClientIP,
			entry.Source.ForwardedFor,
			entry.Source.RealIP,
			entry.Source.Origin,
			entry.Source.Referer,
			entry.Source.UserAgent,
		)
	}
	if entry.Request != nil {
		parts = append(parts,
			entry.Request.Method,
			entry.Request.Path,
			entry.Request.Route,
			entry.Request.StatusClass,
			strconv.Itoa(entry.Request.Status),
		)
	}
	if entry.WebSocket != nil {
		parts = append(parts, entry.WebSocket.Event, entry.WebSocket.Type, entry.WebSocket.RoomID)
	}
	return strings.Join(parts, " ")
}

// ClearLogs 清空日志缓冲
func ClearLogs() {
	if globalLogger == nil {
		return
	}

	globalLogger.mutex.Lock()
	defer globalLogger.mutex.Unlock()

	globalLogger.entries = make([]LogEntry, 0, globalLogger.maxSize)
}

// Close 关闭日志管理器
func (lb *LogBuffer) Close() error {
	if lb.logFile != nil {
		return lb.logFile.Close()
	}
	return nil
}

// CloseLogger 关闭全局日志管理器
func CloseLogger() error {
	if globalLogger != nil {
		return globalLogger.Close()
	}
	return nil
}

func RedactSensitiveString(raw string) string {
	if raw == "" {
		return raw
	}
	redacted := raw
	for key := range sensitiveLogKeys {
		pattern := regexp.MustCompile(`(?i)(` + regexp.QuoteMeta(key) + `)(=|%3D|:)\s*([^&\s,;]+)`)
		redacted = pattern.ReplaceAllString(redacted, `${1}${2}`+redactedLogValue)
	}
	return redacted
}

func SanitizeURLPath(rawPath string) string {
	if rawPath == "" {
		return rawPath
	}
	parsed, err := url.Parse(rawPath)
	if err != nil {
		return RedactSensitiveString(rawPath)
	}
	query := parsed.Query()
	for key := range query {
		if isSensitiveLogKey(key) {
			query.Set(key, redactedLogValue)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func SanitizeHeaderValue(key string, value string) string {
	if value == "" {
		return ""
	}
	if isSensitiveLogKey(key) {
		return redactedLogValue
	}
	return RedactSensitiveString(value)
}

func NormalizeUserAgent(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 180 {
		return value[:177] + "..."
	}
	return value
}

func SourceFromRequest(r *http.Request, clientIP string) *LogSource {
	if r == nil {
		return &LogSource{ClientIP: clientIP}
	}
	return &LogSource{
		ClientIP:       clientIP,
		ForwardedFor:   SanitizeHeaderValue("X-Forwarded-For", r.Header.Get("X-Forwarded-For")),
		RealIP:         SanitizeHeaderValue("X-Real-IP", r.Header.Get("X-Real-IP")),
		ForwardedProto: SanitizeHeaderValue("X-Forwarded-Proto", r.Header.Get("X-Forwarded-Proto")),
		ForwardedHost:  SanitizeHeaderValue("X-Forwarded-Host", r.Header.Get("X-Forwarded-Host")),
		Origin:         SanitizeHeaderValue("Origin", r.Header.Get("Origin")),
		Referer:        SanitizeHeaderValue("Referer", r.Header.Get("Referer")),
		UserAgent:      NormalizeUserAgent(SanitizeHeaderValue("User-Agent", r.Header.Get("User-Agent"))),
	}
}

func StatusClass(status int) string {
	if status <= 0 {
		return ""
	}
	return fmt.Sprintf("%dxx", status/100)
}

func IntPtr(value int) *int {
	return &value
}

func isSensitiveLogKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if sensitiveLogKeys[normalized] {
		return true
	}
	return strings.Contains(normalized, "token") ||
		strings.Contains(normalized, "password") ||
		strings.Contains(normalized, "secret") ||
		strings.Contains(normalized, "authorization") ||
		strings.Contains(normalized, "cookie")
}
