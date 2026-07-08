# Performance Monitoring Runbook

## 概述

本文档描述如何使用 ChemistryUNO 内置的 Prometheus 指标监控系统性能，并提供常见问题的排查指南。

---

## 指标端点

服务启动后，性能指标通过标准 Prometheus 格式暴露：

```
GET /metrics
```

**示例：** `curl http://localhost:8080/metrics`

---

## 核心指标

### API 请求延迟

```promql
# P95 延迟（按端点）
histogram_quantile(0.95,
  sum(rate(chemistryuno_api_request_duration_seconds_bucket[5m])) by (endpoint, le)
)

# 所有端点平均延迟
rate(chemistryuno_api_request_duration_seconds_sum[5m])
/ rate(chemistryuno_api_request_duration_seconds_count[5m])
```

**告警阈值：**
- Warning: P95 > 200ms
- Critical: P95 > 500ms

### 数据库查询延迟

```promql
# P99 数据库查询延迟（按表）
histogram_quantile(0.99,
  sum(rate(chemistryuno_db_query_duration_seconds_bucket[5m])) by (table, le)
)

# 慢查询速率（>100ms）
rate(chemistryuno_db_query_duration_seconds_bucket{le="0.1"}[5m])
```

**告警阈值：**
- Warning: P99 > 100ms
- Critical: P99 > 500ms

### 缓存命中率

```promql
# 反应缓存命中率
rate(chemistryuno_cache_hits_total{cache="reaction"}[5m])
/
(
  rate(chemistryuno_cache_hits_total{cache="reaction"}[5m])
  + rate(chemistryuno_cache_misses_total{cache="reaction"}[5m])
)
```

**目标：** 缓存热身（30分钟）后命中率 > 80%

### WebSocket 在线计数

```promql
# 当前在线用户数
chemistryuno_online_users_total

# WebSocket 广播速率
rate(chemistryuno_websocket_messages_total[1m])
```

---

## Grafana Dashboard 配置

在 Grafana 中创建 Dashboard，添加以下 Panel：

### Panel 1: API P95 Latency Heatmap
- **类型:** Heatmap
- **Query:** `histogram_quantile(0.95, sum(rate(chemistryuno_api_request_duration_seconds_bucket[5m])) by (le, endpoint))`

### Panel 2: Cache Hit Rate
- **类型:** Gauge (0-100%)
- **Query:** 上方缓存命中率公式，目标线 80%

### Panel 3: Online Users
- **类型:** Stat
- **Query:** `chemistryuno_online_users_total`

### Panel 4: DB Query P99 by Table
- **类型:** Bar Gauge
- **Query:** 上方 DB P99 公式

### Panel 5: WebSocket Broadcast Throughput
- **类型:** Time Series
- **Query:** `rate(chemistryuno_websocket_messages_total[1m])`

---

## 告警规则（Prometheus AlertManager）

```yaml
groups:
  - name: chemistryuno_performance
    rules:
      - alert: HighAPILatency
        expr: |
          histogram_quantile(0.95,
            sum(rate(chemistryuno_api_request_duration_seconds_bucket[5m])) by (endpoint, le)
          ) > 0.5
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "API P95 latency > 500ms on {{ $labels.endpoint }}"

      - alert: LowCacheHitRate
        expr: |
          (
            rate(chemistryuno_cache_hits_total{cache="reaction"}[10m])
            / (rate(chemistryuno_cache_hits_total{cache="reaction"}[10m]) + rate(chemistryuno_cache_misses_total{cache="reaction"}[10m]))
          ) < 0.5
        for: 30m
        annotations:
          summary: "Reaction cache hit rate below 50%"

      - alert: HighDBLatency
        expr: |
          histogram_quantile(0.99,
            sum(rate(chemistryuno_db_query_duration_seconds_bucket[5m])) by (table, le)
          ) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "DB P99 latency > 100ms on table {{ $labels.table }}"
```

---

## 常见问题排查

### 缓存命中率低

**症状：** `cache_misses_total{cache="reaction"}` 增长速率远高于 `cache_hits_total`

**排查步骤：**
1. 确认 `ENABLE_REACTION_CACHE=true` 环境变量已设置
2. 检查服务是否刚重启（缓存冷启动期约 30 分钟）
3. 检查 `reaction_cache.go` 的 LRU 驱逐率是否过高（maxSize=10000）

**解决：** 冷启动期正常，30 分钟后命中率应恢复。如持续低，考虑增大 `maxSize`。

### 游戏历史查询慢

**症状：** `chemistryuno_db_query_duration_seconds{table="game_history"}` P95 > 100ms

**排查步骤：**
1. 检查 `game_history_players` 表是否存在且有数据：
   ```sql
   SELECT COUNT(*) FROM game_history_players;
   ```
2. 确认 `USE_OPTIMIZED_HISTORY_QUERIES=true` 已启用
3. 检查是否有慢查询日志（`SLOW_QUERY_LOG_LEVEL=warn`）

**解决：** 启用优化查询后，必须等待 `game_history_players` 填充（新游戏自动填充，历史数据需手动 backfill）。

### 内存持续增长

**症状：** 进程内存随时间线性增长，无稳定

**排查步骤：**
1. 检查速率限制清理是否运行：日志中搜索 `Rate limiter cleanup completed`
2. 确认 `RATE_LIMIT_CLEANUP_ENABLED=true`
3. 检查 cleanup 日志中的 `keys_before` vs `keys_after`

**解决：** 设置 `RATE_LIMIT_CLEANUP_ENABLED=true` 并重启服务。

### WebSocket 广播延迟高

**症状：** `chemistryuno_websocket_broadcast_duration_seconds` P99 > 10ms

**排查步骤：**
1. 检查当前在线用户数（`chemistryuno_online_users_total`）
2. 高并发时正常，检查是否有连接积压（send 通道满）

---

## 性能基准目标

| 指标 | 基准（优化前） | 目标（优化后） |
|------|---------------|---------------|
| 游戏历史查询 P95 | 500ms+ | < 50ms |
| 反应判断延迟 | 30-50μs | < 10μs (cached) |
| 反作弊处理延迟 | 250ms | < 100ms |
| WebSocket 广播 (50 clients) | 基准 | 2-3x 提升 |
| 速率限制器内存 | 无限增长 | 稳定（24h+）|
