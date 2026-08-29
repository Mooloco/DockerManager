#!/bin/bash
# 测试机 deb 安装/运行/卸载全流程验证
set -e

DEB=/tmp/docker-manager_0.1.0_amd64.deb
scp /c/Users/ms/Desktop/Hermes/GitHub/docker-manager/build/docker-manager_0.1.0_amd64.deb root@192.168.1.24:$DEB

ssh root@192.168.1.24 << 'EOF'
set -e
echo "=== [1] 安装 deb ==="
dpkg -i /tmp/docker-manager_0.1.0_amd64.deb

echo "=== [2] 检查安装结果 ==="
which docker-manager
ls -la /usr/bin/docker-manager /usr/lib/systemd/system/docker-manager.service
id docker-manager
ls -la /var/lib/docker-manager/
cat /etc/docker-manager/config.yaml | head -6

echo "=== [3] 配置初始密码并启动服务 ==="
echo "DM_ADMIN_PASSWORD=test1234" > /etc/docker-manager/env
chmod 600 /etc/docker-manager/env
systemctl daemon-reload
systemctl start docker-manager
sleep 2
systemctl status docker-manager --no-pager -l | head -6

echo "=== [4] 功能验证 ==="
TOKEN=$(curl -s -D - -o /dev/null -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' -d '{"username":"admin","password":"test1234"}' \
  | grep -i set-cookie | sed 's/Set-Cookie: //' | cut -d';' -f1)
curl -s -H "Cookie: $TOKEN" http://localhost:8080/api/v1/system/info | python3 -c "import json,sys; d=json.load(sys.stdin)['data']; print('Docker', d['server_version'], '| 容器', d['containers'])"
curl -s -o /dev/null -w "前端页面: %{http_code}\n" http://localhost:8080/

echo "=== [5] 卸载测试 ==="
systemctl stop docker-manager
dpkg -r docker-manager
systemctl status docker-manager --no-pager 2>&1 | head -2 || true
echo "卸载后 docker-manager 用户: $(id docker-manager 2>&1)"
ls /usr/bin/docker-manager 2>&1 || echo "二进制已移除"

echo "=== [6] 清理 ==="
rm -f /tmp/docker-manager_0.1.0_amd64.deb /etc/docker-manager/env
rm -rf /var/lib/docker-manager
echo "=== deb 全流程验证完成 ==="
EOF
