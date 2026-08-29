# AI Agent 开发指南 — Docker Manager

> 本文档写给**接手的 AI Agent** 看。它总结了项目的开发方式、用户(开发者)的习惯与要求、测试环境与发布流程。接手任何任务前请先读此文件,并遵守其中的约定。

## 一、项目是什么

Docker Manager 是一个**自研自托管的 Web 版 Docker 管理工具**,核心卖点:
- 单二进制交付(前端 go:embed 嵌入,Linux amd64 静态编译,14MB)
- 支持 **Docker Compose 项目管理**(自动发现已有项目、新建、up/stop/rebuild 等)
- 中文界面,自用为主(单管理员)

技术栈(严格固定,不要擅自更换):
- 后端:Go + Chi + Docker SDK(v28)+ modernc.org/sqlite(纯 Go,免 CGO)+ slog
- 前端:Vue3 + TypeScript + Vite + Element Plus + Pinia
- 认证:Argon2id + Session Cookie(不用 JWT)
- 项目目录:`backend/`(Go module `github.com/Mooloco/docker-manager/backend`)+ `frontend/`

## 二、用户(开发者)的习惯与要求 ⚠️ 最重要

1. **中文交流**:所有回复、文档、UI 文案用中文
2. **先讨论设计再动手**:新功能先给出方案(表格对比/API 设计/界面描述),用户确认后再开发。用户偏好"先看产物(方案文档/yaml 示例),不要急着跑演示"
3. **技术选型先讨论取舍**:如 Go vs Python、架构方案,先讲清取舍,凭实测数据决策;不要擅自选型
4. **倾向正确架构而非修补**:修补式方案反复失败时,用户倾向直接换更彻底的方案
5. **发布前不推送**:开发、验证满意后才推 GitHub/Docker Hub;推送前先经用户确认(用户曾明确叫停过推送)
6. **版本规则**:`V1.主版本.次版本` — 新功能 → 主版本位 +1;修 Bug → 次版本位 +1。记录在 `dev.md`
7. **UI 交互偏好**:批量操作统一用"复选框 + 第二行操作栏"模式;操作按钮带文字(不只图标);点击名称进详情;删除等危险操作有确认框
8. **部署形态决策**:
   - 交付 = Docker 镜像 + **deb 安装包**(二进制 + conf + systemd)+ Linux amd64 单文件
   - 用户自用工具,**Web Terminal 功能不开发**(保留入口占位即可)
   - 容器部署下"已有项目 YAML 读取"受文件系统隔离限制,用户已确认保持现状(deb 部署自然可读)

## 三、测试环境与流程

| 环境 | 用途 |
|---|---|
| **192.168.1.24**(Ubuntu,root 免密 SSH) | Linux 功能测试机:docker build + 容器运行(dm-test:18080,admin/test1234)、deb 安装测试 |
| **192.168.1.1**(iStoreOS 主路由,root 密码 JDunix786) | 只负责把容器跑起来验证,**不动其设置**(有需要先问) |

标准开发循环:
1. 本地改代码 → 前端 `npm run build` → 后端 `go build ./...` 验证
2. 打包同步:`tar czf - --exclude=node_modules --exclude=dist ... | ssh root@192.168.1.24 "mkdir -p /mnt/scsi1/docker-manager && cd ... && find . -mindepth 1 -maxdepth 1 ! -name data -exec rm -rf {} + && tar xzf -"`
   - ⚠️ **不要 `rm -rf /opt/docker-manager`**:它是软链接指向 `/mnt/scsi1/docker-manager`,rm 会删软链接导致数据回系统盘
3. 测试机 `docker build -t docker-manager:dev . && docker rm -f dm-test; docker run -d --name dm-test ...`
4. 功能验证:Playwright 无头浏览器(Edge channel)全流程检查 + API curl/脚本验证

**测试机注意**:
- 系统盘 `/` 只有 20GB,易满;构建前先 `docker builder prune -f`
- 数据/项目放新盘 `/mnt/scsi1`(软链接 `/opt/docker-manager`)

## 四、常见坑(踩过的)

1. **`fs.Sub` + `http.FileServer` 的路径怪癖**:`http.FileServer` 传带前导斜杠路径给 `fs.Sub` 的 FS 会失败(301)。SPA 静态服务**手写**(`fs.ReadFile` + 手动 Content-Type),不要用 FileServer
2. **WebSocket 升级 500**:日志中间件的 ResponseWriter 包装器必须实现 `http.Hijacker`,否则 gorilla Upgrade 的 Hijack 断言失败
3. **WSClient 自动重连**:收到 `end`/`error` 消息必须主动 `close()`,否则 3 秒后自动重连造成"重复拉取/重复日志"
4. **Element Plus**:`el-tooltip` 包 `el-dropdown` 会导致下拉项渲染但不可见(popper 冲突),直接裸用 dropdown
5. **表格 fixed="right" 列遮挡**:列总宽超视口时 fixed 列覆盖前面列;压缩列宽或去掉 fixed
6. **复选框勾选刷新丢失**:表格 `row-key` + 选择列 `reserve-selection` 按 ID 保留
7. **chi 空 Route 分组**:分组内无子路由时中间件不生效(404);Docker 不可用降级需加占位通配路由
8. **Docker SDK v28 类型**:以模块缓存源码为准(如 `StopOptions.Timeout` 是 `*int`、`mount.Type` 需转 string)
9. **同步目录**:**直接同步到 `/mnt/scsi1/docker-manager`**(真实路径),不要动 `/opt/docker-manager` 软链接
10. **Playwright 路径文本**:`text=/mnt/bind` 会被当正则,用 `td:has-text(...)` 或普通字符串

## 五、代码约定

- 后端 API 统一响应:`{success, data, error:{code, message}}`;Docker 错误统一转友好中文提示
- 项目名等输入白名单校验(防路径穿越);exec 参数走数组(防注入)
- 前端页面风格:CSS 变量 `--dm-sidebar-width: 220px`、`--dm-header-height: 56px`;暗色/亮色主题
- 单测:auth/config/database/docker/compose 包均有 Go 测试(改这些包要跑 `go test ./...`)
- 版本号同步:`dev.md` 更新 + 镜像 tag + deb 版本号保持一致

## 六、发布流程

1. 更新 `dev.md`(新版本记录)
2. 前端 build → 交叉编译 Linux amd64 → 单文件验证
3. 构建多平台镜像并推送 Docker Hub(`mooloco/dockermanager:latest` + 版本 tag)
4. 生成 deb 包并在测试机安装验证(systemd 服务)
5. git commit + push GitHub(`Mooloco/DockerManager`)
6. **所有推送先经用户确认**
