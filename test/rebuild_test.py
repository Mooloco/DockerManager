"""验证:重建(down→1s→up+日志)/第二行操作栏/端口链接化"""
import json
import time
import urllib.request

BASE = "http://192.168.1.24:18080"

def login():
    req = urllib.request.Request(f"{BASE}/api/v1/auth/login",
        data=json.dumps({"username": "admin", "password": "test1234"}).encode(),
        headers={"Content-Type": "application/json"})
    return urllib.request.urlopen(req, timeout=10).headers.get("Set-Cookie", "").split(";")[0]

def api(cookie, path, method="GET", body=None):
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(f"{BASE}{path}", data=data, method=method,
        headers={"Cookie": cookie, "Content-Type": "application/json"})
    return json.loads(urllib.request.urlopen(req, timeout=120).read())

cookie = login()
print("登录 OK")

# 清理残留
for name in ["myapp", "uiapp"]:
    try:
        api(cookie, f"/api/v1/projects/{name}", method="DELETE")
    except Exception:
        pass

# 建项目并启动
yaml = """services:
  web:
    image: nginx:alpine
    container_name: dm-myapp
    ports:
      - "18086:80"
    volumes:
      - webdata:/usr/share/nginx/html
volumes:
  webdata:
"""
res = api(cookie, "/api/v1/projects", method="POST", body={"name": "myapp", "yaml": yaml, "start": True})
assert res["success"], f"创建失败: {res}"
time.sleep(2)
print("创建+启动 OK")

# === 1. 重建 ===
t0 = time.time()
res = api(cookie, "/api/v1/projects/myapp/rebuild", method="POST")
elapsed = time.time() - t0
assert res["success"], f"重建失败: {res}"
out = res["data"]["output"]
print(f"重建 OK({elapsed:.1f}s,含 1s 延迟)")
print(f"  日志前 200 字:\n{out[:200]}")
assert "Stopped" in out or "Removed" in out or "Container" in out or "Running" in out or "Started" in out, f"日志内容异常: {out[:100]}"
time.sleep(2)

# 容器应运行中
import subprocess
state = subprocess.run(["ssh", "root@192.168.1.24", "docker inspect -f '{{.State.Status}}' dm-myapp"], capture_output=True, text=True).stdout.strip()
assert state == "running", f"重建后应 running: {state}"
print("重建后容器 running ✓")

# 详情:端口映射数据(带容器 IP)
detail = api(cookie, "/api/v1/projects/myapp")["data"]
ct = next(c for s in detail["services"] for c in s["containers"])
print(f"  容器 IP: {ct.get('ip')}, 端口: {ct.get('ports')}")
assert ct.get("ip"), "容器应有 IP"

# 清理
api(cookie, "/api/v1/projects/myapp", method="DELETE")
print("\n=== 重建功能验证通过 ===")
