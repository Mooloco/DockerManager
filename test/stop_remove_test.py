"""验证:停止=compose stop(容器保留)/删除语义/启动按钮/网络端口"""
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
    return json.loads(urllib.request.urlopen(req, timeout=60).read())

def ssh(cmd):
    import subprocess
    return subprocess.run(["ssh", "root@192.168.1.24", cmd], capture_output=True, text=True).stdout.strip()

cookie = login()
print("登录 OK")

# 先清理可能残留的 myapp
try:
    api(cookie, "/api/v1/projects/myapp", method="DELETE")
except Exception:
    pass

# === 1. 新建项目(不启动)→ 应能 up(启动按钮修复) ===
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
res = api(cookie, "/api/v1/projects", method="POST", body={"name": "myapp", "yaml": yaml, "start": False})
assert res["success"] and not res["data"]["started"], "新建(不启动)失败"
projects = api(cookie, "/api/v1/projects")["data"]
myapp = next(p for p in projects if p["name"] == "myapp")
assert myapp["has_containers"] is False, "未启动应无容器"
print("新建不启动 OK(has_containers=false)")

# === 2. up ===
res = api(cookie, "/api/v1/projects/myapp/up", method="POST")
assert res["success"], f"up 失败: {res}"
time.sleep(2)
print("up OK")

# === 3. stop = compose stop(容器保留) ===
res = api(cookie, "/api/v1/projects/myapp/stop", method="POST")
assert res["success"], f"stop 失败: {res}"
time.sleep(1)
# 容器应存在但停止
state = ssh("docker inspect -f '{{.State.Status}}' dm-myapp")
assert state == "exited", f"stop 后容器应为 exited,实际: {state}"
count = ssh("docker ps -a --filter name=dm-myapp --format '{{.Names}}' | wc -l")
assert count.strip() == "1", "容器应保留(未被删除)"
print(f"stop=compose stop OK(容器保留,状态 {state})✓")

# 详情应显示 running=0
detail = api(cookie, "/api/v1/projects/myapp")["data"]
assert detail["running"] == 0 and detail["has_containers"], "停止后 running=0 且容器存在"
print("详情状态 OK(running=0,容器存在)")

# === 4. 再次 up(复用容器) ===
res = api(cookie, "/api/v1/projects/myapp/up", method="POST")
assert res["success"], "再次 up 失败"
time.sleep(2)
state = ssh("docker inspect -f '{{.State.Status}}' dm-myapp")
assert state == "running", f"再次 up 后应 running: {state}"
print("再次 up 复用容器 OK ✓")

# === 5. 详情:网络 + 端口 ===
detail = api(cookie, "/api/v1/projects/myapp")["data"]
nets = detail["networks"]
print(f"  网络: {[(n['name'], n['driver'], [(c['name'], c['ip']) for c in n['containers']]) for n in nets]}")
assert any("myapp" in n["name"] for n in nets), "应有项目网络"
ports_found = any(p["public_port"] for s in detail["services"] for c in s["containers"] for p in c.get("ports", []))
assert ports_found, "端口映射缺失"
print("详情网络/端口 OK ✓")

# === 6. 删除 managed = down + 删文件 ===
res = api(cookie, "/api/v1/projects/myapp", method="DELETE")
assert res["success"] and res["data"]["file_gone"] is True, f"删除失败: {res}"
time.sleep(1)
# 容器应被删除
count = ssh("docker ps -a --filter name=dm-myapp --format '{{.Names}}' | wc -l")
assert count.strip() == "0", "删除后容器应消失"
# 项目文件应被删
exists = ssh("ls /mnt/scsi1/docker-manager/data/projects/myapp/ 2>&1")
assert "No such file" in exists, "项目文件应已删除"
# 列表应无 myapp
projects = api(cookie, "/api/v1/projects")["data"]
assert "myapp" not in [p["name"] for p in projects], "myapp 应从列表消失"
print("删除 managed=down+删文件 OK ✓")

print("\n=== 停止/删除语义验证全部通过 ===")
