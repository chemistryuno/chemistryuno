# 反作弊模块 (Anticheat Module)

该模块提供了完整的游戏作弊检测和防护功能，包括快速反应检测、可疑活动监控、作弊举报管理等。

## 主要组件

### 1. FastReactionChecker（快速反应检查器）

检测和记录玩家的快速反应行为。

**功能：**
- 记录玩家的快速反应动作
- 统计每个玩家的快速反应次数
- 检测哪些玩家有快速反应异常

**使用示例：**
```go
checker := anticheat.NewFastReactionChecker()

// 记录快速反应
checker.RecordFastReaction(userID, isFastReaction)

// 检测作弊
result := checker.DetectCheat()
if result.CheatDetected {
    log.Printf("检测到作弊用户: %v", result.CheatUIDs)
}
```

### 2. SnapshotBuilder（回放快照构建器）

构建游戏回放快照，用于保存游戏过程和作弊检测信息。

**功能：**
- 链式API构建复杂的回放快照
- 自动序列化为JSON
- 包含游戏过程中的所有作弊检测数据

**使用示例：**
```go
snapshot, err := anticheat.NewSnapshotBuilder(roomID).
    WithParticipants(participantsData).
    WithEvents(eventList).
    WithCheatDetection(cheatResult).
    WithReason("game_finished").
    WithStartedAt(gameStartTime).
    WithGameStatus("finished", finishedPlayers, playerCount, quittedCount).
    Build()
```

### 3. SuspiciousActivityDetector（可疑活动检测器）

监控玩家的实时活动，检测异常频繁的操作。

**功能：**
- 记录玩家操作时间
- 基于时间窗口检测超高频操作
- 获取所有可疑用户列表
- 自动清除过期的日志

**使用示例：**
```go
// 创建检测器：10秒窗口内最多5个操作
detector := anticheat.NewSuspiciousActivityDetector(10*time.Second, 5)

// 记录用户操作
detector.RecordAction(userID)

// 检查用户是否可疑
if detector.IsSuspicious(userID) {
    log.Printf("用户 %d 有可疑活动", userID)
}

// 获取所有可疑用户
suspiciousUsers := detector.GetSuspiciousUsers()

// 定期清理过期日志（建议在游戏循环中调用）
detector.ClearOldLogs()
```

### 4. CheatReportManager（作弊举报管理器）

管理玩家的作弊举报，支持举报提交、查询和状态管理。

**功能：**
- 提交作弊举报
- 查询特定用户的举报
- 更新举报状态
- 追踪调查进度

**使用示例：**
```go
manager := anticheat.NewCheatReportManager()

// 提交举报
report := &anticheat.CheatReport{
    RoomID:      roomID,
    ReportedUID: suspectUID,
    ReporterUID: reporterUID,
    Reason:      "Fast reaction",
    Evidence:    "...",
}
manager.SubmitReport(report)

// 查询某个用户的举报
reports := manager.GetReportsByUID(userID)

// 更新举报状态
manager.UpdateReportStatus(reportID, "confirmed")
```

## 数据结构

### CheatDetectionResult
```go
type CheatDetectionResult struct {
    CheatDetected bool  // 是否检测到作弊
    CheatUIDs     []int // 作弊用户ID列表
}
```

### ReplaySnapshot
```go
type ReplaySnapshot struct {
    Version              int                   // 版本号
    RoomID               int64                 // 房间ID
    GeneratedAt          string                // 生成时间
    Participants         json.RawMessage       // 参与者信息
    Events               []map[string]interface{} // 游戏事件列表
    CheatDetected        bool                  // 是否检测到作弊
    CheatUIDs            []int                 // 作弊用户ID
    Reason               string                // 生成原因
    StartedAt            string                // 游戏开始时间
    Status               string                // 游戏状态
    FinishedPlayers      []int                 // 完成的玩家
    OriginalPlayerCount  int                   // 原始玩家数
    QuittedCount         int                   // 退出的玩家数
}
```

### CheatReport
```go
type CheatReport struct {
    RoomID      int64     // 房间ID
    ReportedUID int       // 被举报用户ID
    ReporterUID int       // 举报者ID
    Reason      string    // 举报原因
    Evidence    string    // 举报证据
    CreatedAt   time.Time // 创建时间
    Status      string    // 状态: pending/investigating/confirmed/dismissed
}
```

## 集成指南

在游戏管理器中集成反作弊模块：

1. **初始化**
```go
import "chemistryuno/backend/anticheat"

type GameRoom struct {
    // ... 现有字段
    FastReactionChecker *anticheat.FastReactionChecker
    ActivityDetector    *anticheat.SuspiciousActivityDetector
    ReportManager       *anticheat.CheatReportManager
}

func NewGameRoom() *GameRoom {
    gr := &GameRoom{
        FastReactionChecker: anticheat.NewFastReactionChecker(),
        ActivityDetector:    anticheat.NewSuspiciousActivityDetector(10*time.Second, 20),
        ReportManager:       anticheat.NewCheatReportManager(),
    }
    return gr
}
```

2. **记录玩家动作**
```go
func (gr *GameRoom) PlayerMove(uid int, isFast bool) {
    gr.FastReactionChecker.RecordFastReaction(uid, isFast)
    gr.ActivityDetector.RecordAction(uid)
    
    if gr.ActivityDetector.IsSuspicious(uid) {
        log.Printf("警告: 用户 %d 活动频繁，可能作弊", uid)
    }
}
```

3. **生成回放快照**
```go
func (gr *GameRoom) GenerateReplaySnapshot() (string, error) {
    cheatResult := gr.FastReactionChecker.DetectCheat()
    
    return anticheat.NewSnapshotBuilder(gr.Room.ID).
        WithParticipants(participantData).
        WithEvents(gr.ReplayEvents).
        WithCheatDetection(cheatResult).
        WithReason("game_finished").
        WithStartedAt(gr.GameStartedAt).
        WithGameStatus(gr.GameState.Status, gr.GameState.FinishedPlayers, 
                      gr.GameState.OriginalPlayerCount, gr.GameState.QuittedCount).
        Build()
}
```

## 测试

运行单元测试：
```bash
go test -v ./backend/anticheat/
```

## 性能考虑

- **内存管理**：`SuspiciousActivityDetector` 会持续记录用户活动，建议定期调用 `ClearOldLogs()` 清理过期数据
- **时间复杂度**：
  - `RecordFastReaction`：O(1)
  - `DetectCheat`：O(n log n)（其中n为有反应的用户数）
  - `IsSuspicious`：O(n)（其中n为某用户的操作数）

## 安全建议

1. 始终在服务端进行作弊检测，不要依赖客户端数据
2. 保存完整的回放日志以供人工审核
3. 定期分析作弊数据，调整检测参数
4. 对确认的作弊行为实施适当的惩罚机制
5. 为管理员提供完整的审计日志
