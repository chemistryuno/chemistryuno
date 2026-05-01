## 1. 数据模型和数据库扩展

- [x] 1.1 在 `backend/database/models.go` 中新增 `CheatRiskScore` 模型
- [x] 1.2 在 `backend/database/models.go` 中新增 `CheatSanction` 模型
- [x] 1.3 在 `backend/database/models.go` 中新增 `CheatAppeal` 模型
- [x] 1.4 在 `backend/database/models.go` 中新增 `CheatAuditLog` 模型
- [x] 1.5 创建数据库迁移脚本添加新表
- [x] 1.6 在 `backend/repository/` 中创建 `cheat_repository.go` 用于数据库操作

## 2. 核心风险评分引擎

- [x] 2.1 在 `backend/anticheat/` 中创建 `risk_engine.go`，定义 `RiskScoringEngine` 结构
- [x] 2.2 实现 `DetectionStrategy` 接口和注册机制
- [x] 2.3 实现 `ResponseTimeDetector` 策略（分析响应时间分布）
- [x] 2.4 实现 `FrequencyDetector` 策略（检测操作频率异常）
- [x] 2.5 实现 `WinRateDetector` 策略（检测胜率异常）
- [x] 2.6 实现 `PatternDetector` 策略（检测操作模式异常）
- [x] 2.7 实现 `AccountAgeDetector` 策略（考虑账号新旧程度）
- [x] 2.8 实现加权平均计算和最终风险分数生成

## 3. 配置管理系统

- [x] 3.1 在 `backend/config/` 中创建 `anticheat.yaml` 配置文件
- [x] 3.2 创建 `backend/anticheat/config.go` 实现配置加载和解析
- [x] 3.3 实现动态配置重载机制（无需重启应用）
- [x] 3.4 在配置中定义各维度权重、阈值、处罚等级

## 4. 处罚决策和执行

- [x] 4.1 在 `backend/anticheat/` 中创建 `sanction.go` 定义处罚类型和规则
- [x] 4.2 实现 `SanctionDecider` 根据风险分数确定处罚类型
- [x] 4.3 实现处罚的持久化保存（到数据库）
- [x] 4.4 实现警告消息通知逻辑
- [x] 4.5 实现禁言状态标记和检查
- [x] 4.6 实现与账号封禁流程的集成

## 5. 玩家申诉系统

- [x] 5.1 在 `backend/anticheat/` 中创建 `appeal.go` 实现申诉管理
- [x] 5.2 实现提交申诉的接口
- [x] 5.3 实现申诉状态管理（待审核、已批准、已拒绝）
- [x] 5.4 实现撤销处罚的逻辑（当申诉被批准）
- [x] 5.5 实现申诉历史查询

## 6. 审计日志系统

- [x] 6.1 在 `backend/anticheat/` 中创建 `audit.go` 实现审计日志记录
- [x] 6.2 在游戏结束时自动记录风险评分日志
- [x] 6.3 在处罚执行时记录处罚日志
- [x] 6.4 在人工审核时记录审核日志
- [x] 6.5 实现按玩家ID查询审计日志
- [x] 6.6 实现按时间范围导出审计日志

## 7. 管理员 API 接口

- [x] 7.1 在 `backend/handlers/admin.go` 中新增路由和方法处理
- [x] 7.2 实现 `GET /api/admin/anticheat/detection-list` 查询检测列表
- [x] 7.3 实现 `GET /api/admin/anticheat/detection/:id` 查询检测详情
- [x] 7.4 实现 `POST /api/admin/anticheat/detection/:id/review` 人工审核接口
- [x] 7.5 实现 `GET /api/admin/anticheat/appeals` 查询申诉列表
- [x] 7.6 实现 `POST /api/admin/anticheat/appeals/:id/approve` 批准申诉
- [x] 7.7 实现 `POST /api/admin/anticheat/appeals/:id/reject` 拒绝申诉
- [x] 7.8 实现 `GET /api/admin/anticheat/config` 获取当前配置
- [x] 7.9 实现 `POST /api/admin/anticheat/config` 更新配置
- [x] 7.10 实现 `GET /api/admin/anticheat/audit-log` 查询审计日志

## 8. 玩家 API 接口

- [x] 8.1 在 `backend/handlers/game.go` 中新增路由
- [x] 8.2 实现 `GET /api/game/:roomId/anticheat-check` 查询当前房间检测结果
- [x] 8.3 实现 `POST /api/game/:roomId/appeal` 提交申诉接口
- [x] 8.4 实现 `GET /api/player/appeals` 查询个人申诉历史
- [x] 8.5 实现 `GET /api/player/sanctions` 查询个人处罚历史

## 9. 游戏循环集成

- [x] 9.1 在 `backend/game/manager.go` 的 `OnGameFinish()` 中调用风险评分引擎
- [x] 9.2 在 `OnGameFinish()` 中获取处罚结果并应用
- [x] 9.3 修改 `captureReplaySnapshotLocked()` 以使用新的风险评分（而非简单的快速反应计数）
- [x] 9.4 添加玩家动作处理中的数据收集（响应时间、操作时间戳等）

## 10. 前端集成

- [x] 10.1 创建玩家通知组件，显示检测到的异常行为和处罚类型
- [x] 10.2 创建申诉表单，允许玩家提交申诉和查看历史
- [ ] 10.3 在管理后台创建检测列表页面
- [ ] 10.4 在管理后台创建检测详情和人工审核页面
- [ ] 10.5 在管理后台创建申诉管理页面
- [ ] 10.6 在管理后台创建配置管理页面
- [ ] 10.7 在管理后台创建审计日志查询页面

## 11. 单元测试和集成测试

- [x] 11.1 为风险评分引擎编写单元测试
- [x] 11.2 为各检测策略编写单元测试
- [x] 11.3 为处罚决策编写单元测试
- [x] 11.4 为申诉系统编写单元测试
- [x] 11.5 为审计日志编写单元测试
- [x] 11.6 编写端到端集成测试（从游戏结束到处罚应用）
- [x] 11.7 编写管理员 API 测试
- [x] 11.8 编写玩家 API 测试

## 12. 文档和配置示例

- [ ] 12.1 编写反作弊系统的完整文档
- [ ] 12.2 提供配置文件示例
- [ ] 12.3 编写管理员使用指南
- [ ] 12.4 编写玩家申诉流程说明
- [ ] 12.5 创建反作弊系统的架构图

## 13. 性能优化和部署

- [ ] 13.1 进行性能基准测试（风险评分的延迟）
- [ ] 13.2 优化数据库查询（添加适当的索引）
- [ ] 13.3 实现审计日志的异步处理
- [ ] 13.4 制定灰度上线计划
- [ ] 13.5 准备回滚方案

## 14. 监控和数据分析

- [ ] 14.1 添加监控指标（检测数、风险分布、申诉率等）
- [ ] 14.2 创建数据分析脚本用于评估模型准确性
- [ ] 14.3 建立定期审查流程（每周/每月）
- [ ] 14.4 定义参数调整的标准和流程
