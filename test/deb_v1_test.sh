#!/bin/bash
# 测试机 deb 安装/运行/功能/卸载 全流程
set -e
DEB=/tmp/docker-manager_1.3.0_amd64.deb
scp /c/Users/ms/Desktop/Hermes/GitHub/docker-manager/build/docker-manager_1.3.0_amd64.deb root@192.168.1.24:$DEB

ssh root@192.168.1.24 << 'EOF'
set -e
echo "=== [1] 卸载旧包 + 安装 ==="
dpkg -P docker-manager >/dev/null 2>&1 || true
userdel docker-manager >/dev/null 2>&1 || true
rm -rf /etc/docker-manager /var/lib/docker-manager
dpkg -i /tmp/docker-manager_1.3.0_amd64.deb
which docker-manager && ls -la /usr/bin/docker-manager | awk '{print $5}'

echo "=== [2] 配置并启动 ==="
PORT=18082
echo "DM_ADMIN_PASSWORD=test1234" > /etc/docker-manager/env
echo "SERVER_PORT=$PORT" >> /etc/docker-manager/env
chmod 600 /etc/docker-manager/env
systemctl start docker-manager
sleep 2
systemctl status docker-manager --no-pager | head -3
ss -tlnp | grep $PORT | head -1

echo "=== [3] API 验证 ==="
TOKEN=$(curl -s -D - -o /dev/null -X POST http://localhost:$PORT/api/v1/auth/login -H "Content-Type: application/json" -d '{"username":"admin","password":"test1234"}' | grep -i set-cookie | sed 's/.*dm_session=\([^;]*\).*/dm_session=\1/' | tr -d '\r')
echo "登录: ${TOKEN:0:20}..."
echo "--- 系统信息 ---"
curl -s -H "Cookie: $TOKEN" http://localhost:$PORT/api/v1/system/info | head -c 200
echo
echo "--- 项目列表(deb 部署应有已有项目) ---"
curl -s -H "Cookie: $TOKEN" http://localhost:$PORT/api/v1/projects | head -c 300
echo
echo "--- 已有项目 YAML 读取(deb 部署下应可读!) ---"
curl -s -H "Cookie: $TOKEN" http://localhost:$PORT/api/v1/projects/guacamole/yaml | head -c 300
echo
echo "--- 新建项目 up ---"
curl -s -X POST -H "Cookie: $TOKEN" -H "Content-Type: application/json" http://localhost:$PORT/api/v1/projects -d '{"name":"debtest","yaml":"services:\n  hello:\n    image: busybox:1.36\n    container_name: dm-debtest\n    command: [\"sleep\", \"3600\"]\n","start":true}' | head -c 200
echo
echo "=== [4] 卸载清理 ==="
systemctl stop docker-manager
dpkg -r docker-manager
userdel docker-manager 2>/dev/null || true
rm -rf /var/lib/docker-manager /etc/docker-manager
docker rm -f dm-debtest 2>/dev/null || true
echo "--- 清理完成 ---"
EOF
