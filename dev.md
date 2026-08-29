# Docker Manager 开发记录 (dev.md)

> 本文件记录 Docker Manager 的开发演进过程。
>
> **版本规则**:`V1.主版本.次版本`
> - **第 2 位(主版本位)**:每次新增功能 +1(如 V1.0.0 → V1.1.0)
> - **第 3 位(次版本位)**:每次修复 Bug +1(如 V1.1.0 → V1.1.1)
>
> 当前版本:**V1.3.0**(V1 大版本完成)

---

## V1.0.0 — MVP 初始版本

**技术栈**:Go 1.27 + Chi + Docker SDK v28 + modernc.org/sqlite(纯 Go,交叉编译免 CGO)+ slog;Vue3 + TS + Vite + Element Plus;前端产物 go:embed 进单个二进制。

**功能**:
- 登录认证:Argon2id 密码哈希 + Session Cookie,初始密码由环境变量 `DM_ADMIN_PASSWORD` 指定,登录后可在设置页修改
- 总览:引擎/系统信息/容器统计
- 容器:列表(实时 CPU/内存)、启停/重启/暂停/恢复/强制终止/删除、详情(Overview/Inspect 分组 + Raw JSON)
- 日志:WebSocket 实时流(stdout/stderr 分离、跟随、行数控制)
- 实时监控:WebSocket 推送 CPU/内存/网络/磁盘/PIDs
- 镜像:列表/删除、拉取实时分层进度(WebSocket)
- 网络/卷:列表/详情/删除
- Web Terminal:**占位入口**(不开发,提示"终端功能未开放")
- 交付形态:Linux amd64 静态单文件(14MB,前端嵌入);Dockerfile 多阶段构建 + compose
- 验证:192.168.1.24 测试机容器化全功能验证、主路由 192.168.1.1 容器运行验证(22 个真实容器管理正常)

---

## V1.0.1 — Bug 修复

- **修复**镜像拉取完成后页面无限提示"镜像拉取完成":WebSocket 客户端"意外断开 3 秒自动重连"逻辑导致收到结束消息后重新拉取;收到 `end`/`error` 后主动关闭连接(日志流同类问题一并修复)
- **修复**拉取对话框缺少"完成"状态按钮(增加第三态:拉取中/完成/拉取)
- **修复**容器表格操作列 `fixed="right"` 遮挡创建时间列:去掉 fixed + 压缩列宽

---

## V1.1.0 — 容器页重构与体验增强(新功能)

- 容器页仿 **OpenWrt Dockerman**:表格复选框批量选择 + 工具栏带文字操作按钮(启动/停止/重启/暂停/恢复/强制终止/删除),行内操作移除
- 镜像拉取输入校验:非法输入提示"输入错误,请重新输入"(按 Docker 官方镜像引用规范)
- 左侧菜单 Dashboard → **总览**
- 顶栏新增**刷新频率按钮**(1/2/3/5/7 秒,持久化),总览页自动刷新跟随

---

## V1.1.1 — 交互修复

- **修复**自动刷新导致复选框勾选丢失:表格 `row-key` + 选择列 `reserve-selection`,按 ID 保留选中;批量操作后自动清空
- 容器操作按钮移至工具栏第二行(第一行搜索/过滤,第二行批量操作)
- 删除容器列表 CPU/内存列(资源占用看详情页)
- 刷新机制分层:容器页自动刷新**固定 1 分钟**(不跟顶栏,避免打断勾选);详情页实时监控跟随顶栏频率;总览页跟随顶栏

---

## V1.2.0 — Compose 项目管理(新功能)

- 左侧菜单新增**项目**(容器与镜像之间),标题标注"项目 · Docker Compose"
- **自动发现已有 compose 项目**:扫描容器 `com.docker.compose.project` 标签分组还原(测试机 guacamole 等已有项目自动识别)
- 新建项目:compose YAML 落盘(`DM_PROJECTS_DIR`,相对路径语义正确)、编辑、删除(删除保护:已有项目不可删)
- 项目操作:启动(`up -d`)/停止(`down`)/重启,执行引擎为 `docker compose` CLI
- 项目详情:服务列表(容器级启停、进容器详情)、**卷信息**(volume/bind 区分、宿主机位置、读写)、**compose 文件名**(非默认文件名可识别)、YAML 查看/编辑
- 镜像内置 `docker-cli` + `docker-cli-compose`(容器部署下可执行 compose)

---

## V1.2.1 — 按钮语义修正(Bug 修复)

- **修复**停止按钮语义:`docker compose down`(删容器)→ **`docker compose stop`**(停止容器但保留,可随时启动)
- **修复**删除按钮语义:删除 = `compose down`(删容器/网络)+ 删本工具文件;已有项目只执行 down,**不碰宿主机 compose 文件**
- **修复**删除确认文案自相矛盾(如实描述 down 行为)
- **修复**新建未启动项目的启动按钮灰色:`has_containers=false` 被禁用条件误判,改为只看 `running > 0`

---

## V1.4.0 — OpenWrt x86_64 支持(procd + LuCI 入口)

- **二进制复用**:Linux amd64 静态编译二进制直接在 OpenWrt(iStoreOS)运行,零改动
- **procd init 脚本**:/etc/init.d/dockermanager,UCI 配置(enabled/port/bind_addr/data_dir/projects_dir/admin_password)
- **ipk 打包**:现代 OpenWrt 格式(整包 gzip tar,内含 debian-binary + control.tar.gz + data.tar.gz),替换路由上 iStoreOS 自带 dockermanager
- **LuCI 应用 luci-app-dockermanager**:菜单入口 + 跳转管理界面(纯静态,无 RPC)
- 验证:路由 192.168.1.1(iStoreOS 24.10.4 / Docker 27.3.1)自动发现 12 个 compose 项目

**踩坑记录**(详见 AI_Agent-README.md):
- 现代 OpenWrt ipk = gzip 压缩 tar(非三段拼接)
- procd_set_param env 多次调用互相覆盖,须一次传全部
- rc.common config_get 在该版本异常(CONFIG 变量为空),init 脚本直接 uci 命令读取
- rpcd ucode 返回格式必须 `{ '对象名': { 方法: {call} } }`;ucode 用 import 语法
- LuCI 菜单 depends.acl 需对应 ACL 文件存在,否则页面 4xx

## V1.3.0 — 项目增强 + 网络管理

- **重建按钮**(列表页 + 详情页):`compose down` → 1 秒延迟 → `compose up -d`,弹窗展示完整日志
- 项目操作按钮移至工具栏第二行
- 项目详情**网络/端口卡片**:网络(名称/driver/internal/容器 IP)+ 端口映射链接化(`192.168.1.24:8081 → 172.18.0.2:80`,两端可点击打开);宿主机 0.0.0.0 自动用当前访问地址;**IPv4/IPv6 双绑定去重**
- **docker run 命令转 compose**:新建项目支持粘贴 `docker run` 命令一键转换(纯 Go 解析,支持引号/转义/端口/卷/环境变量/重启策略/TTY/特权等,8 个单元测试)
- 项目页复选框多选(替代行选中歧义,与容器页一致)
- **网络管理**:复选框批量删除(被运行中容器引用的网络拒绝删除并提示具体容器)、点击名称进网络详情页(基本信息 + 连接容器 IP/MAC)

---

## V1.3.0 之后(发布)

- 多平台镜像:`mooloco/dockermanager:latest` + `v1.3.0`(linux/amd64 + linux/arm64)
- 源码发布:GitHub `Mooloco/DockerManager`
