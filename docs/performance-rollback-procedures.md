# Performance Optimization Rollback Procedures

所有优化均通过环境变量特性开关控制，支持**无代码变更回滚**（只需修改 `.env` 并重启）。

---

## 快速回滚参考表

| 优化项 | 回滚操作 | 恢复时间 |
|--------|----------|---------|
| 反应缓存 | `ENABLE_REACTION_CACHE=false` + 重启 | < 1 分钟 |
| 速率限制清理 | `RATE_LIMIT_CLEANUP_ENABLED=false` + 重启 | < 1 分钟 |
| 游戏历史优化查询 | `USE_OPTIMIZED_HISTORY_QUERIES=false` + 重启 | < 1 分钟 |
| 反作弊批量查询 | `ENABLE_ANTICHEAT_BATCH=false` + 重启 | < 1 分钟 |
| WebSocket 锁优化 | 需代码回滚（见下方） | ~5 分钟 |
| 数据库迁移（junction table） | 删除表（不影响业务逻辑） | ~5 分钟 |

---

## 各阶段详细回滚步骤

### Phase 1: Prometheus 监控指标

**回滚原因：** `/metrics` 端点暴露了不应公开的信息，或 Prometheus 收集导致资源消耗

**回滚步骤：**
1. 在 `backend/router/api_routes.go` 注释掉 `/metrics` 路由
2. 如需保留但限制访问，在路由前添加 IP 白名单中间件
3. 编译部署

**数据影响：** 无，仅影响可观测性

---

### Phase 2a: 反应缓存

**回滚步骤（环境变量回滚）：**
```bash
# 在 .env 中设置
ENABLE_REACTION_CACHE=false
```
然后重启服务。

**行为恢复：** `JudgeReaction` 跳过缓存，直接计算，回到原始延迟（30-50μs）

**数据影响：** 无，缓存不持久化，重启后自然清空

---

### Phase 2b: 速率限制内存清理

**回滚步骤（环境变量回滚）：**
```bash
RATE_LIMIT_CLEANUP_ENABLED=false
```
然后重启服务。

**行为恢复：** 速率限制器不再自动清理，长期运行内存会缓慢增长（恢复为原始行为）

**数据影响：** 无，内存中数据重启后消失

---

### Phase 2c: WebSocket 锁优化

**回滚原因：** 发现并发 bug 或消息重复发送

**回滚步骤（代码回滚）：**
```bash
git diff HEAD~1 -- backend/websocket/hub.go
git checkout HEAD~1 -- backend/websocket/hub.go
go build ./...
```

或者直接还原 `BroadcastToRoom`、`BroadcastToAll`、`SendToUID` 使用原有的 `defer h.mutex.RUnlock()` 模式。

**数据影响：** 无，仅影响消息发送机制

---

### Phase 3: 数据库迁移（Junction Table）

**回滚步骤（禁用查询优化）：**
```bash
USE_OPTIMIZED_HISTORY_QUERIES=false
```
重启后即恢复使用 LIKE 查询。

**完全回滚（删除 junction table，如有必要）：**
```sql
-- MySQL
DROP TABLE IF EXISTS game_history_players;

-- 验证现有查询仍然工作
SELECT COUNT(*) FROM game_history WHERE players LIKE '%1%';
```

**注意：** junction table 是纯加速索引，删除不影响任何业务数据。迁移是幂等的，重新运行不会报错。

---

### Phase 4: 反作弊批量查询

**回滚步骤（环境变量回滚）：**
```bash
ENABLE_ANTICHEAT_BATCH=false
```
重启后恢复单次查询模式。

**行为恢复：** `enrichDetectionContext` 回到逐个调用 `GetPlayerBaselines` 和 `GetPlayerRiskProfile`

**数据影响：** 无，仅影响查询策略

---

## 回滚验证清单

回滚后，通过以下步骤确认系统恢复正常：

1. **基本功能验证：**
   ```bash
   # 健康检查
   curl http://localhost:8080/api/health
   
   # 检查游戏大厅
   curl -H "Authorization: Bearer <token>" http://localhost:8080/api/game/rooms
   ```

2. **数据库连接验证：**
   - 检查 `/api/health` 返回 `"database": "ok"`

3. **游戏历史查询验证：**
   - 登录后进入个人主页，确认历史记录正常加载

4. **性能指标（如保留监控）：**
   - 检查 P95 延迟是否回到回滚前水平

---

## 紧急情况联系

如果回滚后仍有问题：
1. 检查服务日志中是否有 `panic` 或 `FATAL` 关键词
2. 确认数据库连接池无泄漏（`SHOW PROCESSLIST` 在 MySQL 中）
3. 如果 Redis 不可用导致连锁故障，临时将所有特性开关设为 `false` 并重启
