# Docker Manager

一个轻量级的 Web 版 Docker 管理工具,支持 **Docker Compose 项目管理**。单二进制文件交付,前端界面完全嵌入后端,无需额外部署静态资源。

![Go](https://img.shields.io/badge/Go-1.27-00ADD8) ![Vue](https://img.shields.io/badge/Vue-3-42b883) ![License](https://img.shields.io/badge/License-MIT-green)

## ✨ 功能特性

| 模块 | 功能 |
|---|---|
| 🔐 认证 | 单管理员登录,Argon2id 密码哈希 + Session Cookie,登录后可修改密码 |
| 📊 总览 | 引擎/系统信息、容器统计,自动刷新(1/2/3/5/7 秒可调) |
| 🐳 容器 | 列表(复选框批量操作)、启停/重启/暂停/恢复/强制终止/删除、详情(Overview/实时日志/实时监控/Inspect) |
| 📦 项目(Compose) | **自动发现已有 compose 项目**、新建(支持 **docker run 命令一键转 compose**)、启动/停止(容器保留)/重启/重建(日志弹窗)、删除;详情含服务列表、卷信息(volume/bind 区分)、网络与端口映射(链接可点击)、YAML 查看与编辑 |
| 🖼️ 镜像 | 列表/删除、拉取实时分层进度、输入合法性校验 |
| 🌐 网络 | 批量删除(被运行中容器引用的网络自动拦截并提示)、详情页(子网/网关/连接容器 IP) |
| 💾 卷 | 列表/详情/删除(使用中提示) |
| 🌗 界面 | 暗色/亮色主题、全中文界面、WebSocket 实时数据流 |

## 🏗️ 架构

```
┌─────────────────────────────────────────────┐
│              Docker Manager (单二进制)        │
│  ┌──────────┐   ┌────────────────────────┐  │
│  │ 前端 Vue3 │◄──►│ 后端 Go (Chi + Docker  │  │
│  │ (go:embed)│   │ SDK + SQLite)          │  │
│  └──────────┘   └───────────┬────────────┘  │
│                             │                │
└─────────────────────────────┼────────────────┘
                              │ docker.sock
                    ┌─────────▼─────────┐
                    │   Docker Engine    │
                    └───────────────────┘
```

- **后端**:Go + Chi + Docker SDK v28 + modernc.org/sqlite(纯 Go,免 CGO,支持交叉编译)
- **前端**:Vue3 + TypeScript + Vite + Element Plus,构建产物嵌入二进制
- **认证**:Argon2id + SQLite Session
- **Compose 执行**:调用 `docker compose` CLI(镜像内置 docker-cli + compose 插件)

## 🚀 快速开始

### 方式一:Docker 运行(推荐)

```bash
docker run -d --name dockermanager \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /data/dockermanager:/data \
  -e DM_ADMIN_PASSWORD=你的初始密码 \
  -e DM_PROJECTS_DIR=/data/projects \
  mooloco/dockermanager:latest
```

> 支持 linux/amd64 与 linux/arm64。
> 若需读取宿主机上已有项目的 compose 文件,可挂载宿主目录(如 `-v /opt:/opt:ro`)。

### 方式二:单文件二进制(Linux x86_64)

```bash
# 从 Release 下载 docker-manager-linux-amd64
chmod +x docker-manager-linux-amd64
DM_ADMIN_PASSWORD=你的初始密码 ./docker-manager-linux-amd64
```

### 访问

浏览器打开 `http://<服务器IP>:8080`,使用 `admin` / 初始密码登录(首次登录后建议立即修改密码)。

## ⚙️ 配置

| 环境变量 | 默认值 | 说明 |
|---|---|---|
| `DM_ADMIN_PASSWORD` | 无(必填) | 初始管理员密码 |
| `SERVER_PORT` | 8080 | 监听端口 |
| `DATABASE_PATH` | ./data/docker-manager.db | SQLite 数据库路径 |
| `DM_PROJECTS_DIR` | ./data/projects | Compose 项目 YAML 存放目录 |
| `DOCKER_HOST` | unix:///var/run/docker.sock | Docker 引擎地址 |
| `AUTH_SESSION_TTL_HOURS` | 24 | 会话有效期(小时) |

## 📖 使用说明

### Compose 项目管理

- **已有项目自动识别**:通过容器标签自动发现宿主机上所有 compose 项目(含非默认文件名,如 `compose-test.yaml`)
- **新建项目**:粘贴 compose YAML 或输入 `docker run` 命令一键转换
- **按钮语义**:
  - 启动 = `docker compose up -d`(幂等,复用已停止容器)
  - 停止 = `docker compose stop`(**容器保留**,可随时启动)
  - 重建 = `compose down` → 1 秒 → `up -d`(弹窗显示日志)
  - 删除 = `compose down` + 删除本工具保存的 YAML(已有项目只 down,不碰宿主机文件)

## 🛠️ 开发

```bash
# 后端
cd backend
go build -o docker-manager ./cmd/server

# 前端
cd frontend
npm install
npm run build   # 产物嵌入后端

# 交叉编译 Linux amd64 单文件
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ../build/docker-manager-linux-amd64 ./cmd/server
```

开发过程记录见 [dev.md](dev.md);面向 AI Agent 的开发说明见 [AI_Agent-README.md](AI_Agent-README.md)。

## 📦 镜像

- `mooloco/dockermanager:latest`
- `mooloco/dockermanager:v1.3.0`(linux/amd64 + linux/arm64)

## 📄 License

MIT
