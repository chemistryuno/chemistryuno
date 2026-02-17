# 等级系统实现文档

## 系统概览

已经完成了 Chemistry UNO 的等级系统核心功能，包括：
- ✅ 数据库结构设计
- ✅ 经验计算逻辑
- ✅ 自动升级系统
- ✅ API 接口
- ✅ 游戏结算XP奖励
- ⏳ 匹配机制（待实现）
- ⏳ 前端显示（待实现）

## 等级系统设计

### 1. 等级范围与段位
- **等级**: 1-100级
- **段位系统**:
  - 🥉 青铜 (1-10级)
  - 🥈 白银 (11-25级)
  - 🥇 黄金 (26-45级)
  - 💎 铂金 (46-70级)
  - 💠 钻石 (71-90级)
  - ⭐ 大师 (91-100级)

### 2. 经验值公式

```
升级所需XP = 基础经验 * (1 + (等级-1) * 增长系数) ^ 缩放因子

参数（增强版 - 升级更难）:
- 基础经验 = 100 XP
- 增长系数 = 0.12 (每级增长12%)
- 缩放因子 = 2.0 (指数增长)
```

**示例**:
- 1级→2级: 100 XP
- 10级→11级: 484 XP
- 20级→21级: 1156 XP
- 30级→31级: 2116 XP
- 50级→51级: 4900 XP
- 70级→71级: 9604 XP
- 99级→100级: 16589 XP

**总经验值参考**:
- 达到 10级: ~2,000 XP (约20-30局)
- 达到 25级: ~15,000 XP (约150-200局)
- 达到 50级: ~100,000 XP (约1000-1500局)
- 达到 100级: ~650,000 XP (约6500-10000局)

### 3. 经验获取规则

#### 基础经验
- 参与游戏: +20 XP

#### 排名奖励
- 🥇 第1名: +50 XP
- 🥈 第2名: +30 XP
- 🥉 第3名: +20 XP
- 第4名: +10 XP
- 其他: +5 XP

#### 特殊成就
- 使用双联反应: +5 XP
- 打出复杂物质(5+原子): +3 XP
- 连击反应: +2 XP/次

#### 难度修正
**AI对战**:
- 难度 < 50: XP × 0.5
- 难度 50-100: XP × (难度/100)

**对战真人（等级差距系统）**:

*挑战高手奖励*：
- 每级高 +5% XP
- 最多 +100% XP（完全翻倍）
- 示例：
  - 对战高5级玩家：+25% XP
  - 对战高10级玩家：+50% XP
  - 对战高20级玩家：+100% XP（封顶）

*虐菜惩罚*：
- 等级差距 < 5级：无惩罚
- 低 5-9级：-40% XP（0.6倍）
- 低 10-19级：-60% XP（0.4倍）
- 低 20级以上：-80% XP（0.2倍）
- 示例：
  - 30级打25级：无惩罚
  - 30级打20级：-40% XP
  - 30级打15级：-60% XP
  - 30级打5级：-80% XP

**设计理念**：
- 鼓励挑战强者（最多翻倍奖励）
- 严惩虐菜行为（最多减少80%）
- 允许±5级内的正常匹配
- 促进玩家选择同等级对手

## 数据库结构

### Users 表新增字段
```sql
ALTER TABLE users
ADD COLUMN level INTEGER DEFAULT 1,
ADD COLUMN xp INTEGER DEFAULT 0,
ADD COLUMN total_xp INTEGER DEFAULT 0;
```

### LevelConfigs 表
```sql
CREATE TABLE level_configs (
    level INTEGER PRIMARY KEY,
    required_xp INTEGER NOT NULL,
    tier VARCHAR(20) NOT NULL,
    tier_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 已实现的文件

### 后端

1. **database/migrate_level_system.go**
   - 数据库迁移逻辑
   - 等级配置初始化
   - 自动生成1-100级配置

2. **game/level_system.go**
   - `CalculateXPReward()`: 计算经验奖励
   - `AwardXP()`: 授予经验并检查升级
   - `GetLevelInfo()`: 获取玩家等级信息

3. **handlers/level.go**
   - `GET /api/level/info`: 获取自己的等级信息
   - `GET /api/level/user/:uid`: 获取指定用户等级信息
   - `GET /api/level/leaderboard`: 等级排行榜
   - `GET /api/level/configs`: 获取所有等级配置

4. **repository/user_repository.go**
   - `UpdateXP()`: 更新经验和等级
   - `AddXP()`: 增加经验值

5. **game/manager.go**
   - 在 `handlePointsCalculation()` 中添加XP奖励

6. **models/user.go**
   - 新增 `LevelInfo` 结构体

## API 接口文档

### 获取自己的等级信息
```
GET /api/level/info
Authorization: Bearer <token>

Response:
{
  "level": 25,
  "xp": 150,
  "total_xp": 3500,
  "tier": "silver",
  "tier_name": "白银",
  "next_level_xp": 300,
  "progress_percent": 50
}
```

### 获取用户等级信息
```
GET /api/level/user/:uid

Response: (同上)
```

### 获取等级排行榜
```
GET /api/level/leaderboard?limit=100

Response:
[
  {
    "uid": 1,
    "nickname": "玩家A",
    "avatar": "...",
    "level": 85,
    "total_xp": 25000,
    "tier": "diamond",
    "tier_name": "钻石"
  },
  ...
]
```

### 获取等级配置
```
GET /api/level/configs

Response:
[
  {
    "level": 1,
    "required_xp": 100,
    "tier": "bronze",
    "tier_name": "青铜"
  },
  ...
]
```

## WebSocket 事件

### 升级通知
```javascript
// 当玩家升级时，会收到以下消息
{
  "type": "level_up",
  "data": {
    "level": 26,
    "tier": "gold",
    "tier_name": "黄金",
    "xp": 50,
    "total_xp": 4200
  }
}

// 同时会收到浮窗提示
{
  "type": "action_toast",
  "data": "🎉 恭喜升级！你现在是 黄金 26 级研究员！"
}
```

## 待实现功能

### 1. 等级匹配机制

需要在创建/加入房间时添加等级限制：

#### 快速匹配
```go
// 在 handlers/game.go 的 CreateRoom 或 JoinRoom 中添加
func (gr *GameRoom) CanJoin(playerLevel int) bool {
    if gr.Room.IsPrivate {
        return true // 私密房间豁免等级限制
    }

    // 获取房主等级
    hostLevel := getHostLevel(gr.Room.CreatedByUID)

    // ±5级匹配
    levelDiff := abs(playerLevel - hostLevel)
    return levelDiff <= 5
}
```

#### 排位模式
```go
// 创建房间时添加参数
type CreateRoomRequest struct {
    // ... 现有字段
    IsRanked       bool `json:"is_ranked"`       // 是否排位模式
    LevelRange     int  `json:"level_range"`     // 等级范围 (3/5/10)
}

// 严格匹配
func (gr *GameRoom) CanJoinRanked(playerLevel int) bool {
    if gr.Room.IsPrivate {
        return true
    }

    hostLevel := getHostLevel(gr.Room.CreatedByUID)
    levelDiff := abs(playerLevel - hostLevel)
    return levelDiff <= gr.Room.LevelRange
}
```

### 2. 前端显示等级信息

#### 用户资料卡片
```vue
<!-- 在所有显示玩家信息的地方添加等级显示 -->
<template>
  <div class="player-card">
    <img :src="player.avatar" />
    <div class="player-info">
      <span class="nickname">{{ player.nickname }}</span>
      <!-- 新增：等级徽章 -->
      <div class="level-badge" :class="tierClass">
        <span class="tier-icon">{{ tierIcon }}</span>
        <span class="level">Lv.{{ player.level }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
const tierIcons = {
  bronze: '🥉',
  silver: '🥈',
  gold: '🥇',
  platinum: '💎',
  diamond: '💠',
  master: '⭐'
}

const tierClass = computed(() => `tier-${player.tier}`)
const tierIcon = computed(() => tierIcons[player.tier] || '🎮')
</script>

<style>
.level-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.tier-bronze { background: linear-gradient(135deg, #cd7f32 0%, #b87333 100%); }
.tier-silver { background: linear-gradient(135deg, #c0c0c0 0%, #a8a8a8 100%); }
.tier-gold { background: linear-gradient(135deg, #ffd700 0%, #ffb700 100%); }
.tier-platinum { background: linear-gradient(135deg, #e5e4e2 0%, #b0c4de 100%); }
.tier-diamond { background: linear-gradient(135deg, #b9f2ff 0%, #4da6ff 100%); }
.tier-master { background: linear-gradient(135deg, #ff6b6b 0%, #ee5a6f 100%); }
</style>
```

#### 等级进度条组件
```vue
<!-- components/LevelProgress.vue -->
<template>
  <div class="level-progress-card">
    <div class="header">
      <h3>{{ tierName }} {{ level }} 级</h3>
      <span class="xp">{{ xp }} / {{ nextLevelXP }} XP</span>
    </div>

    <div class="progress-bar">
      <div class="progress-fill" :style="{ width: progressPercent + '%' }"></div>
    </div>

    <div class="stats">
      <span>总经验: {{ totalXP.toLocaleString() }} XP</span>
      <span>距离下一级: {{ (nextLevelXP - xp) }} XP</span>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { api } from '@/utils/api'

const level = ref(1)
const xp = ref(0)
const totalXP = ref(0)
const tierName = ref('青铜')
const nextLevelXP = ref(100)
const progressPercent = ref(0)

onMounted(async () => {
  const res = await api.get('/level/info')
  level.value = res.data.level
  xp.value = res.data.xp
  totalXP.value = res.data.total_xp
  tierName.value = res.data.tier_name
  nextLevelXP.value = res.data.next_level_xp
  progressPercent.value = res.data.progress_percent
})
</script>
```

#### 升级动画
```vue
<script setup>
import { onMounted } from 'vue'
import { useWebSocket } from '@/composables/useWebSocket'

const { ws } = useWebSocket()

onMounted(() => {
  ws.on('level_up', (data) => {
    // 显示升级动画
    showLevelUpAnimation({
      level: data.level,
      tier: data.tier_name,
      icon: getTierIcon(data.tier)
    })
  })
})

function showLevelUpAnimation(data) {
  // 实现华丽的升级动画
  // 可以使用 confetti、fireworks 等特效库
}
</script>
```

#### 修改位置清单

需要在以下前端组件中添加等级显示：

1. **src/pages/Profile.vue** - 个人资料页
2. **src/pages/Lobby.vue** - 大厅玩家列表
3. **src/pages/GameRoom.vue** - 游戏房间玩家列表
4. **src/components/ChatBox.vue** - 聊天用户名旁
5. **src/pages/Ranking.vue** - 排行榜
6. **src/components/profile/MatchHistory.vue** - 战绩历史

### 3. 匹配机制实现步骤

#### Step 1: 修改房间创建
```go
// models/game.go 添加字段
type Room struct {
    // ... 现有字段
    IsRanked       bool `json:"is_ranked"`
    LevelRange     int  `json:"level_range"`  // 3, 5, 或 10
    MinLevel       int  `json:"min_level"`
    MaxLevel       int  `json:"max_level"`
}
```

#### Step 2: 修改创建房间API
```go
// handlers/game.go
func CreateRoom(c *gin.Context) {
    // ... 现有逻辑

    // 计算等级范围
    if !req.IsPrivate && req.IsRanked {
        user, _ := repository.UserRepo.FindByUID(uid)
        room.MinLevel = user.Level - req.LevelRange
        room.MaxLevel = user.Level + req.LevelRange
    }
}
```

#### Step 3: 修改加入房间验证
```go
// handlers/game.go
func JoinRoom(c *gin.Context) {
    // ... 现有验证

    // 等级验证
    if room.IsRanked && !room.IsPrivate {
        user, _ := repository.UserRepo.FindByUID(uid)
        if user.Level < room.MinLevel || user.Level > room.MaxLevel {
            c.JSON(400, gin.H{"error": "等级不符合房间要求"})
            return
        }
    }
}
```

## 启动项目后的初始化

启动服务后，等级系统会自动：
1. ✅ 创建 `level_configs` 表
2. ✅ 初始化1-100级配置
3. ✅ 为所有现有用户添加 level、xp、total_xp 字段（默认值）
4. ✅ 游戏结算时自动授予XP

## 测试建议

### 1. 经验计算测试
```bash
# 创建一个测试游戏，确保XP正确授予
# 观察日志输出：[XP] 玩家 1 获得 75 XP (基础:20 排名:50 成就:5 倍率:1.00)
```

### 2. 升级测试
```bash
# 手动添加XP测试升级
UPDATE users SET xp = 95, level = 1 WHERE uid = 1;
# 再玩一局，应该会升级并收到通知
```

### 3. API测试
```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/level/info
curl http://localhost:8080/api/level/leaderboard
curl http://localhost:8080/api/level/configs
```

## 总结

等级系统的核心后端功能已完成，包括：
- ✅ 完整的数据库设计和迁移
- ✅ 智能的经验计算算法
- ✅ 自动升级和通知系统
- ✅ RESTful API 接口
- ✅ 游戏结算集成

还需完成：
- ⏳ 等级匹配机制（后端）
- ⏳ 前端UI组件和动画
- ⏳ 等级相关的游戏逻辑优化

估计还需2-3小时完成剩余部分。
