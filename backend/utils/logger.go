package utils

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// LogEntry 日志条目结构
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"` // info, warning, error, debug
	Message   string `json:"message"`
}

// LogBuffer 日志缓冲区管理器
type LogBuffer struct {
	entries   []LogEntry
	maxSize   int
	mutex     sync.RWMutex
	logFile   *os.File
	logWriter *log.Logger
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
		entries:   make([]LogEntry, 0, maxSize),
		maxSize:   maxSize,
		logFile:   logFile,
		logWriter: log.New(logFile, "", log.LstdFlags),
	}

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
	entry := LogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     level,
		Message:   message,
	}

	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// 写入文件
	lb.logWriter.Printf("[%s] %s", entry.Level, entry.Message)

	// 写入内存缓冲
	lb.entries = append(lb.entries, entry)

	// 如果超过最大大小，删除最旧的日志
	if len(lb.entries) > lb.maxSize {
		lb.entries = lb.entries[1:]
	}

	// 同时输出到标准日志
	log.Printf("[%s] %s", level, message)
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
	if globalLogger == nil {
		return []LogEntry{}
	}

	globalLogger.mutex.RLock()
	defer globalLogger.mutex.RUnlock()

	var filtered []LogEntry
	for _, entry := range globalLogger.entries {
		if entry.Level == level {
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
