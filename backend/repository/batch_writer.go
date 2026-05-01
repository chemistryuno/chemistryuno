package repository

import (
	"chemistryuno/backend/database"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// BatchWriter 提供批量写入功能的通用批处理器
type BatchWriter struct {
	db          *gorm.DB
	batchSize   int
	flushTicker *time.Ticker
	mu          sync.Mutex
	buffer      []interface{}
	modelType   interface{}
	closed      bool
	flushDone   chan struct{}
}

// NewBatchWriter 创建新的批量写入器
// modelType 是要批量插入的模型类型（如 GameHistory{}）
// batchSize 是缓冲区大小，默认100
// flushInterval 是自动刷新间隔，默认1秒
func NewBatchWriter(db *gorm.DB, modelType interface{}, batchSize int, flushInterval time.Duration) *BatchWriter {
	if batchSize <= 0 {
		batchSize = 100
	}
	if flushInterval <= 0 {
		flushInterval = 1 * time.Second
	}

	bw := &BatchWriter{
		db:        db,
		batchSize: batchSize,
		buffer:    make([]interface{}, 0, batchSize),
		modelType: modelType,
		flushDone: make(chan struct{}),
	}

	// 启动定时刷新
	bw.flushTicker = time.NewTicker(flushInterval)
	go bw.flushLoop()

	return bw
}

// Add 添加一条记录到缓冲区
func (bw *BatchWriter) Add(record interface{}) error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	if bw.closed {
		return gorm.ErrInvalidDB
	}

	bw.buffer = append(bw.buffer, record)

	// 如果缓冲区满了，立即刷新
	if len(bw.buffer) >= bw.batchSize {
		return bw.flushLocked()
	}

	return nil
}

// AddBatch 添加多条记录
func (bw *BatchWriter) AddBatch(records ...interface{}) error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	if bw.closed {
		return gorm.ErrInvalidDB
	}

	bw.buffer = append(bw.buffer, records...)

	// 如果缓冲区满了，立即刷新
	if len(bw.buffer) >= bw.batchSize {
		return bw.flushLocked()
	}

	return nil
}

// flushLoop 定期刷新缓冲区
func (bw *BatchWriter) flushLoop() {
	defer close(bw.flushDone)

	for range bw.flushTicker.C {
		bw.mu.Lock()
		if len(bw.buffer) > 0 && !bw.closed {
			if err := bw.flushLocked(); err != nil {
				log.Printf("❌ 批量写入失败: %v", err)
			}
		}
		bw.mu.Unlock()
	}
}

// flushLocked 内部刷新方法（需要已获取锁）
func (bw *BatchWriter) flushLocked() error {
	if len(bw.buffer) == 0 {
		return nil
	}

	records := make([]interface{}, len(bw.buffer))
	copy(records, bw.buffer)
	bw.buffer = bw.buffer[:0] // 清空缓冲区

	// 在锁外执行数据库操作
	go func() {
		if err := bw.db.CreateInBatches(records, bw.batchSize).Error; err != nil {
			log.Printf("❌ 批量写入失败 (%d 条记录): %v", len(records), err)
		} else {
			log.Printf("✅ 批量写入成功 (%d 条记录)", len(records))
		}
	}()

	return nil
}

// Flush 手动刷新缓冲区
func (bw *BatchWriter) Flush() error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	if bw.closed {
		return gorm.ErrInvalidDB
	}

	return bw.flushLocked()
}

// Close 关闭写入器，刷新剩余数据
func (bw *BatchWriter) Close() error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	if bw.closed {
		return nil
	}

	bw.closed = true
	bw.flushTicker.Stop()

	if err := bw.flushLocked(); err != nil {
		return err
	}

	// 等待刷新循环退出
	<-bw.flushDone

	return nil
}

// GameHistoryBatchWriter 专用于游戏历史的批量写入器
type GameHistoryBatchWriter struct {
	bw *BatchWriter
}

// NewGameHistoryBatchWriter 创建游戏历史批量写入器
func NewGameHistoryBatchWriter(db *gorm.DB) *GameHistoryBatchWriter {
	bw := NewBatchWriter(db, database.GameHistory{}, 50, 2*time.Second)
	return &GameHistoryBatchWriter{bw: bw}
}

// AddGameHistory 添加游戏历史记录
func (ghbw *GameHistoryBatchWriter) AddGameHistory(history *database.GameHistory) error {
	return ghbw.bw.Add(history)
}

// AddGameHistories 批量添加游戏历史记录
func (ghbw *GameHistoryBatchWriter) AddGameHistories(histories ...*database.GameHistory) error {
	records := make([]interface{}, len(histories))
	for i, h := range histories {
		records[i] = h
	}
	return ghbw.bw.AddBatch(records...)
}

// Flush 刷新待写入的游戏历史
func (ghbw *GameHistoryBatchWriter) Flush() error {
	return ghbw.bw.Flush()
}

// Close 关闭写入器
func (ghbw *GameHistoryBatchWriter) Close() error {
	return ghbw.bw.Close()
}

// GameHistoryRepository 添加批量写入辅助方法
type GameHistoryBatchOps struct {
	db *gorm.DB
}

// NewGameHistoryBatchOps 创建游戏历史批量操作对象
func NewGameHistoryBatchOps(db *gorm.DB) *GameHistoryBatchOps {
	return &GameHistoryBatchOps{db: db}
}

// CreateBatch 批量创建游戏历史记录
func (ops *GameHistoryBatchOps) CreateBatch(histories []database.GameHistory) error {
	if len(histories) == 0 {
		return nil
	}

	// 按 50 条为一个批次进行插入
	const batchSize = 50
	for i := 0; i < len(histories); i += batchSize {
		end := i + batchSize
		if end > len(histories) {
			end = len(histories)
		}

		batch := histories[i:end]
		if err := ops.db.CreateInBatches(batch, batchSize).Error; err != nil {
			log.Printf("❌ 批量创建游戏历史失败: %v", err)
			return err
		}

		log.Printf("✅ 批量创建游戏历史: %d/%d", end, len(histories))
	}

	return nil
}

// CreateBatchAsync 异步批量创建游戏历史记录（不阻塞）
func (ops *GameHistoryBatchOps) CreateBatchAsync(histories []database.GameHistory) {
	if len(histories) == 0 {
		return
	}

	go func() {
		if err := ops.CreateBatch(histories); err != nil {
			log.Printf("❌ 异步批量创建游戏历史失败: %v", err)
		}
	}()
}

// ChatMessageBatchWriter 专用于聊天消息的批量写入器
type ChatMessageBatchWriter struct {
	bw *BatchWriter
}

// NewChatMessageBatchWriter 创建聊天消息批量写入器
func NewChatMessageBatchWriter(db *gorm.DB) *ChatMessageBatchWriter {
	bw := NewBatchWriter(db, database.GlobalChat{}, 100, 3*time.Second)
	return &ChatMessageBatchWriter{bw: bw}
}

// AddMessage 添加聊天消息
func (cmbw *ChatMessageBatchWriter) AddMessage(msg *database.GlobalChat) error {
	return cmbw.bw.Add(msg)
}

// Flush 刷新待写入的消息
func (cmbw *ChatMessageBatchWriter) Flush() error {
	return cmbw.bw.Flush()
}

// Close 关闭写入器
func (cmbw *ChatMessageBatchWriter) Close() error {
	return cmbw.bw.Close()
}
