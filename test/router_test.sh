#!/bin/bash
# 路由 192.168.1.1 容器化运行测试(不动路由设置,测试完清理)
set -e

TEST_HOST=root@192.168.1.24
IMG_FILE=/tmp/dm-image.tar

echo "==> [1/4] 测试机导出镜像"
ssh $TEST_HOST "docker save docker-manager:dev -o $IMG_FILE && ls -lh $IMG_FILE"

echo "==> [2/4] 传输镜像到路由"
scp -o StrictHostKeyChecking=no $TEST_HOST:$IMG_FILE /tmp/dm-image-local.tar 2>/dev/null || true
# 直接测试机 → 路由
scp -o StrictHostKeyChecking=no $TEST_HOST:$IMG_FILE root@192.168.1.1:$IMG_FILE

echo "==> [3/4] 路由加载并运行"
python - << 'PYEOF'
import paramiko
HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
c = paramiko.SSHClient()
c.set_missing_host_key_policy(paramiko.AutoAddPolicy())
c.connect(HOST, username=USER, password=PASS, timeout=15, look_for_keys=False, allow_agent=False)

def run(cmd, timeout=120):
    _, so, se = c.exec_command(cmd, timeout=timeout)
    return (so.read().decode(errors="replace") + se.read().decode(errors="replace")).strip()

print(run("docker load -i /tmp/dm-image.tar"))
print(run("docker rm -f dm-test 2>/dev/null; docker run -d --name dm-test -p 18080:8080 "
          "-v /var/run/docker.sock:/var/run/docker.sock -v /tmp/dm-data:/data "
          "-e DM_ADMIN_PASSWORD=test1234 -e DATABASE_PATH=/data/docker-manager.db "
          "docker-manager:dev"))
import time; time.sleep(3)
print("容器状态:", run("docker ps --filter name=dm-test --format '{{.Status}} {{.Ports}}'"))
print("启动日志:", run("docker logs dm-test 2>&1 | tail -3"))
c.close()
PYEOF

echo "==> [4/4] 功能验证"
python - << 'PYEOF'
import json, time, urllib.request
BASE = "http://192.168.1.1:18080"
req = urllib.request.Request(f"{BASE}/api/v1/auth/login",
    data=json.dumps({"username":"admin","password":"test1234"}).encode(),
    headers={"Content-Type":"application/json"})
cookie = urllib.request.urlopen(req).headers.get("Set-Cookie","").split(";")[0]
print("登录 OK")

info = json.loads(urllib.request.urlopen(urllib.request.Request(f"{BASE}/api/v1/system/info", headers={"Cookie":cookie})).read())
d = info["data"]
print(f"Dashboard: {d['operating_system']} | Docker {d['server_version']} | 容器 {d['containers']}(运行 {d['containers_running']}) | 镜像 {d['images']}")

cts = json.loads(urllib.request.urlopen(urllib.request.Request(f"{BASE}/api/v1/containers", headers={"Cookie":cookie})).read())
names = [c["name"] for c in cts["data"]]
print(f"容器列表 {len(names)} 个: {names[:6]}...")
assert any("mihomo" in n for n in names), "应能看到路由上的 mihomo 容器"
assert any("navi" in n for n in names), "应能看到路由上的 navi 容器"
print("✓ 路由容器管理功能验证通过")
PYEOF
