# 反作弊指标重新设计方案

> 状态：设计草案（待评审）
> 范围：重新设计检测指标体系；处置策略沿用现有分档，数据采取直接替换
> 关联代码：`backend/anticheat/`、`backend/game/manager.go`（数据采集）、`backend/database/models.go`（CheatRiskScore）

---

## 1. 背景与目标

Chemistry UNO 是**回合制、限时出牌、以化学反应判定为核心**的多人卡牌游戏。
它的作弊向量与常见实时竞技游戏（FPS/MOBA）截然不同：

- 没有瞄准/移动，无法用"人类反应极限"这类经典指标
- 回合制天然限制操作频率，高频操作检测意义有限
- **真正的作弊是"化学知识作弊"**：用外部工具/脚本瞬间算出最优反应组合、
  从不失误、面对复杂化学式零思考

因此本次重设计的目标：

1. **只用真实可测的信号**，剔除当前编造/估算的输入
2. **针对化学回合制游戏的作弊指纹**设计指标（决策质量 + 思考时长）
3. 从"多个弱指标简单加权"转向"少数强证据 + 人群统计基线"
4. 降低对高手和新手的误判率

---

## 2. 现有系统的根本缺陷

### 2.1 `response_time` 维度的输入是编造的（最严重，权重 0.25）

`backend/game/manager.go:848`：

```go
if count, exists := gr.FastReactionUIDs[uid]; exists {
    context.ResponseTimes = append(context.ResponseTimes, int64(count*50)) // 估计响应时间
}
```

`FastReactionUIDs[uid]` 只是"快速出牌"事件的**计数器**，`count*50` 是把次数当毫秒。
`ResponseTimeDetector`（`detectors.go:29`）对这个假数据算平均、算异常比例、
对 <50ms 额外 ×1.5 惩罚——全部建立在虚构输入上。这个 0.25 权重的维度无效。

### 2.2 `pattern`（操作间隔方差）拿不到有效数据（权重 0.15）

依赖 `OperationTimes`，但时间戳由**服务器**在 `appendReplayEventLocked` 时打
（`manager.go:717`，`now.Format(time.RFC3339Nano)`），反映的是网络到达+服务器处理
顺序，而非客户端真实操作时刻。用它算"机器人式规律节奏"的方差不成立。

### 2.3 `win_rate` 用全局累计胜率且每局触发（权重 0.20）

`context.WinCount / TotalGames`（`manager.go:882`）是账号**历史总胜率**。问题：

- 高手（200 局 90%）每局结束都被扣分并反复标记 → 主要误判来源
- 无法区分"长期高手"和"近期突然异常"
- 未考虑对手强度：虐菜和战胜高手同等计分

### 2.4 `frequency` 对回合制意义不大（权重 0.25）

"10 秒内 20 次操作"在回合制限时出牌里几乎不可能自然触发，也无法区分作弊与
网络重传/重连补发。

### 2.5 `account_age` 双重惩罚

- 作为独立维度给分：`detectors.go:246`，<1 天给 20 基础分
- 又在引擎里给**所有**维度 ×1.5：`risk_engine.go:253`

新手被惩罚两次，加剧新手误判（新手保护特性也只是在此之上打补丁）。

### 2.6 缺失游戏特有的核心检测

当前无任何维度检测"决策质量异常"——而这恰恰是化学 UNO 最强的作弊指纹：
新账号却总能打出最优双联反应、从不失误、面对复杂反应零思考时间。

### 2.7 评分结构问题

- 简单加权平均会**稀释强证据**：一个决定性异常被一堆 0 分维度拉低
- 硬编码阈值（100ms / 85% / 50ms）缺乏人群分布依据
- 已有的 `adaptive_threshold` / `zscore` 优化维度方向正确，但默认关闭、仅作补充

---

## 3. 新指标体系

核心原则：**真实信号 + 人群基线 + 证据分级**。

| 指标 | 检测目标 | 数据来源 | 权重 | 需改造 |
|------|---------|---------|------|--------|
| `decision_optimality` | 出牌长期接近理论最优，尤其低经验账号 | 化学引擎实时算最优解 vs 实际出牌 | 0.30 | 服务端 |
| `think_time` | 复杂局面决策耗时超人 | 客户端埋点：回合开始→提交真实耗时 | 0.25 | 前端+服务端 |
| `recent_performance` | 近期滑动窗口胜率 + 对手强度 | 近 N 局结果 + 对手等级 | 0.15 | 服务端 |
| `multi_account` | 多开/小号（同 IP/设备聚集） | 登录侧 IP/设备指纹 | 0.20 | 登录侧 |
| `player_reports` | 玩家举报去重加权 | 已有 | 0.10 | 无 |

> 说明：本文档聚焦指标定义与评分。`think_time` 的前端埋点与 `multi_account`
> 的登录侧信号是独立子项，可分阶段落地（见第 6 节灰度计划）。

### 3.1 `decision_optimality`（决策最优度）— 杀手锏指标

**思路**：复用现有化学引擎（`backend/game/judge.go` 的 `JudgeReaction` /
`computeReaction`）。每次玩家出牌时，服务端根据其**当时手牌 + 场面状态**枚举所有
合法出法，计算"是否打出了最优/次优解"。单局结束后汇总匹配率。

**为什么强**：
- 真人即使是高手也会有次优选择、偶尔失误、保守打法
- 脚本/工具作弊者会**长期稳定接近 100% 最优**，这是无法伪装的统计指纹
- 与账号经验交叉：一个 10 局的新账号却有 98% 最优率，极可疑

**评分（0-100）**：
```
optimalityRate = 最优出牌数 / 总决策数      // 本局
样本足够时（决策数 ≥ minDecisions，如 15）:
  score = sigmoid((optimalityRate - populationMean) / populationStdDev)
  即：偏离人群均值越多标准差，分数越高
样本不足：按比例衰减 score *= 决策数 / minDecisions
```
用**人群基线**（全局 optimalityRate 的均值/标准差）而非硬阈值，自动适应版本变化。

### 3.2 `think_time`（思考时长异常）

**思路**：修复 2.1 的根本问题——真实决策耗时必须由**客户端**测量。
客户端在"轮到我"时记 `turnStartMs`，提交出牌时记 `submitMs`，上报
`thinkMs = submitMs - turnStartMs`。服务端做合理性校验（不得超过回合限时、
不得为负、与服务器观测的到达间隔大致吻合，防伪造）。

**评分**：
```
对每个决策，按"局面复杂度"分档（可选反应数、是否双联、化学式复杂度）：
  复杂局面却 thinkMs 极短（如 < 300ms）→ 累积异常
score = 复杂决策中的"超人决策"占比 归一化到 0-100
```
关键：**只对复杂局面计分**。简单局面（只有一种合法出法）秒出是正常的。

**防伪造**：客户端上报的 thinkMs 与服务器测得的 `submitArrival - turnBroadcast`
做上界校验（thinkMs 不能显著大于服务器观测窗口）。伪造大 thinkMs 会被夹住，
伪造小 thinkMs 才是作弊者动机——而那正是我们要抓的。

### 3.3 `recent_performance`（近期战绩）

替换旧 `win_rate`。改为**近 N 局滑动窗口**（如最近 20 局）：
```
recentWinRate = 窗口内胜局 / 窗口局数
opponentFactor = 窗口内对手平均等级 / 自身等级   // 战胜强敌加权更高
score = f(recentWinRate, opponentFactor)，样本不足则衰减
```
好处：长期高手的历史不再反复计分；突然的近期异常连胜（尤其虐强敌）才触发。

### 3.4 `multi_account`（多开/小号）

同 IP / 设备指纹在短时间内聚集多个账号，或小号给主号送分。
数据来自登录侧（需接入）。**此维度不受新手保护放宽**（防小号钻空子，
现有 `optimization_scoring.go:216` 已有 `multiAccountSignals` 白名单机制，保留）。

### 3.5 `player_reports`（举报信号）

沿用现有实现（`risk_engine.go:276`）：去重后按数量加权，作为**辅助**证据，
权重 0.10，不单独定罪。

---

## 4. 评分算法：从"加权平均"到"证据分级"

### 4.1 问题

现有 `weightedTotal / totalWeight`（`risk_engine.go:311`）是加权**平均**，
单个强证据会被一堆 0 分维度稀释。例如 decision_optimality=95 但其他维度=0，
平均下来可能只有 30 分，够不上处罚。

### 4.2 新方案：加权和 + 强证据下限（floor）

```
base = Σ(score_i × weight_i)                    // 加权和（不除以总权重）
      归一化到 0-100（按最大可能加权和）

// 强证据下限：任一核心指标极端异常时，风险分不低于某个 floor
floor = 0
for each 核心指标 i in {decision_optimality, think_time, multi_account}:
    if score_i ≥ strongEvidenceThreshold(如 85):
        floor = max(floor, evidenceFloor[i])   // 如 decision_optimality→60

riskScore = clamp(max(base, floor), 0, 100)
```

这样：
- 多个中等信号叠加 → base 抬高（协同证据）
- 单个决定性证据 → floor 保底触发人工复核，不被稀释

### 4.3 处置分档：沿用现有阈值

不改 `SanctionThresholds`（observe 20-40 / warning 40-60 / mute 60-80 / ban 80-100）
和 `sanction.go` 的处置逻辑。**自动封禁保持保守**：建议 decision_optimality 等
新指标初期只用于 observe/warning，ban 仍需人工复核（延续现有 `NewPlayerObserve`
抑制自动封禁的思路）。

### 4.4 保留并转正的优化特性

- `adaptive_threshold` + `zscore`：本次设计的 decision_optimality / think_time
  评分**内建**人群基线思想，等价于把这两个优化维度的理念转正为默认。
  旧的独立优化维度可下线或合并。
- `risk_decay`（历史风险衰减）：保留，防止一次异常永久定罪。
- `new_player`（新手保护）：保留，但因新指标本身已交叉账号经验，可调低放宽力度。

## 5. 数据采集改造（关键前置）

指标再好，采集不到真实信号就是空谈。需要改造 `DetectionContext` 的填充：

### 5.1 服务端（`manager.go` collectAnticheatDataLocked）

- **移除** `ResponseTimes = count*50` 的假数据（2.1）
- **新增** 每次出牌时记录 `decision_optimality` 所需数据：
  当时手牌快照、场面物质、玩家实际出法、引擎算出的最优出法集合。
  建议在 `PlayCard` / `DoublePlay` 成功路径落一条结构化 replay 事件，
  单局结束时汇总最优匹配率。
- **新增** `recent_performance`：查近 N 局 game_history + 对手等级
  （`GameHistoryPlayer` junction 表已有，见 README 性能优化项）。

### 5.2 前端（think_time 埋点）

- 在 GameRoom 收到"轮到我"（turn_end_time 更新且 isMyTurn）时记 `turnStartMs`
- 出牌 API 请求体带上 `think_ms`
- 出牌 handler（`backend/handlers/game.go`）接收并透传给 replay 事件
- 反作弊读取真实 think_ms，服务端做上界校验（见 3.2 防伪造）

### 5.3 化学引擎复用

`judge.go` 的 `JudgeReaction(s1, s2)` 已可判两物质能否反应。需要一个新的
只读辅助（放 `backend/game/` 或 `backend/anticheat/`）：给定手牌+场面，
枚举合法出法并按"能否反应/是否双联/消耗牌数"排序，得出最优解集合。
**注意**：该计算较重，应在出牌当下（已持有游戏状态）顺带算，
不要在反作弊阶段重放整局。

---

## 6. 数据库迁移（直接替换策略）

现有 `CheatRiskScore` 表把 5 个旧维度**硬编码为列**
（`ResponseTimeDim` / `FrequencyDim` / `WinRateDim` / `PatternDim` / `AccountAgeDim`，
`models.go:490-494`）。新指标不同，方案：

1. **保留** `IndicatorDetails JSON` 列作为**唯一的维度明细存储**（已存在，
   本就是 `[]RiskIndicatorDetail` 的 JSON）。新指标全部走这里，schema 无需为
   每个新维度加列——避免重蹈硬编码覆辙。
2. 旧的 5 个 `*Dim` 列：AutoMigrate 不会删列，**保留但停止写入**（置 0），
   或后续单独清理迁移。前端详情面板改为纯读 `IndicatorDetails`（AdminAnticheat
   已主要依赖 `indicator_details`，见组件 `translateIndicatorName`）。
3. `EffectiveWeights` / `ThresholdSource` / `BaselineSnapshot` 等溯源列复用。
4. 因选择"直接替换"：旧 `CheatRiskScore` 历史记录**保留可查**，但新检测一律
   按新指标产出；不再生成旧式评分。管理端历史面板需兼容两种 `indicator_details`
   结构（旧维度名 vs 新维度名，`translateIndicatorName` 补齐新名称映射）。

---

## 7. 灰度与验证计划

即便选择"直接替换"，仍建议**先影子运行**再切换处置，避免误封：

1. **阶段一（影子）**：新指标计算并写入 `IndicatorDetails`，但 `SuggestedAction`
   仍可先不据其自动处置——先积累数据，对比新旧评分分布，标定人群基线
   （decision_optimality 均值/标准差、think_time 复杂局面阈值）。
2. **阶段二（observe/warning）**：新指标接入处置，但只到 warning 档，
   ban 仍纯人工。观察申诉率。
3. **阶段三（全量）**：确认误判率可接受后，新指标全权重生效。

配置开关：沿用 `anticheat.yaml` 的 `enabled_strategies` 机制，新指标各自可开关。

---

## 8. 需要新写/改的文件清单（供实现阶段参考）

**服务端评分**
- `backend/anticheat/detectors.go`：删旧 4 检测器，新增 decision_optimality /
  think_time / recent_performance 检测器
- `backend/anticheat/risk_engine.go`：改评分为"加权和 + 强证据 floor"（4.2）；
  移除 account_age 全局 ×1.5 双重惩罚（2.5）；DetectionContext 增新字段
- `backend/anticheat/config.go` + `backend/config/anticheat.yaml`：新维度权重/阈值
- `backend/anticheat/optimization_scoring.go`：合并/下线旧 adaptive/zscore 独立维度

**数据采集**
- `backend/game/manager.go`：collectAnticheatDataLocked 改造（5.1）；
  出牌路径记录最优解数据
- `backend/game/judge.go` 或新文件：最优出法枚举辅助（5.3）
- `backend/handlers/game.go`：接收前端 think_ms

**前端**
- `frontend/src/pages/GameRoom.vue` + `useCardActions.ts`：think_time 埋点上报
- `frontend/src/pages/AdminAnticheat.vue`：`translateIndicatorName` 增新指标中文名

**数据库**
- 无需加列（走 IndicatorDetails JSON）；旧 *Dim 列停写

**测试**
- `tests/_backend/anticheat/`：新检测器单测、评分 floor 逻辑、防伪造校验
- 人群基线标定脚本（阶段一数据分析）

---

## 9. 风险与权衡

- **decision_optimality 计算成本**：需在出牌当下枚举合法出法。手牌通常不大
  （初始手牌数可配），组合有限，成本可控；但双联反应是组合爆炸点，需限枚举深度。
- **think_time 依赖客户端诚实**：已用服务端上界校验缓解；且作弊动机是"伪造小
  think_time"，正是我们要抓的，伪造方向对我们有利。
- **人群基线冷启动**：新指标上线初期无基线，阶段一影子运行专门解决此问题。
- **版本漂移**：化学数据/规则更新会改变"最优解"分布，基线需定期重标（自适应
  机制已内建，但需监控）。


