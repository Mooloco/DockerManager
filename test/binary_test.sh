#!/bin/bash
# 单文件二进制独立运行验证(测试机)
set -e

BIN=/tmp/docker-manager-linux-amd64
scp /c/Users/ms/Desktop/Hermes/GitHub/docker-manager/build/docker-manager-linux-amd64 root@192.168.1.24:$BIN

ssh root@192.168.1.24 << 'EOF'
chmod +x /tmp/docker-manager-linux-amd64
cd /tmp
DM_ADMIN_PASSWORD=test1234 SERVER_PORT=18081 DATABASE_PATH=/tmp/dm-binary-test.db ./docker-manager-linux-amd64 > /tmp/dm-binary.log 2>&1 &
SRV=$!
sleep 2

echo "=== 登录页 ==="
curl -s -o /dev/null -w "HTTP %{http_code}\n" http://localhost:18081/

TOKEN=$(curl -s -D - -o /dev/null -X POST http://localhost:18081/api/v1/auth/login \
  -H 'Content-Type: application/json' -d '{"username":"admin","password":"test1234"}' \
  | grep -i set-cookie | sed 's/Set-Cookie: //' | cut -d';' -f1)

echo "=== system/info(单文件直连 docker.sock) ==="
curl -s -H "Cookie: $TOKEN" http://localhost:18081/api/v1/system/info | head -c 150
echo

echo "=== containers ==="
curl -s -H "Cookie: $TOKEN" http://localhost:18081/api/v1/containers | python3 -c "import json,sys; d=json.load(sys.stdin); print('容器数:', len(d['data']))"

kill $SRV 2>/dev/null || true
rm -f /tmp/docker-manager-linux-amd64 /tmp/dm-binary-test.db /tmp/dm-binary.log
echo "=== 单文件独立运行验证完成 ==="
EOF
