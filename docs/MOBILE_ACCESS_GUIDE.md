# 📱 化学UNO - 移动端访问指�?

本指南介绍如何在手机、平板等移动设备上访问和游玩化学UNO�?

## 📋 目录

- [前提条件](#前提条件)
- [网络连接](#网络连接)
- [获取访问地址](#获取访问地址)
- [移动端访问](#移动端访�?
- [二维码快速加入](#二维码快速加�?
- [常见问题](#常见问题)
- [故障排除](#故障排除)

## �?前提条件

### 必要条件
1. **电脑已启动化学UNO服务�?*
   ```bash
   pnpm start  # �?pnpm run dev
   ```

2. **手机和电脑在同一WiFi网络**
   - 必须连接到同一个路由器
   - 不能一个用WiFi，一个用有线网络（除非在同一路由器下�?
   - 公司或学校网络可能有隔离，需要确�?

3. **防火墙允许访�?*
   - Windows防火墙需要允许Node.js
   - 端口3000�?000需要开�?

## 🌐 网络连接

### 检查网络连�?

#### 方法一：确认在同一WiFi
- 电脑和手机的WiFi名称（SSID）必须相�?
- 查看手机设置 �?WiFi �?当前连接的网络名�?
- 查看电脑设置 �?网络 �?当前连接的WiFi名称

#### 方法二：测试连通�?
```bash
# 在电脑上查看IP地址（后续步骤会用到�?
# Windows
ipconfig

# macOS/Linux
ifconfig
# �?
ip addr
```

## 🔍 获取访问地址

### Windows系统

```powershell
# 方法1：使用ipconfig
ipconfig

# 查找输出中的 "IPv4 地址"
# 无线局域网适配�?WLAN:
#   IPv4 地址 . . . . . . . . . . . . : 192.168.1.100

# 方法2：使用PowerShell脚本
(Get-NetIPAddress -AddressFamily IPv4 -InterfaceAlias "WLAN*").IPAddress
```

### macOS系统

```bash
# 方法1：使用ifconfig
ifconfig en0 | grep "inet " | awk '{print $2}'

# 方法2：使用networksetup
ipconfig getifaddr en0
```

### Linux系统

```bash
# 方法1：使用ip命令
ip addr show | grep "inet " | grep -v 127.0.0.1 | awk '{print $2}' | cut -d/ -f1

# 方法2：使用hostname命令
hostname -I | awk '{print $1}'
```

### 自动获取（推荐）

访问服务器状态页会自动显示：
```bash
# 电脑浏览器访�?
http://localhost:4001

# 会显示：
# 移动端访问地址：http://192.168.1.100:3000
```

## 📱 移动端访�?

### 步骤1：获取电脑IP地址

假设你的电脑IP地址是：`192.168.1.100`

### 步骤2：手机浏览器访问

在手机浏览器中输入：
```
http://192.168.1.100:3000
```

> 💡 提示�?
> - 注意�?`http://` 不是 `https://`
> - IP地址替换为你电脑的实际IP
> - 不要使用 `localhost` �?`127.0.0.1`

### 步骤3：加入游�?

1. **输入玩家名称**
2. **输入房间�?*（由电脑玩家提供�?
3. **点击"加入游戏"**

## 📷 二维码快速加�?

### 电脑端（房主�?

1. 创建游戏房间
2. 房间创建成功后会自动显示二维�?
3. 二维码包含：
   - 服务器地址
   - 房间�?

### 手机�?

1. 使用手机相机扫描二维�?
2. 或使用微�?浏览器扫一扫功�?
3. 自动跳转到游戏页�?
4. 输入玩家名称即可加入

## 🔥 Windows防火墙配�?

如果手机无法连接，需要配置防火墙�?

### 方法一：允许Node.js

1. 打开"Windows Defender 防火�?
2. 点击"允许应用或功能通过Windows Defender防火�?
3. 点击"更改设置"
4. 找到"Node.js"�?node.exe"
5. 勾�?专用"�?公用"
6. 点击"确定"

### 方法二：允许端口

```powershell
# 以管理员身份运行PowerShell

# 允许端口3000（前端）
netsh advfirewall firewall add rule name="ChemistryUNO-Client" dir=in action=allow protocol=TCP localport=3000

# 允许端口4001（后端）
netsh advfirewall firewall add rule name="ChemistryUNO-Server" dir=in action=allow protocol=TCP localport=4001
```

### 方法三：临时关闭防火墙（仅测试用�?

```powershell
# 关闭防火墙（不推荐用于生产环境）
netsh advfirewall set allprofiles state off

# 测试完成后记得重新开�?
netsh advfirewall set allprofiles state on
```

## 🍎 macOS防火墙配�?

```bash
# 查看防火墙状�?
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate

# 允许Node.js
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --add /usr/local/bin/node
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --unblockapp /usr/local/bin/node
```

## 🐧 Linux防火墙配�?

### UFW（Ubuntu�?

```bash
# 允许端口
sudo ufw allow 3000/tcp
sudo ufw allow 4001/tcp

# 查看状�?
sudo ufw status
```

### Firewalld（CentOS�?

```bash
# 允许端口
sudo firewall-cmd --permanent --add-port=3000/tcp
sudo firewall-cmd --permanent --add-port=4001/tcp
sudo firewall-cmd --reload

# 查看状�?
sudo firewall-cmd --list-all
```

## �?常见问题

### Q1: 手机无法连接

**可能原因**�?
1. 不在同一WiFi网络
2. 防火墙阻�?
3. IP地址错误
4. 服务未启�?

**解决步骤**�?
```bash
# 1. 确认服务正在运行
# 电脑上访�?http://localhost:3000
# 应该能看到游戏界�?

# 2. 确认防火墙设�?
# 参考上面的防火墙配置部�?

# 3. 确认网络连接
# 电脑和手机ping互相的IP

# 4. 确认IP地址正确
# 不要使用127.0.0.1或localhost
```

### Q2: 连接后无法通信

**可能原因**：WebSocket连接被阻�?

**解决方案**�?
1. 检查浏览器控制台错�?
2. 确认端口4001也已开�?
3. 尝试使用其他浏览器（Chrome、Safari�?

### Q3: 公司/学校网络无法连接

**原因**：网络有客户端隔离策�?

**解决方案**�?
1. 使用手机热点
2. 请求网络管理员开放权�?
3. 使用VPN连接

### Q4: IP地址经常变化

**解决方案**�?
```bash
# 设置静态IP（Windows�?
# 1. 打开网络设置
# 2. 更改适配器选项
# 3. WiFi属�?�?IPv4 �?使用下面的IP地址
# 4. 设置固定IP（如192.168.1.100�?

# 或使用DHCP预留（路由器设置�?
```

### Q5: 显示连接但游戏无响应

**检查项**�?
1. 查看浏览器开发者工具（F12）的Console
2. 查看Network标签的WebSocket连接状�?
3. 刷新页面重新连接

## 🔧 调试技�?

### 手机端调�?

**Android Chrome**�?
1. 电脑Chrome打开 `chrome://inspect`
2. USB连接手机并启用USB调试
3. 在inspect页面查看手机浏览�?

**iOS Safari**�?
1. 手机设置 �?Safari �?高级 �?Web检查器（开启）
2. Mac上Safari �?开�?�?[你的iPhone]
3. 选择对应页面进行调试

### 网络测试

```bash
# 电脑上测试能否访�?
curl http://localhost:3000

# 手机浏览器访问测试页
http://192.168.1.100:4001/api/mobile-info

# 应该返回服务器信�?
```

## 📊 网络拓扑示例

```
┌─────────────────────�?
�?  WiFi路由�?       �?
�?  (192.168.1.1)     �?
└──────────┬──────────�?
           �?
     ┌─────┴─────�?
     �?          �?
┌────▼────�?┌───▼────�?
�? 电脑    �?�? 手机  �?
�?.100    �?�? .101  �?
�?Server  �?�?Client �?
└─────────�?└────────�?

访问地址: http://192.168.1.100:3000
```

## �?成功连接的标�?

1. **手机浏览器显示游戏界�?*
2. **能够输入玩家名称和房间号**
3. **能够加入游戏并看到其他玩�?*
4. **能够正常出牌和接收更�?*

## 💡 最佳实�?

1. **使用稳定的WiFi网络**
   - 避免使用公共WiFi
   - 确保信号强度良好

2. **固定电脑IP地址**
   - 避免IP变化导致连接中断
   - 使用DHCP预留或静态IP

3. **关闭省电模式**
   - 手机省电模式可能影响网络连接
   - 保持屏幕常亮

4. **使用现代浏览�?*
   - Chrome、Safari、Edge
   - 避免使用过老的浏览器版�?

## 📚 相关文档

- [WebUI设置指南](WEBUI_SETUP.md) - 详细的网络配�?
- [快速开始指南](GETTING_STARTED.md) - 基本游戏流程
- [常见问题](README.md#常见问题) - 更多故障排除

---

[�?返回文档中心](README.md)
