#!/bin/bash
# 更新 Docker Hub repo overview(description + full_description)
set -e
ssh root@192.168.1.24 << 'EOF'
set -e

AUTH=$(python3 -c "
import json, base64
cfg = json.load(open('/root/.docker/config.json'))
a = cfg['auths'].get('https://index.docker.io/v1/', cfg['auths'].get('docker.io', {}))
print(base64.b64decode(a['auth']).decode() if 'auth' in a else a.get('username','') + ':' + a.get('password',''))
")
USER=$(echo "$AUTH" | cut -d: -f1)
PASS=$(echo "$AUTH" | cut -d: -f2-)
echo "登录用户: $USER"

JWT=$(curl -s -X POST https://hub.docker.com/v2/users/login/ \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USER\",\"password\":\"$PASS\"}" | python3 -c "import json,sys; print(json.load(sys.stdin)['token'])")
echo "JWT 获取成功: ${JWT:0:16}..."

# 更新 description + full_description
curl -s -X PATCH https://hub.docker.com/v2/repositories/mooloco/dockermanager/ \
  -H "Authorization: JWT $JWT" \
  -H "Content-Type: application/json" \
  -d @- << 'JSON' | head -c 200
{
  "description": "Docker 管理工具:容器/Compose 项目/镜像/网络/卷,中文界面,单二进制交付",
  "full_description": "# Docker Manager\n\n轻量级 Web 版 Docker 管理工具(Go + Vue3 + Element Plus),单二进制交付,全中文界面。\n\n## 功能\n- 容器管理:批量操作、实时日志、实时监控(WebSocket)\n- **Docker Compose 项目管理**:自动发现已有项目、新建、启动/停止/重建、docker run 命令一键转 compose\n- 镜像管理:拉取实时进度、批量删除\n- 网络管理:批量删除(运行中容器引用拦截)、详情页\n- 卷管理、暗色主题、刷新频率可调\n\n## 快速开始\n```bash\ndocker run -d --name dockermanager \\\n  -p 8080:8080 \\\n  -v /var/run/docker.sock:/var/run/docker.sock \\\n  -v /data/dockermanager:/data \\\n  -e DM_ADMIN_PASSWORD=你的初始密码 \\\n  mooloco/dockermanager:latest\n```\n\n支持 linux/amd64 与 linux/arm64。\n源码:https://github.com/Mooloco/DockerManager"
}
JSON
echo
echo "--- 验证 ---"
curl -s https://hub.docker.com/v2/repositories/mooloco/dockermanager/ | python3 -c "
import json, sys
d = json.load(sys.stdin)
print('description:', d['description'])
print('full_description 前 100 字:', d['full_description'][:100].replace(chr(10),' '))
"
EOF
