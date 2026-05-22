# BD-UI

Beads UI 的 Go 版本实现 —— 一个用于 [Beads](https://github.com/steveyegge/beads) issue 跟踪系统的本地 Web 界面。

基于原版 [beads-ui](https://github.com/mantoni/beads-ui)（Node.js）重写，后端使用 Go，前端使用 Vue 3 + Element Plus。

## 功能

- ✅ **零配置** — 运行 `bd-ui` 即可启动
- 📺 **实时更新** — 监控 Beads 数据库变化，自动刷新
- 🔎 **问题视图** — 筛选、搜索、内联编辑
- ⛰️ **需求视图** — 展示需求进度、子任务统计
- 🏂 **看板视图** — 阻塞 / 就绪 / 进行中 / 已关闭 四列看板
- 🔀 **多工作空间** — 下拉切换项目，自动注册工作空间
- 🌐 **国际化** — 中英文切换，Element Plus 组件联动
- 🌙 **暗色模式** — 一键切换亮/暗主题
- 📦 **单文件部署** — 前端嵌入 Go 二进制，无需额外文件

## 快速开始

### 从源码构建

```bash
# 构建
git clone <repo-url> && cd bd-ui
cd web && npm install && npm run build && cd ..
go build -o bd-ui.exe .

# 运行（在 Beads 项目目录下）
bd-ui.exe --dir /path/to/your/beads/project
```

### 运行方式

```bash
# 方式 1：指定项目目录（推荐）
bd-ui.exe --dir "D:\my-project"

# 方式 2：在项目目录下运行
cd D:\my-project
bd-ui.exe

# 方式 3：指定数据库路径
set BEADS_DB=D:\my-project\.beads\project.db
bd-ui.exe
```

### 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--dir` | Beads 项目目录 | 当前目录 |
| `--host` | 监听地址 | 127.0.0.1 |
| `--port` | 监听端口 | 3000 |
| `--open` | 启动后打开浏览器 | false |

### 环境变量

| 变量 | 说明 |
|------|------|
| `BD_BIN` | `bd` 可执行文件路径（默认从 PATH 查找） |
| `BEADS_DB` | 数据库路径（默认自动查找 `.beads/*.db`） |
| `HOST` | 监听地址 |
| `PORT` | 监听端口 |

## 项目结构

```
bd-ui/
├── main.go                  # 入口：CLI 参数 + 嵌入前端 + 启动服务
├── server/
│   ├── config.go            # 配置（host/port/rootDir）
│   ├── db.go                # 数据库路径解析（.beads/*.db）
│   ├── bd.go                # bd CLI 调用（带队列串行化）
│   ├── server.go            # HTTP 服务器 + 静态文件 + API
│   ├── ws.go                # WebSocket 处理器（全部消息类型）
│   ├── subscriptions.go     # 订阅注册表 + 增量推送
│   ├── list_adapters.go     # 订阅类型 → bd 命令映射
│   ├── watcher.go           # 数据库文件监视（fsnotify）
│   └── registry.go          # 工作空间注册表
├── web/                     # Vue 3 前端源码
│   ├── src/
│   │   ├── App.vue          # 主布局
│   │   ├── router.js        # 哈希路由
│   │   ├── composables/
│   │   │   └── useWs.js     # WebSocket 客户端
│   │   ├── stores/
│   │   │   ├── issues.js    # 问题状态管理
│   │   │   └── workspace.js # 工作空间管理
│   │   ├── locales/
│   │   │   ├── zh-CN.js     # 中文语言包
│   │   │   └── en.js        # 英文语言包
│   │   ├── components/
│   │   │   ├── NewIssueDialog.vue
│   │   │   └── IssueDetail.vue
│   │   └── views/
│   │       ├── IssuesView.vue
│   │       ├── EpicsView.vue
│   │       └── BoardView.vue
│   └── vite.config.js
└── web-dist/                # 前端构建产物（go:embed 嵌入）
```

## 技术栈

**后端：**
- Go 1.x
- [gorilla/websocket](https://github.com/gorilla/websocket) — WebSocket 通信
- [fsnotify](https://github.com/fsnotify/fsnotify) — 文件系统监控
- `os/exec` — 调用 `bd` CLI

**前端：**
- Vue 3 + Composition API
- Element Plus — UI 组件库
- Pinia — 状态管理
- Vue Router — 哈希路由
- vue-i18n — 国际化
- Vite — 构建工具

## 开发

```bash
# 前端开发（热更新）
cd web && npm run dev         # http://localhost:5173，自动代理到后端

# 后端开发
go run . --dir /path/to/project --port 3000

# 构建生产版本
cd web && npm run build && cd ..
go build -o bd-ui.exe .
```

## License

Apache-2.0
