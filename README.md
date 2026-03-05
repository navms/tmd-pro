# TMD Pro

> 基于 [tmd](https://github.com/unkmonster/tmd) 打造的 X 媒体批量下载桌面管理工具

TMD Pro 是一个 macOS 桌面端应用，为 [tmd](https://github.com/unkmonster/tmd) 提供可视化的操作界面。你可以通过图形界面管理待下载的 X 用户列表，启动/停止定时轮转扫描，并实时查看执行日志，无需手动敲命令行。

---

## 功能特性

- **用户管理**：可视化增删 screen_name，支持搜索过滤、批量添加（逗号/换行分隔）
- **定时轮转扫描**：按配置的时间间隔自动轮转处理所有用户，支持断点续扫（游标机制）
- **单次执行**：手动触发一次扫描，处理当前批次用户
- **实时日志**：终端风格日志控制台，成功/失败/系统消息自动着色
- **完整配置界面**：扫描间隔、代理、数据库连接等所有配置项均可在界面内修改保存
- **CLI 模式**：保留完整命令行能力，GUI 和 CLI 共享同一个二进制

---

## 技术栈


| 层次   | 技术                                              |
| ---- | ----------------------------------------------- |
| 桌面框架 | [Wails v2](https://wails.io)                    |
| 前端   | React 18 + TypeScript + Tailwind CSS            |
| 后端   | Go 1.24                                         |
| 数据库  | MySQL（GORM）                                     |
| 配置   | Viper（YAML）                                     |
| 日志   | Zap + Lumberjack                                |
| 核心工具 | [tmd](https://github.com/unkmonster/tmd)（内嵌二进制） |


---

## 前置要求

- Go 1.21+
- Node.js 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)
- MySQL 数据库

安装 Wails CLI：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

---

## 配置文件

首次运行前，在 `~/.tmd-pro/conf/app.yaml` 创建配置文件：

```yaml
app:
  name: tmd-pro
  env: prod
  data_dir: ~/.tmd-pro
  tmd_binary_name: tmd
  scan_interval_minutes: 30
  proxy:
    http_proxy: "http://127.0.0.1:7890"
    https_proxy: "http://127.0.0.1:7890"
    no_proxy: "localhost,127.0.0.1"

storage:
  host: 127.0.0.1
  port: 3306
  username: root
  password: your_password
  database: tmd_pro
  charset: utf8mb4

logger:
  env: prod
  level: info
  file-name: ~/.tmd-pro/logs/app.log
  max-size: 100
  max-backups: 10
  max-age: 30
  compress: true
  show-caller: true
  output-console: true
```

> 所有配置项也可以在应用启动后通过「系统配置」界面直接修改保存，无需手动编辑文件。

---

## 构建 & 运行

### 开发模式（热更新）

```bash
wails dev
```

### 生产构建

```bash
wails build
```

构建产物在 `build/bin/` 目录下。

---

## CLI 命令

除 GUI 外，同一个二进制也支持命令行操作：

```bash
# 默认启动 GUI
./tmd-pro

# 显式启动 GUI
./tmd-pro gui

# 在后台启动定时轮转扫描（无 GUI）
./tmd-pro rotate

# 添加用户
./tmd-pro add elonmusk sama

# 删除用户
./tmd-pro remove elonmusk
```

---

## 相关项目

- **[tmd](https://github.com/unkmonster/tmd)**：本项目的核心下载引擎，跨平台 X 媒体下载 CLI 工具

---

## License

MIT