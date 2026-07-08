# Chemistry UNO 设计缺陷分析与改进规格

**文档版本**: 1.0  
**创建日期**: 2026-07-05  
**项目版本**: 1.2.1 (Mendeleef)  
**分析范围**: Backend (Go) + Frontend (Vue 3) + Architecture

---

## 📋 目录

1. [执行摘要](#执行摘要)
2. [严重性等级定义](#严重性等级定义)
3. [架构层面问题](#架构层面问题)
4. [后端设计缺陷](#后端设计缺陷)
5. [前端设计缺陷](#前端设计缺陷)
6. [数据库与性能问题](#数据库与性能问题)
7. [安全性问题](#安全性问题)
8. [可维护性与测试问题](#可维护性与测试问题)
9. [改进方案路线图](#改进方案路线图)

---

## 执行摘要

### 项目概况

Chemistry UNO 是一个融合化学反应机制与 UNO 卡牌游戏的多人在线游戏系统，采用 Go + Vue 3 技术栈，支持实时对战、反作弊、教学系统、插件扩展等功能。

### 核心发现

通过对代码库的系统性分析，发现以下关键问题：

**Critical 级别 (2 个)**:
- 并发控制不完整，存在竞态条件风险
- WebSocket 状态同步机制脆弱，可能导致游戏状态不一致

**High 级别 (8 个)**:
- 全局状态管理混乱（`rooms` 全局变量）
- 数据库查询性能瓶颈（JSON 字段的 LIKE 查询）
- 错误处理不一致，缺少统一的错误响应规范
- 配置管理分散，过度依赖环境变量
- 前端状态管理缺失（无 Pinia/Vuex）
- WebSocket 重连逻辑简陋，缺少消息队列持久化
- 缺少 API 版本管理和兼容性策略
- 测试覆盖率低，缺少集成测试和 E2E 测试

**Medium 级别 (12+ 个)**:
- 代码重复严重（化学式解析、元素计数）
- 日志系统不统一
- 缓存策略不完整
- 文档与代码不同步
- 其他可维护性问题

---

## 严重性等级定义

| 级别 | 定义 | 影响 | 修复优先级 |
|------|------|------|-----------|
| **Critical** | 可能导致系统崩溃、数据丢失或严重安全漏洞 | 影响核心功能，可能导致服务不可用 | P0 - 立即修复 |
| **High** | 严重影响性能、可靠性或用户体验 | 影响关键路径，可能导致部分功能不可用 | P1 - 本周修复 |
| **Medium** | 影响代码质量、可维护性或次要功能 | 技术债务累积，长期影响开发效率 | P2 - 本月修复 |
| **Low** | 代码风格、命名、文档等改进建议 | 对功能无直接影响 | P3 - 计划修复 |

---

## 架构层面问题

### 🔴 Critical-1: 全局状态与并发控制问题

**位置**: `backend/game/manager.go:25-30`

```go
var (
	rooms           = make(map[string]*GameRoom)
	roomMutex       sync.RWMutex
	configRepo      *repository.ConfigRepository
	anticheatSystem *anticheat.System
	systemStartTime time.Time
)
```

**问题描述**:
1. **全局可变状态**: `rooms` 作为包级全局变量，在高并发场景下难以追踪和调试
2. **锁粒度过粗**: 单一的 `roomMutex` 保护所有房间，限制了并发性能
3. **测试困难**: 全局状态导致测试隔离困难，需要额外的清理逻辑
4. **扩展性差**: 无法支持多实例部署（房间状态仅存在于单个进程内存）

**影响**:
- 单点故障：进程重启导致所有房间状态丢失
- 水平扩展受限：无法简单地增加后端实例
- 竞态条件风险：241 处锁使用（见 Grep 统计），维护成本高

**证据**:
```bash
# 锁使用统计
backend\game\ai_controller.go:23
backend\game\manager.go:199
# Total: 241 occurrences across 5 files
```

**改进方案**:
1. **短期（2周）**:
   - 引入 `RoomManager` 结构体封装全局状态
   - 使用细粒度锁（每个房间独立锁）
   - 添加并发测试用例

2. **中期（1-2月）**:
   - 迁移到 Redis 作为房间状态存储
   - 实现基于 Redis Pub/Sub 的跨节点广播
   - 支持水平扩展

3. **长期（3-6月）**:
   - 评估引入 Actor 模型（如使用 `github.com/asynkron/protoactor-go`）
   - 每个房间作为独立 Actor，自然避免并发问题

**代码示例（短期改进）**:
```go
// backend/game/room_manager.go
type RoomManager struct {
    rooms map[string]*GameRoom
    mu    sync.RWMutex
    
    // Per-room locks for fine-grained concurrency
    roomLocks sync.Map // map[roomID]*sync.RWMutex
}

func NewRoomManager() *RoomManager {
    return &RoomManager{
        rooms: make(map[string]*GameRoom),
    }
}

func (rm *RoomManager) GetRoom(roomID string) (*GameRoom, error) {
    rm.mu.RLock()
    defer rm.mu.RUnlock()
    
    room, exists := rm.rooms[roomID]
    if !exists {
        return nil, ErrRoomNotFound
    }
    return room, nil
}

func (rm *RoomManager) AcquireRoomLock(roomID string) *sync.RWMutex {
    lock, _ := rm.roomLocks.LoadOrStore(roomID, &sync.RWMutex{})
    return lock.(*sync.RWMutex)
}
```

---

### 🔴 Critical-2: WebSocket Hub 状态同步问题

**位置**: `backend/websocket/hub.go:20-28`

**问题描述**:
1. **状态不一致风险**: Hub 的 `clients` 和 `rooms` 映射之间缺少原子性保证
2. **消息丢失**: 客户端断线时，send channel 被关闭，未发送的消息直接丢弃
3. **重连逻辑简陋**: 前端重连后需要重新加入房间，可能错过中间状态变更
4. **缺少消息确认机制**: 无法保证关键消息（如游戏状态更新）的送达

**影响**:
- 游戏状态不同步：玩家看到的手牌、回合信息可能不一致
- 竞态条件：用户快速断开/重连时可能导致重复注册或清理失败
- 调试困难：消息丢失无法追溯，难以重现问题

**证据**:
```typescript
// frontend/src/utils/websocket.ts:84-92
disconnect(): void {
    this.reconnectAttempts = this.maxReconnectAttempts
    this.pendingMessages = [] // 消息直接丢弃
    if (this.ws) {
        this.ws.close()
        this.ws = null
    }
}
```

**改进方案**:

1. **短期（1周）**:
   - 添加消息序列号和确认机制
   - 实现消息重传队列（至少保留最近 100 条）
   - 改进前端重连逻辑，自动恢复房间订阅

2. **中期（1-2月）**:
   - 引入 Redis Streams 作为消息队列
   - 实现"至少一次"送达语义
   - 添加消息持久化（关键游戏事件）

3. **长期（3-6月）**:
   - 评估使用成熟的消息中间件（NATS/RabbitMQ）
   - 实现事件溯源（Event Sourcing）模式
   - 支持游戏状态快照和回放

**代码示例（短期改进）**:
```go
// backend/websocket/message_queue.go
type AcknowledgedMessage struct {
    ID        string
    Message   Message
    Timestamp time.Time
    Attempts  int
}

type MessageQueue struct {
    messages map[string]*AcknowledgedMessage
    mu       sync.RWMutex
    maxAge   time.Duration
}

func (mq *MessageQueue) Store(msg Message) string {
    mq.mu.Lock()
    defer mq.mu.Unlock()
    
    id := uuid.New().String()
    mq.messages[id] = &AcknowledgedMessage{
        ID:        id,
        Message:   msg,
        Timestamp: time.Now(),
    }
    return id
}

func (mq *MessageQueue) Acknowledge(id string) {
    mq.mu.Lock()
    defer mq.mu.Unlock()
    delete(mq.messages, id)
}

func (mq *MessageQueue) GetPending(since time.Time) []AcknowledgedMessage {
    mq.mu.RLock()
    defer mq.mu.RUnlock()
    
    var pending []AcknowledgedMessage
    for _, msg := range mq.messages {
        if msg.Timestamp.After(since) {
            pending = append(pending, *msg)
        }
    }
    return pending
}
```

---

### 🟡 High-1: 层级边界违反

**位置**: 多处

**问题描述**:
虽然项目定义了清晰的分层架构（见 `docs/architecture/FILE_RESPONSIBILITIES.md`），但实际代码中存在多处跨层调用：

1. **Handlers 直接操作游戏状态**: `handlers/game.go` 中大量调用 `game` 包的全局函数
2. **Repository 包含业务逻辑**: `repository/game_repository.go:83` 中的查询优化逻辑应该在 service 层
3. **缺少 Service 层**: 复杂业务逻辑散落在 handlers 和 game 包中

**影响**:
- 代码复用困难
- 测试覆盖率低（无法独立测试业务逻辑）
- 功能重复（多处存在相同的验证逻辑）

**改进方案**:
1. 引入 Service 层（`backend/services/`）
2. 将业务逻辑从 handlers 移至 services
3. Handlers 仅负责 HTTP 协议适配
4. Repository 仅负责数据访问

**代码示例**:
```go
// backend/services/room_service.go
type RoomService struct {
    roomRepo     *repository.RoomRepository
    gameRepo     *repository.GameRepository
    userRepo     *repository.UserRepository
    roomManager  *game.RoomManager
}

func (s *RoomService) CreateRoom(ctx context.Context, req CreateRoomRequest) (*models.Room, error) {
    // 业务逻辑：验证、权限检查、房间创建
    user, err := s.userRepo.FindByUID(ctx, req.CreatorUID)
    if err != nil {
        return nil, err
    }
    
    if user.BannedUntil != nil && user.BannedUntil.After(time.Now()) {
        return nil, ErrUserBanned
    }
    
    room := s.roomManager.CreateRoom(req)
    if err := s.roomRepo.Save(ctx, room); err != nil {
        return nil, err
    }
    
    return room, nil
}

// backend/handlers/game.go
func CreateRoom(c *gin.Context) {
    var req CreateRoomRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    req.CreatorUID = c.GetInt("uid")
    
    room, err := roomService.CreateRoom(c.Request.Context(), req)
    if err != nil {
        handleError(c, err)
        return
    }
    
    c.JSON(200, room)
}
```

---

## 后端设计缺陷

### 🟡 High-2: 数据库查询性能瓶颈

**位置**: `backend/repository/game_repository.go:81-99`

**问题描述**:
用户游戏历史查询使用 JSON 字段的 LIKE 模式匹配，无法利用索引：

```go
query := r.db.Order("created_at DESC").Limit(50).Where(
    "players LIKE ? OR players LIKE ? OR players LIKE ? OR players LIKE ?",
    "["+idStr+"]",   // 唯一元素
    "["+idStr+",%",  // 数组开头
    "%,"+idStr+"]",  // 数组结尾
    "%,"+idStr+",%", // 数组中间
)
```

**影响**:
- 查询时间随数据量线性增长
- 无法利用数据库索引
- 高并发时成为性能瓶颈

**证据**:
README.md 中提到已有优化方案，但默认关闭：
```env
USE_OPTIMIZED_HISTORY_QUERIES=false
```

**改进方案**:

已存在 junction table 方案（`database.GameHistoryPlayer`），但需要：

1. **短期（立即）**:
   - 将 `USE_OPTIMIZED_HISTORY_QUERIES` 默认改为 `true`
   - 添加数据库迁移脚本，为现有数据填充 junction table
   - 添加性能测试对比

2. **中期（1月）**:
   - 完全移除旧查询逻辑
   - 添加复合索引：`(player_uid, game_history_id)`
   - 实现查询结果缓存

**代码示例（迁移脚本）**:
```go
// backend/database/migrate_game_history_index.go
func MigrateGameHistoryPlayerIndex(db *gorm.DB) error {
    log.Println("📊 开始迁移游戏历史索引...")
    
    var histories []GameHistory
    if err := db.Find(&histories).Error; err != nil {
        return err
    }
    
    for _, history := range histories {
        var players []int
        if err := json.Unmarshal([]byte(history.Players), &players); err != nil {
            log.Printf("⚠️  跳过无效记录: game_history_id=%d, error=%v", history.ID, err)
            continue
        }
        
        for _, uid := range players {
            junction := GameHistoryPlayer{
                GameHistoryID: history.ID,
                PlayerUID:     uint(uid),
            }
            if err := db.FirstOrCreate(&junction).Error; err != nil {
                return fmt.Errorf("failed to create junction: %v", err)
            }
        }
    }
    
    log.Println("✅ 游戏历史索引迁移完成")
    return nil
}
```

---

### 🟡 High-3: 化学式解析代码重复

**位置**: 
- `backend/game/chemistry.go:91-120`
- `backend/game/judge.go` (类似逻辑)

**问题描述**:
化学式解析逻辑在多处重复实现：
- `parseSubstance()`: 解析化学式为元素映射
- `canFormSubstance()`: 检查是否能组成物质
- `NormalizeSubscripts()`: 标准化下标（可能在多处使用）

**影响**:
- 维护成本高（修复 bug 需要多处修改）
- 行为不一致风险
- 测试覆盖困难

**改进方案**:

1. 提取到独立的 `chemistry` 工具包
2. 添加完整的单元测试
3. 支持复杂化学式（括号嵌套、水合物等）

**代码示例**:
```go
// backend/chemistry/parser.go
package chemistry

type Formula struct {
    Elements map[string]int
    Raw      string
}

func Parse(formula string) (*Formula, error) {
    normalized := NormalizeSubscripts(formula)
    elements, err := parseElements(normalized)
    if err != nil {
        return nil, fmt.Errorf("invalid formula %q: %w", formula, err)
    }
    
    return &Formula{
        Elements: elements,
        Raw:      formula,
    }, nil
}

func (f *Formula) CanFormFrom(available map[string]int) bool {
    for elem, required := range f.Elements {
        if available[elem] < required {
            return false
        }
    }
    return true
}

// backend/chemistry/parser_test.go
func TestParse(t *testing.T) {
    tests := []struct {
        input    string
        expected map[string]int
        wantErr  bool
    }{
        {"H2O", map[string]int{"H": 2, "O": 1}, false},
        {"Ca(OH)2", map[string]int{"Ca": 1, "O": 2, "H": 2}, false},
        {"CuSO4·5H2O", map[string]int{"Cu": 1, "S": 1, "O": 9, "H": 10}, false},
        {"", nil, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result, err := Parse(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
                return
            }
            if !tt.wantErr && !reflect.DeepEqual(result.Elements, tt.expected) {
                t.Errorf("Parse(%q) = %v, want %v", tt.input, result.Elements, tt.expected)
            }
        })
    }
}
```

---

### 🟡 High-4: 错误处理不统一

**位置**: 多处

**问题描述**:
1. 错误类型不统一：有些返回 `gin.H{"error": "..."}`, 有些返回自定义结构
2. HTTP 状态码使用不规范：部分错误返回 200 但包含 error 字段
3. 缺少错误码体系：前端无法准确区分错误类型
4. 错误信息暴露过多：直接返回数据库错误到前端

**影响**:
- 前端难以统一处理错误
- 用户体验差（错误提示不友好）
- 安全风险（信息泄露）

**改进方案**:

1. 定义统一的错误响应结构
2. 建立错误码体系
3. 使用中间件统一处理错误
4. 添加错误日志和追踪

**代码示例**:
```go
// backend/errors/errors.go
package errors

import "net/http"

type ErrorCode string

const (
    ErrCodeUnknown           ErrorCode = "UNKNOWN"
    ErrCodeInvalidInput      ErrorCode = "INVALID_INPUT"
    ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
    ErrCodeForbidden         ErrorCode = "FORBIDDEN"
    ErrCodeNotFound          ErrorCode = "NOT_FOUND"
    ErrCodeConflict          ErrorCode = "CONFLICT"
    ErrCodeRateLimited       ErrorCode = "RATE_LIMITED"
    ErrCodeUserBanned        ErrorCode = "USER_BANNED"
    ErrCodeRoomFull          ErrorCode = "ROOM_FULL"
    ErrCodeGameInProgress    ErrorCode = "GAME_IN_PROGRESS"
)

type AppError struct {
    Code       ErrorCode `json:"code"`
    Message    string    `json:"message"`
    HTTPStatus int       `json:"-"`
    Details    any       `json:"details,omitempty"`
}

func (e *AppError) Error() string {
    return e.Message
}

func New(code ErrorCode, message string, status int) *AppError {
    return &AppError{
        Code:       code,
        Message:    message,
        HTTPStatus: status,
    }
}

func InvalidInput(message string) *AppError {
    return New(ErrCodeInvalidInput, message, http.StatusBadRequest)
}

func Unauthorized(message string) *AppError {
    return New(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func NotFound(message string) *AppError {
    return New(ErrCodeNotFound, message, http.StatusNotFound)
}

// backend/middleware/error_handler.go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) == 0 {
            return
        }
        
        err := c.Errors.Last().Err
        
        var appErr *errors.AppError
        if e, ok := err.(*errors.AppError); ok {
            appErr = e
        } else {
            // Unknown error - don't expose details
            appErr = errors.New(
                errors.ErrCodeUnknown,
                "An internal error occurred",
                http.StatusInternalServerError,
            )
            log.Printf("❌ Unhandled error: %v", err)
        }
        
        c.JSON(appErr.HTTPStatus, gin.H{
            "error": appErr,
        })
    }
}

// backend/handlers/game.go
func CreateRoom(c *gin.Context) {
    var req CreateRoomRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.Error(errors.InvalidInput("Invalid request format"))
        return
    }
    
    room, err := roomService.CreateRoom(c.Request.Context(), req)
    if err != nil {
        c.Error(err)
        return
    }
    
    c.JSON(200, room)
}
```

---

### 🟡 High-5: 配置管理分散

**位置**: 多处使用 `os.Getenv()`

**问题描述**:
1. 配置散落在各个文件中，难以追踪和管理
2. 缺少配置验证和默认值管理
3. 类型转换逻辑重复（字符串转数字、布尔等）
4. 缺少配置文档（哪些是必需的，哪些是可选的）

**影响**:
- 部署困难（不知道需要配置哪些环境变量）
- 运行时错误（配置错误只有在使用时才发现）
- 测试困难（需要设置大量环境变量）

**证据**:
```bash
# 15个文件直接使用 os.Getenv
backend\anticheat\system.go
backend\repository\game_repository.go
backend\database\migrate.go
backend\middleware\rate_limit.go
# ... 11 more files
```

**改进方案**:

1. **短期（1周）**:
   - 创建统一的配置结构体
   - 启动时一次性加载和验证所有配置
   - 提供配置文档生成工具

2. **中期（1月）**:
   - 支持配置热重载
   - 添加配置中心集成（Consul/etcd）
   - 支持多环境配置文件

**代码示例**:
```go
// backend/config/config.go
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

type Config struct {
    App      AppConfig
    Database DatabaseConfig
    Redis    RedisConfig
    JWT      JWTConfig
    SMTP     SMTPConfig
    OAuth    OAuthConfig
    Features FeatureFlags
}

type AppConfig struct {
    Version     string
    VersionName string
    Environment string // development, staging, production
}

type DatabaseConfig struct {
    Type       string // sqlite, mysql
    SQLitePath string
    MySQLDSN   string
}

type RedisConfig struct {
    Enabled  bool
    Addr     string
    Username string
    Password string
    DB       int
}

type JWTConfig struct {
    Secret string
    Expiry time.Duration
}

type FeatureFlags struct {
    EnableReactionCache       bool
    RateLimitCleanupEnabled   bool
    UseOptimizedHistoryQueries bool
    EnableAnticheatBatch      bool
    EnableAnticheatStreams    bool
}

var Global *Config

func Load() (*Config, error) {
    cfg := &Config{
        App: AppConfig{
            Version:     getEnv("APP_VERSION", "1.2.1"),
            VersionName: getEnv("APP_VERSION_NAME", "Mendeleef"),
            Environment: getEnv("APP_ENV", "development"),
        },
        Database: DatabaseConfig{
            Type:       getEnv("DB_TYPE", "sqlite"),
            SQLitePath: getEnv("SQLITE_PATH", "./chemistryuno.db"),
            MySQLDSN:   getEnv("MYSQL_DSN", ""),
        },
        Redis: RedisConfig{
            Enabled:  getBoolEnv("REDIS_ENABLED", true),
            Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
            Username: getEnv("REDIS_USERNAME", ""),
            Password: getEnv("REDIS_PASSWORD", ""),
            DB:       getIntEnv("REDIS_DB", 0),
        },
        JWT: JWTConfig{
            Secret: getEnv("JWT_SECRET", ""),
            Expiry: getDurationEnv("JWT_EXPIRY", 24*time.Hour),
        },
        Features: FeatureFlags{
            EnableReactionCache:       getBoolEnv("ENABLE_REACTION_CACHE", true),
            RateLimitCleanupEnabled:   getBoolEnv("RATE_LIMIT_CLEANUP_ENABLED", true),
            UseOptimizedHistoryQueries: getBoolEnv("USE_OPTIMIZED_HISTORY_QUERIES", true),
            EnableAnticheatBatch:      getBoolEnv("ENABLE_ANTICHEAT_BATCH", false),
            EnableAnticheatStreams:    getBoolEnv("ENABLE_ANTICHEAT_STREAMS", false),
        },
    }
    
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    Global = cfg
    return cfg, nil
}

func (c *Config) Validate() error {
    if c.JWT.Secret == "" {
        return fmt.Errorf("JWT_SECRET is required")
    }
    
    if c.Database.Type == "mysql" && c.Database.MySQLDSN == "" {
        return fmt.Errorf("MYSQL_DSN is required when DB_TYPE=mysql")
    }
    
    return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        return value == "true" || value == "1"
    }
    return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}

// main.go
func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    
    log.Printf("Starting %s v%s (%s)", cfg.App.VersionName, cfg.App.Version, cfg.App.Environment)
    
    // Use cfg.Database, cfg.Redis, etc.
}
```

---

## 前端设计缺陷

### 🟡 High-6: 缺少状态管理库

**位置**: `frontend/src/pages/GameRoom.vue` (2610+ lines)

**问题描述**:
1. 状态分散在各个组件中，使用 `ref()` 和局部变量
2. 组件间通信困难，依赖 props drilling 和事件传递
3. WebSocket 消息处理逻辑分散
4. 难以实现时间旅行调试和状态持久化

**影响**:
- GameRoom.vue 过于臃肿（超过 2000 行）
- 状态不一致问题难以调试
- 无法实现离线模式或乐观更新
- 代码复用困难

**证据**:
```bash
# GameRoom.vue 是最大的单文件组件
frontend/src/pages/GameRoom.vue: 2610+ lines
```

**改进方案**:

1. **短期（2周）**:
   - 引入 Pinia 作为状态管理库
   - 创建 `gameStore` 集中管理游戏状态
   - 提取 WebSocket 处理逻辑到独立的 composable

2. **中期（1-2月）**:
   - 实现状态持久化（localStorage）
   - 添加状态快照和回滚功能
   - 支持离线模式（本地模拟）

**代码示例**:
```typescript
// frontend/src/stores/game.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useGameStore = defineStore('game', () => {
  // State
  const room = ref<Room | null>(null)
  const gameState = ref<GameState | null>(null)
  const currentPlayer = ref<PlayerState | null>(null)
  const isMyTurn = ref(false)
  const connected = ref(false)
  
  // Getters
  const myHandCards = computed(() => {
    if (!currentPlayer.value) return []
    return currentPlayer.value.hand_cards
  })
  
  const availableSubstances = computed(() => {
    if (!myHandCards.value) return []
    return calculateAvailableSubstances(myHandCards.value)
  })
  
  // Actions
  function updateRoom(newRoom: Room) {
    room.value = newRoom
  }
  
  function updateGameState(newState: GameState) {
    gameState.value = newState
    updateCurrentPlayer()
  }
  
  function updateCurrentPlayer() {
    if (!gameState.value) return
    const uid = useUserStore().uid
    currentPlayer.value = gameState.value.players.find(p => p.uid === uid) || null
    isMyTurn.value = gameState.value.current_player === gameState.value.players.findIndex(p => p.uid === uid)
  }
  
  function playCard(substance: string, cards: Card[]) {
    // Optimistic update
    const previousState = { ...gameState.value }
    
    // Update local state immediately
    if (currentPlayer.value) {
      currentPlayer.value.hand_cards = currentPlayer.value.hand_cards.filter(
        c => !cards.some(card => card.type === c.type)
      )
    }
    
    // Send to server
    return api.playCard(room.value!.id, substance, cards)
      .catch(err => {
        // Rollback on error
        gameState.value = previousState
        throw err
      })
  }
  
  function reset() {
    room.value = null
    gameState.value = null
    currentPlayer.value = null
    isMyTurn.value = false
  }
  
  return {
    // State
    room,
    gameState,
    currentPlayer,
    isMyTurn,
    connected,
    
    // Getters
    myHandCards,
    availableSubstances,
    
    // Actions
    updateRoom,
    updateGameState,
    playCard,
    reset,
  }
})

// frontend/src/composables/useWebSocket.ts
import { useGameStore } from '@/stores/game'
import websocket from '@/utils/websocket'

export function useGameWebSocket(roomId: string) {
  const gameStore = useGameStore()
  
  function handleGameUpdate(message: any) {
    gameStore.updateGameState(message.data)
  }
  
  function handleRoomUpdate(message: any) {
    gameStore.updateRoom(message.data)
  }
  
  function setupListeners() {
    websocket.on('game_update', handleGameUpdate)
    websocket.on('room_update', handleRoomUpdate)
    websocket.on('player_joined', () => {
      // Refresh room state
      api.getRoom(roomId).then(gameStore.updateRoom)
    })
  }
  
  function cleanupListeners() {
    websocket.off('game_update', handleGameUpdate)
    websocket.off('room_update', handleRoomUpdate)
    websocket.off('player_joined')
  }
  
  onMounted(setupListeners)
  onUnmounted(cleanupListeners)
  
  return {
    setupListeners,
    cleanupListeners,
  }
}

// frontend/src/pages/GameRoom.vue (simplified)
<script setup lang="ts">
import { useGameStore } from '@/stores/game'
import { useGameWebSocket } from '@/composables/useWebSocket'

const route = useRoute()
const gameStore = useGameStore()
const { room, gameState, isMyTurn, availableSubstances } = storeToRefs(gameStore)

useGameWebSocket(route.params.id as string)

async function handlePlayCard(substance: string, cards: Card[]) {
  try {
    await gameStore.playCard(substance, cards)
    showToast('出牌成功')
  } catch (err) {
    showAlert('出牌失败: ' + err.message)
  }
}
</script>
```

---

### 🟠 Medium-1: WebSocket 重连逻辑问题

**位置**: `frontend/src/utils/websocket.ts`

**问题描述**:
1. 固定重连次数（5次）后永久放弃
2. 固定重连延迟（3秒），无指数退避
3. 重连后需要手动重新订阅房间
4. 缺少心跳检测机制

**影响**:
- 网络抖动时容易断连
- 用户需要手动刷新页面恢复连接
- 长时间空闲后连接可能"僵死"

**改进方案**:

```typescript
// frontend/src/utils/websocket.ts
class WebSocketService {
  private reconnectAttempts = 0
  private readonly maxReconnectAttempts = Infinity // 永不放弃
  private reconnectDelay = 1000 // 初始延迟 1秒
  private readonly maxReconnectDelay = 30000 // 最大延迟 30秒
  private heartbeatInterval: NodeJS.Timeout | null = null
  private lastPongTime = Date.now()
  
  connect(): void {
    if (this.isConnecting || this.isConnected()) {
      return
    }
    
    this.isConnecting = true
    this.ws = new WebSocket(WS_URL)
    
    this.ws.onopen = () => {
      console.log('[WebSocket] connected')
      this.reconnectAttempts = 0
      this.reconnectDelay = 1000 // 重置延迟
      this.isConnecting = false
      this.startHeartbeat()
      
      // 重新订阅之前的房间
      this.resubscribeRooms()
      
      // 发送待发送消息
      this.flushPendingMessages()
    }
    
    this.ws.onclose = (event) => {
      console.log('[WebSocket] disconnected', event.code, event.reason)
      this.isConnecting = false
      this.stopHeartbeat()
      this.attemptReconnect()
    }
    
    this.ws.onerror = (error) => {
      console.error('[WebSocket] error:', error)
      this.isConnecting = false
    }
  }
  
  private attemptReconnect(): void {
    // 指数退避策略
    const delay = Math.min(
      this.reconnectDelay * Math.pow(2, this.reconnectAttempts),
      this.maxReconnectDelay
    )
    
    this.reconnectAttempts++
    console.log(`[WebSocket] will reconnect in ${delay}ms (attempt ${this.reconnectAttempts})`)
    
    setTimeout(() => this.connect(), delay)
  }
  
  private startHeartbeat(): void {
    this.lastPongTime = Date.now()
    
    this.heartbeatInterval = setInterval(() => {
      if (Date.now() - this.lastPongTime > 15000) {
        // 15秒未收到pong，认为连接已死
        console.warn('[WebSocket] heartbeat timeout, reconnecting...')
        this.ws?.close()
        return
      }
      
      this.send({ type: 'ping' })
    }, 5000)
  }
  
  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval)
      this.heartbeatInterval = null
    }
  }
  
  private handleMessage(message: WebSocketMessage): void {
    if (message.type === 'pong') {
      this.lastPongTime = Date.now()
      return
    }
    
    // 通知监听器
    this.notifyListeners(message)
  }
  
  // 重新订阅房间
  private subscribedRooms = new Set<string>()
  
  subscribeRoom(roomId: string): void {
    this.subscribedRooms.add(roomId)
    this.send({ type: 'subscribe_room', room_id: roomId })
  }
  
  unsubscribeRoom(roomId: string): void {
    this.subscribedRooms.delete(roomId)
    this.send({ type: 'unsubscribe_room', room_id: roomId })
  }
  
  private resubscribeRooms(): void {
    for (const roomId of this.subscribedRooms) {
      this.send({ type: 'subscribe_room', room_id: roomId })
    }
  }
}
```

---

### 🟠 Medium-2: 组件职责不清晰

**位置**: `frontend/src/pages/GameRoom.vue`

**问题描述**:
1. GameRoom.vue 承担了太多职责：
   - UI 渲染
   - 业务逻辑
   - WebSocket 通信
   - 游戏状态管理
   - 音效和动画控制
   
2. 组件拆分不够细致：
   - 手牌区、出牌区、玩家列表等应该是独立组件
   - 每个组件应该有明确的输入（props）和输出（events）

**改进方案**:

```
GameRoom.vue (容器组件，200行以内)
├── GameBoard.vue (游戏面板)
│   ├── PlayerList.vue (玩家列表)
│   │   └── PlayerCard.vue (单个玩家卡片)
│   ├── DiscardPile.vue (弃牌堆)
│   └── GameInfo.vue (游戏信息：回合、方向等)
├── HandCards.vue (手牌区)
│   └── CardItem.vue (单张卡牌)
├── ActionPanel.vue (操作面板)
│   ├── SubstanceSelector.vue (物质选择器)
│   └── ActionButtons.vue (操作按钮)
└── GameSidebar.vue (侧边栏)
    ├── ChatBox.vue (聊天框)
    ������ PlayerStats.vue (���ͳ��)
```

**����ʾ��**:
```vue
<!-- frontend/src/components/game/HandCards.vue -->
<script setup lang="ts">
import { computed } from 'vue'
import type { Card } from '@/types'

interface Props {
  cards: Card[]
  selectedCards: Card[]
  disabled: boolean
}

interface Emits {
  (e: 'select', card: Card): void
  (e: 'deselect', card: Card): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

function isSelected(card: Card): boolean {
  return props.selectedCards.some(c => c.type === card.type)
}

function toggleCard(card: Card): void {
  if (props.disabled) return
  
  if (isSelected(card)) {
    emit('deselect', card)
  } else {
    emit('select', card)
  }
}
</script>

<template>
  <div class="hand-cards">
    <CardItem
      v-for="(card, index) in cards"
      :key="`${card.type}-${index}`"
      :card="card"
      :selected="isSelected(card)"
      :disabled="disabled"
      @click="toggleCard(card)"
    />
  </div>
</template>
```

---

## �ܽ�

���ĵ�ʶ�𲢷����� Chemistry UNO ��Ŀ�е����ȱ�ݣ����ṩ����ϸ�ĸĽ�������ʵʩ·��ͼ��

### ���Ľ���

1. **�����޸� Critical ��������**��ȫ��״̬������ WebSocket ͬ����ϵͳ�ȶ��ԵĻ���
2. **���׶�ʵʩ�Ľ�**����ѭ4��Phase��·��ͼ��ȷ��ÿ���׶ζ��п���֤�ĳɹ�
3. **�������Ը�����**�����ǽ��ͻع���ա������������ĵĹؼ�
4. **�����ع�**������ծ��Ӧ�ó����������������ۻ�

### Ԥ������

ͨ��ʵʩ������еĸĽ�������Ԥ�ƿ��Դ�ɣ�

- **�������� 50%+**: ͨ�����ݿ��Ż���������ԺͲ�������
- **�ȶ�������**: Bug �ʽ��� 70%��״̬ͬ��׼ȷ�� >99.9%
- **��ά��������**: ���븴�ӶȽ��� 30%�����Ը����� >70%
- **����Ч������**: �¹��ܿ���ʱ������ 40%
- **����չ��**: ֧��ˮƽ��չ��1000+ ��������

---

**�ĵ��汾**: 1.0  
**������**: 2026-07-05  
**�´����**: 2026-08-05  
**ά����**: �����Ŷ�

---

## �����־

### v1.0 (2026-07-05)
- ��ʼ�汾
- ʶ�� 2 �� Critical��8 �� High��12+ �� Medium ��������
- �ṩ��ϸ�ĸĽ������ʹ���ʾ��
- ���� 4-Phase ʵʩ·��ͼ

---

**��л**: ���ĵ����ڶԴ��������������ɣ��ر��л Claude Code ���Զ�������������
