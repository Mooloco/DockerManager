"""从本机验证路由 192.168.1.1:18080 上的 Docker Manager"""
import json
import urllib.request

BASE = "http://192.168.1.1:18080"

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
print(f"容器列表 {len(names)} 个: {', '.join(names[:10])}")
assert any("mihomo" in n for n in names), "应能看到 mihomo"
assert any("navi" in n for n in names), "应能看到 navi"
assert any("dm-test" in n for n in names), "应能看到 dm-test"

# 容器 stats(验证 CPU/内存数据)
dm = next(c for c in cts if c["name"] == "dm-test")
print(f"dm-test stats: cpu={dm['cpu_percent']}% mem={dm['memory_bytes']/1048576:.1f}MB")

imgs = api("/api/v1/images")["data"]
print(f"镜像 {len(imgs)} 个")
nets = api("/api/v1/networks")["data"]
print(f"网络 {len(nets)} 个")
vols = api("/api/v1/volumes")["data"]
print(f"卷 {len(vols)} 个")

# 前端页面
page = urllib.request.urlopen(f"{BASE}/", timeout=10).read().decode()
assert "<div id=\"app\">" in page, "前端页面未加载"
print("前端页面 OK")

print("\n=== 路由容器化运行测试全部通过 ===")
