# 🔒 安全配置指南

## ⚠️ 重要提示

本项目包含敏感配置信息，请务必遵循以下安全实践�?

## 🔑 敏感信息清单

### 1. 管理员密�?

**位置**: `client/.env.production`

**配置�?*: `REACT_APP_ADMIN`

**默认�?*: 无（需要手动设置）

**如何设置**:

```bash
# 1. 复制示例文件
cp client/.env.example client/.env.production

# 2. 编辑 client/.env.production
REACT_APP_ADMIN=your_strong_password_here

# 3. 重新构建前端
pnpm run build:client
```

**密码要求**:
- 至少 12 个字�?
- 包含大小写字母、数字和特殊字符
- 不要使用常见密码
- 不要在代码或文档中硬编码密码

### 2. 环境变量文件

**已添加到 .gitignore**:
- `.env`
- `.env.local`
- `.env.*.local`
- `.env.production`
- `client/.env.production`

**示例文件** (可提交到版本控制):
- `client/.env.example`

## 🛡�?安全最佳实�?

### 首次部署

1. **设置强密�?*
   ```bash
   # 编辑 client/.env.production
   REACT_APP_ADMIN=YourStr0ng!P@ssw0rd
   ```

2. **验证 .gitignore**
   ```bash
   # 确保敏感文件不会被提�?
   git status
   ```

3. **构建应用**
   ```bash
   pnpm run build
   ```

### 生产环境

1. **更改默认端口**（可选）
   - 前端: 修改 `client/.env.production` 中的 `PORT`
   - 后端: 修改 `server/index.ts` 中的端口配置

2. **使用环境变量**
   ```bash
   # 通过环境变量传递敏感信�?
   export REACT_APP_ADMIN="your_password"
   pnpm run build
   ```

3. **定期更换密码**
   - 建议�?3-6 个月更换一次管理员密码

4. **限制管理面板访问**
   - 仅在需要时访问 `/admin`
   - 考虑添加 IP 白名�?

### 团队协作

1. **不要提交敏感信息**
   - 检查提交内�? `git diff`
   - 使用 `.gitignore` 排除敏感文件

2. **使用密钥管理工具**
   - 密码管理器（�?1Password, LastPass�?
   - 环境变量管理服务

3. **文档中使用占位符**
   - �?`REACT_APP_ADMIN=your_password_here`
   - ❌ 不要在文档中使用真实密码示例

## 📋 检查清�?

部署前请检查：

- [ ] 已设置强管理员密�?
- [ ] `.env.production` 文件已添加到 `.gitignore`
- [ ] 代码中没有硬编码的密�?
- [ ] 文档中使用占位符而非实际密码
- [ ] 已测试管理面板登录功�?

## 🔍 敏感信息扫描

运行以下命令检查代码中是否有遗留的敏感信息�?

```bash
# 搜索可能的密�?
grep -r "password.*=" . --include="*.js" --include="*.ts" --include="*.tsx"

# 搜索硬编码的密钥
grep -r "REACT_APP_ADMIN.*=" . --include="*.js" --include="*.ts" --include="*.tsx"

# 检�?git 历史
git log -p | grep -i "password"
```

## 🚨 如果密码泄露

1. **立即更改密码**
   ```bash
   # 编辑 client/.env.production
   REACT_APP_ADMIN=NewStr0ng!P@ssw0rd
   
   # 重新构建
   pnpm run build:client
   ```

2. **检查访问日�?*
   - 查看是否有未授权访问

3. **通知团队成员**
   - 确保所有人使用新密�?

## 📚 相关文档

- [快速开始](GETTING_STARTED.md) - 基本配置
- [使用说明](使用说明.md) - 日常使用
- [管理面板指南](ADMIN_PANEL_GUIDE.md) - 管理功能

## 💡 技术说�?

### 密码存储

- 管理员密码通过环境变量传递到前端
- 密码在客户端使用 `sessionStorage` 临时存储
- 关闭浏览器后自动清除

### 未来改进

- [ ] 实现后端验证（而非前端验证�?
- [ ] 添加密码哈希存储
- [ ] 实现 JWT 令牌认证
- [ ] 添加用户管理系统
- [ ] 实现操作审计日志

---

**记住**: 安全是持续的过程，不是一次性任务！🔐
