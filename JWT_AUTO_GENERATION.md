# JWT密钥自动生成功能说明

## 功能概述

系统现在支持在首次启动时自动生成50位的强随机JWT密钥，无需手动配置。

## 工作原理

### 1. 启动时检查

应用启动时会自动调用 `utils.EnsureJWTSecret()` 函数，执行以下检查：

1. **检查环境变量**: 首先检查 `JWT_SECRET` 环境变量是否已设置
2. **检查.env文件**: 如果环境变量未设置，检查 `.env` 文件
3. **自动生成**: 如果两者都不存在，自动生成50位随机密钥
4. **保存配置**: 将生成的密钥保存到 `.env` 文件

### 2. 密钥生成

- 使用 `crypto/rand` 生成密码学安全的随机字节
- 通过 Base64 URL编码生成50位字符串
- 包含字母、数字、特殊字符，安全性高

### 3. 文件处理

- 如果 `.env` 文件不存在，自动从 `.env.example` 复制
- 智能替换 `JWT_SECRET=` 行，保留其他配置
- 如果文件中没有JWT_SECRET配置，自动追加到文件末尾

## 使用场景

### 场景1: 全新安装

```bash
# 首次运行
./main.exe

# 输出：
# ✓ 已自动生成并保存50位JWT密钥到 .env
# ⚠ 重要提示: 请妥善保管 .env 文件
```

### 场景2: .env存在但无JWT_SECRET

系统会自动在现有 `.env` 文件中添加JWT_SECRET配置。

### 场景3: JWT_SECRET已配置

系统检测到已有配置，跳过生成：

```
JWT_SECRET 已配置
```

## 安全特性

### 1. 密码学安全

- ✅ 使用 `crypto/rand` 而非 `math/rand`
- ✅ 50位长度，远超最低安全要求（32字符）
- ✅ 包含大小写字母、数字、特殊字符

### 2. 文件保护

- ✅ `.env` 文件已在 `.gitignore` 中，不会被提交
- ✅ 启动时输出安全提示
- ✅ 文件权限设置为 0644

### 3. 环境变量优先级

```
环境变量 > .env文件 > 自动生成
```

## 示例密钥

生成的密钥示例：

```
JWT_SECRET=pG-JKD9k1z4dfSDD8puJGp6m9RiLG0KNVuGlhvbcfklxDI9cp8
```

特点：

- 长度: 50个字符
- 字符集: `[A-Za-z0-9_-]`
- 熵: ~298 bits (非常安全)

## 日志输出

### 首次生成时

```
2026/02/03 21:00:14 已从 .env.example 创建 .env 文件
2026/02/03 21:00:14 ✓ 已自动生成并保存50位JWT密钥到 D:\...\backend\.env
2026/02/03 21:00:14 ⚠ 重要提示: 请妥善保管 .env 文件，不要将其提交到版本控制系统
```

### 已有配置时

```
2026/02/03 21:00:14 JWT_SECRET 已配置
```

## 手动配置（可选）

如果需要使用特定密钥，可以手动设置：

### 方法1: 修改.env文件

```dotenv
JWT_SECRET=your-custom-50-character-secret-key-here-12345678
```

### 方法2: 设置环境变量

```bash
# Windows PowerShell
$env:JWT_SECRET="your-custom-secret"

# Linux/Mac
export JWT_SECRET="your-custom-secret"
```

## 密钥要求

手动配置时建议遵循：

- ✅ 最少32个字符
- ✅ 建议50个字符或以上
- ✅ 包含随机字符
- ✅ 避免使用常见词汇

## 生成自定义密钥

### PowerShell命令

```powershell
-join ((65..90) + (97..122) + (48..57) | Get-Random -Count 50 | % {[char]$_})
```

### Linux/Mac命令

```bash
openssl rand -base64 48 | cut -c1-50
```

### Python命令

```python
import secrets
print(secrets.token_urlsafe(50)[:50])
```

## 部署建议

### 开发环境

- ✅ 使用自动生成功能
- ✅ `.env` 文件存放在本地
- ✅ 不要提交 `.env` 到Git

### 生产环境

建议使用以下方式之一：

#### 1. 环境变量（推荐）

```bash
export JWT_SECRET="production-secret-key-50-chars"
./main.exe
```

#### 2. 密钥管理服务

- AWS Secrets Manager
- Azure Key Vault  
- HashiCorp Vault

#### 3. 配置文件

```bash
# 使用独立的配置文件
./main.exe --config /secure/path/config.env
```

## 常见问题

### Q: 密钥会被覆盖吗？

A: 不会。只有在检测不到JWT_SECRET时才会生成新密钥。

### Q: 可以修改生成的密钥吗？

A: 可以。生成后可以手动编辑 `.env` 文件修改密钥。

### Q: 密钥长度可以改变吗？

A: 可以。修改 `GenerateRandomSecret(50)` 中的数字即可。建议不少于32。

### Q: 如何重新生成密钥？

A: 删除 `.env` 文件中的JWT_SECRET行或整个文件，重启应用即可。

### Q: 多个实例需要相同密钥吗？

A: 是的。如果有多个后端实例，它们必须使用相同的JWT_SECRET才能互相验证token。

## 安全提醒

⚠️ **重要安全提示:**

1. **不要提交密钥**: `.env` 文件必须在 `.gitignore` 中
2. **不要共享密钥**: 每个环境使用独立密钥
3. **定期轮换**: 生产环境建议定期更换密钥
4. **安全存储**: 使用密钥管理服务存储生产密钥
5. **监控日志**: 检查是否有"使用默认密钥"警告

## 技术实现

### 相关文件

- `backend/utils/secret_generator.go` - 密钥生成逻辑
- `backend/utils/jwt.go` - JWT初始化
- `backend/main.go` - 启动时调用
- `backend/.env.example` - 配置模板
- `backend/.env` - 实际配置（自动生成）

### 核心函数

```go
// 生成指定长度的随机密钥
func GenerateRandomSecret(length int) (string, error)

// 确保JWT密钥存在
func EnsureJWTSecret() error
```

## 更新日志

**2026-02-03**

- ✅ 实现自动生成50位JWT密钥功能
- ✅ 支持从 `.env.example` 自动创建 `.env`
- ✅ 智能更新现有配置文件
- ✅ 添加详细的日志输出
- ✅ 集成到应用启动流程
