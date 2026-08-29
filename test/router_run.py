"""路由 192.168.1.1:加载镜像 + 运行容器 + 功能验证(不动路由设置)"""
import json
import time
import urllib.request
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=15, look_for_keys=False, allow_agent=False)

def run(cmd, timeout=180):
    _, so, se = client.exec_command(cmd, timeout=timeout)
    return (so.read().decode(errors="replace") + se.read().decode(errors="replace")).strip()

print("==> 加载镜像")
print(run("docker load -i /tmp/dm-image.tar"))

print("==> 运行容器(端口 18080)")
print(run("docker rm -f dm-test 2>/dev/null; docker run -d --name dm-test -p 18080:8080 "
          "-v /var/run/docker.sock:/var/run/docker.sock -v /tmp/dm-data:/data "
          "-e DM_ADMIN_PASSWORD=test1234 -e DATABASE_PATH=/data/docker-manager.db "
          "docker-manager:dev"))
time.sleep(3)
print("状态:", run("docker ps --filter name=dm-test --format '{{.Status}} {{.Ports}}'"))
print("日志:", run("docker logs dm-test 2>&1 | tail -3"))

print("\n==> 功能验证")
BASE = "http://127.0.0.1:18080"
req = urllib.request.Request(f"{BASE}/api/v1/auth/login",
    data=json.dumps({"username": "admin", "password": "test1234"}).encode(),
    headers={"Content-Type": "application/json"})
cookie = urllib.request.urlopen(req, timeout=10).headers.get("Set-Cookie", "").split(";")[0]
print("登录 OK")

def api(path):
    r = urllib.request.urlopen(urllib.request.Request(f"{BASE}{path}", headers={"Cookie": cookie}), timeout=15)
    return json.loads(r.read())

info = api("/api/v1/system/info")["data"]
print(f"Dashboard: {info['operating_system']} | Docker {info['server_version']} | "
      f"容器 {info['containers']}(运行 {info['containers_running']}) | 镜像 {info['images']}")

cts = api("/api/v1/containers")["data"]
names = [c["name"] for c in cts]
print(f"容器列表 {len(names)} 个: {', '.join(names[:8])}")
assert any("mihomo" in n for n in names), "应能看到路由上的 mihomo 容器"
assert any("navi" in n for n in names), "应能看到路由上的 navi 容器"

imgs = api("/api/v1/images")["data"]
print(f"镜像列表 {len(imgs)} 个(含 docker-manager:dev: {any('docker-manager' in (t or '') for i in imgs for t in i['repo_tags'])})")

nets = api("/api/v1/networks")["data"]
print(f"网络列表 {len(nets)} 个")

print("\n=== 路由容器化运行测试通过 ===")
client.close()
