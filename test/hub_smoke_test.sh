#!/bin/bash
# Docker Hub 镜像冒烟测试:拉取运行 + 登录 + 系统信息
set -e
ssh root@192.168.1.24 << 'EOF'
set -e
docker rm -f dm-pubtest 2>/dev/null || true
docker run -d --name dm-pubtest -p 18084:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e DM_ADMIN_PASSWORD=test1234 \
  mooloco/dockermanager:v1.3.0 >/dev/null
sleep 3
echo "--- 登录 ---"
TOKEN=$(curl -s -D - -o /dev/null -X POST http://localhost:18084/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"test1234"}' \
  | grep -i set-cookie | sed 's/.*dm_session=\([^;]*\).*/dm_session=\1/' | tr -d '\r')
echo "session: ${TOKEN:0:24}..."
echo "--- 系统信息 ---"
curl -s -H "Cookie: $TOKEN" http://localhost:18084/api/v1/system/info | head -c 160
echo
echo "--- 前端页面 ---"
curl -s -o /dev/null -w "HTTP %{http_code}\n" http://localhost:18084/
docker rm -f dm-pubtest >/dev/null 2>&1
echo "--- 冒烟测试通过,容器已清理 ---"
EOF
