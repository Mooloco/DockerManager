#!/bin/bash
# deb 安装后的启动 + 功能验证 + 卸载(在测试机执行)
set -e
ssh root@192.168.1.24 << 'EOF'
set -e
echo "=== [3] 配置密码并启动 ==="
echo "DM_ADMIN_PASSWORD=test1234" > /etc/docker-manager/env
chmod 600 /etc/docker-manager/env
systemctl start docker-manager
sleep 2
systemctl is-active docker-manager
echo "--- 服务日志 ---"
journalctl -u docker-manager --no-pager -n 5 | tail -5

echo "=== [4] 功能验证 ==="
TOKEN=$(curl -s -D - -o /dev/null -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' -d '{"username":"admin","password":"test1234"}' \
  | grep -i set-cookie | sed 's/Set-Cookie: //' | cut -d';' -f1)
echo "登录: ${TOKEN:0:25}..."
curl -s -H "Cookie: $TOKEN" http://localhost:8080/api/v1/system/info | python3 -c "import json,sys; d=json.load(sys.stdin)['data']; print('Docker', d['server_version'], '| 容器', d['containers'], '| 镜像', d['images'])"
curl -s -o /dev/null -w "前端页面: %{http_code}\n" http://localhost:8080/

echo "=== [5] 卸载 ==="
systemctl stop docker-manager
dpkg -r docker-manager
echo "服务状态: $(systemctl is-active docker-manager 2>&1)"
echo "二进制: $(ls /usr/bin/docker-manager 2>&1)"
echo "用户: $(id docker-manager 2>&1)"

echo "=== [6] 清理测试残留 ==="
rm -f /tmp/docker-manager_0.1.0_amd64.deb
rm -rf /var/lib/docker-manager /etc/docker-manager
echo "=== deb 全流程验证完成 ==="
EOF
