# 后端编译问题解决方案

## 问题描述

如果你看到类似以下的错误：
```
ld.exe: unrecognized option '--high-entropy-va'
```

这是因为系统安装的 MinGW/GCC 版本过旧（4.9.x），不支持现代Go编译器需要的链接选项。

## 解决方案

### 方案1：更新MinGW（推荐）

下载并安装较新的MinGW-w64：
- 下载地址：https://github.com/niXman/mingw-builds-binaries/releases
- 推荐版本：MinGW-w64 GCC 8.1.0 或更高
- 安装后，将新的MinGW bin目录添加到系统PATH的最前面

### 方案2：使用TDM-GCC

TDM-GCC是一个现代化的MinGW发行版：
- 下载地址：https://jmeubank.github.io/tdm-gcc/
- 选择TDM-GCC 10.3.0 或更高版本
- 安装后会自动配置PATH

### 方案3：使用预编译的Go SQLite驱动（不推荐）

如果无法更新GCC，可以考虑将数据库从SQLite迁移到不需要CGO的数据库，如使用纯Go实现的数据库。

## 验证

安装新的GCC后，运行：
```bash
gcc --version
```

应该显示版本 >= 8.0

然后再次尝试启动：
```bash
cd backend
go run main.go
```

## 临时方案

在安装新GCC之前，你可以：
1. 从其他机器编译好的二进制文件
2. 使用Docker容器运行后端
3. 在Linux或Mac系统上开发
