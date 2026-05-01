# 实施任务（Implementation Tasks）

1. 创建 change scaffold
   - 文件：`openspec/changes/unban-message-fuel-compensation/.openspec.yaml`（已创建）

2. 添加配置项
   - 修改文件：`config/anticheat.yaml`（或统一配置文件）
   - 新增配置：`unban.compensation_amount`（默认 100）、`unban.default_message`（默认文案）

3. 后端 - 仓库方法
   - 修改文件：`repository/user_repository.go`
   - 新增方法：`AddFuel(userID int64, amount int) error`（确保幂等与事务支持）

4. 后端 - 业务逻辑
   - 修改文件：`backend/anticheat/appeal.go`（或相应解封逻辑处）
   - 在解封流程中加入：读取配置 → 写解封记录 → 发消息 → 调用 `AddFuel` → 写补偿审计
   - 处理异常：若发放失败，记录 `compensation_status=failed` 并触发告警/记录需人工补偿

5. HTTP/Handler 层
   - 修改文件：`handlers/appeal.go`
   - 确保管理员权限校验和请求参数校验；支持可选覆盖文案与金额的请求参数（供运营使用）

6. 审计字段更新
   - 在现有解封审计记录结构中增加：`compensation_amount`、`compensation_status`、`compensation_note`。
   - 如果无法扩展现有模型，创建数据库迁移并在 `migrations` 中记录变更。

7. 模板与 i18n
   - 将默认文案放入配置或消息模板位置，支持多语言替换（后续优化）。

8. 测试
   - 单元测试：为 `AddFuel`、模板渲染、审计记录写入添加测试。
   - 集成测试：`backend/anticheat/anticheat_integration_test.go`、`handlers/anticheat_test.go` 覆盖完整流程和失败降级路径。

9. 文档与运维说明
   - 在 `docs/guides/` 或 `docs/anticheat/` 添加说明，告知运营如何调整 `unban.default_message` 与 `unban.compensation_amount`。

10. 部署与监控
   - 在发布说明中记录配置项变更。
   - 添加简单监控告警（补偿失败率阈值）。

验收条件（Acceptance Criteria）:
- 管理员触发解封后用户能收到默认解封消息（或被配置覆盖）。
- 指定的默认燃素数量正确到账且在审计记录中可查。
- 在发放失败时系统记录失败状态并能由人工补发，且不会重复发放。
