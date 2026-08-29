"""验证项目(compose)功能:发现/详情/新建/up/down/编辑"""
import json
import urllib.request
import time

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
    return json.loads(urllib.request.urlopen(req, timeout=30).read())

cookie = login()
print("登录 OK")

# === 1. 项目列表:应发现 dmproj(已有项目) ===
projects = api(cookie, "/api/v1/projects")["data"]
names = [p["name"] for p in projects]
print(f"项目列表: {names}")
assert "dmproj" in names, "未发现已有项目 dmproj"
dmproj = next(p for p in projects if p["name"] == "dmproj")
assert dmproj["source"] == "discovered", f"dmproj 应为 discovered: {dmproj['source']}"
assert "compose-test.yaml" in dmproj["config_files"][0], f"compose 文件名不对: {dmproj['config_files']}"
print(f"  发现 dmproj: 容器 {dmproj['containers']}/{dmproj['running']}, compose 文件: {dmproj['config_files']}")
print("已有项目自动发现 OK ✓")

# === 2. 项目详情:服务 + 卷(volume + bind) ===
detail = api(cookie, "/api/v1/projects/dmproj")["data"]
print(f"  服务: {[s['name'] for s in detail['services']]}")
assert any(s["name"] == "web" for s in detail["services"]), "服务 web 缺失"
vols = detail["volumes"]
vol_types = {v["type"] for v in vols}
print(f"  卷: {[(v['type'], v['name'], v['destination']) for v in vols]}")
assert "volume" in vol_types and "bind" in vol_types, f"应同时有 volume 和 bind: {vol_types}"
bind = next(v for v in vols if v["type"] == "bind")
assert bind["name"] == "/tmp/dm-bind-test", f"bind 源路径不对: {bind['name']}"
assert bind["rw"] is False, "bind 应为 ro"
vol = next(v for v in vols if v["type"] == "volume")
assert vol["mountpoint"], "volume 应有宿主机位置"
print("  详情:服务/卷/类型/读写 全部正确 ✓")
print(f"  compose 文件名: {detail['config_files']}")

# === 3. 新建项目(managed)+ up ===
new_yaml = """services:
  hello:
    image: busybox:1.36
    container_name: dm-myapp
    command: ["sleep", "3600"]
    volumes:
      - mydata:/data
volumes:
  mydata:
"""
res = api(cookie, "/api/v1/projects", method="POST", body={
    "name": "myapp", "yaml": new_yaml, "description": "测试项目", "start": True})
assert res["success"] and res["data"]["started"], f"新建失败: {res}"
print(f"新建+启动 OK: {res['data']['compose_file']}")

# 列表应显示 myapp(managed)
projects = api(cookie, "/api/v1/projects")["data"]
myapp = next(p for p in projects if p["name"] == "myapp")
assert myapp["source"] == "managed", f"myapp 应为 managed: {myapp['source']}"
print("myapp 在列表且标记 managed ✓")

# === 4. YAML 读取 ===
y = api(cookie, "/api/v1/projects/myapp/yaml")["data"]
assert "busybox" in y["yaml"], "YAML 内容不对"
print(f"YAML 读取 OK: {y['compose_file']}")

# === 5. down ===
res = api(cookie, "/api/v1/projects/myapp/down", method="POST", body={})
assert res["success"], f"down 失败: {res}"
time.sleep(1)
projects = api(cookie, "/api/v1/projects")["data"]
myapp = next(p for p in projects if p["name"] == "myapp")
assert not myapp["has_containers"], "down 后应有容器"
print("down OK(项目仍在列表,标记已停止)✓")

# === 6. 再次 up + 编辑 ===
res = api(cookie, "/api/v1/projects/myapp/up", method="POST")
assert res["success"], f"再次 up 失败: {res}"
time.sleep(1)
new_yaml2 = new_yaml.replace('command: ["sleep", "3600"]', 'command: ["sleep", "7200"]')
res = api(cookie, "/api/v1/projects/myapp", method="PUT", body={"yaml": new_yaml2})
assert res["success"], f"编辑失败: {res}"
y2 = api(cookie, "/api/v1/projects/myapp/yaml")["data"]
assert "7200" in y2["yaml"], "编辑未生效"
print("编辑 YAML OK ✓")

# === 7. 删除 managed 项目(先 down 再删文件,容器停了才彻底移除) ===
res = api(cookie, "/api/v1/projects/myapp/down", method="POST", body={})
assert res["success"], f"down 失败: {res}"
time.sleep(1)
res = api(cookie, "/api/v1/projects/myapp", method="DELETE")
assert res["success"], f"删除失败: {res}"
projects = api(cookie, "/api/v1/projects")["data"]
assert "myapp" not in [p["name"] for p in projects], "myapp 应已删除"
print("删除 managed 项目 OK ✓(先 down 再删)")

# === 8. 删除保护:discovered 项目不可删 ===
res = api(cookie, "/api/v1/projects/dmproj", method="DELETE")
assert not res["success"], "已有项目不应允许通过工具删除"
print("已有项目删除保护 OK ✓")

print("\n=== 项目功能 API 验证全部通过 ===")
