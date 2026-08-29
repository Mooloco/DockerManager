"""更新路由 config 为匿名 section 并重启验证"""
import json
import paramiko
import time

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=60):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

# 1. 写标准 UCI 配置(匿名 section)
print(run("""cat > /etc/config/dockermanager <<'EOF'
config main
	option enabled '1'
	option port '8080'
	option bind_addr '0.0.0.0'
	option data_dir '/ext/dockermanager'
	option projects_dir '/ext/dockermanager/projects'
	option admin_password 'test1234'
EOF
uci commit dockermanager
uci show dockermanager"""))

# 2. 重启
print("==> 重启服务")
print(run("/etc/init.d/dockermanager stop 2>&1; /etc/init.d/dockermanager start 2>&1"))
time.sleep(4)
print("status:", run("/etc/init.d/dockermanager status"))
print("进程:", run("ps | grep -E 'dockermanager' | grep -v grep | head -2"))

# 3. 验证
TOKEN = run("curl -s -D - -o /dev/null -X POST http://localhost:8080/api/v1/auth/login "
            "-H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"test1234\"}' "
            "| grep -i set-cookie | sed 's/.*dm_session=\\([^;]*\\).*/dm_session=\\1/' | tr -d '\\r'")
print(f"登录: {TOKEN[:28]}...")
if not TOKEN.startswith("dm_session"):
    print("!! 失败:", run("logread | grep -i 'dockermanager' | tail -8"))
else:
    print("系统信息:", run(f"curl -s -H 'Cookie: {TOKEN}' http://localhost:8080/api/v1/system/info")[:150])
    projects = run(f"curl -s -H 'Cookie: {TOKEN}' http://localhost:8080/api/v1/projects")
    try:
        names = [p["name"] for p in json.loads(projects)["data"]]
        print(f"自动发现项目: {names}")
    except Exception:
        print("项目:", projects[:150])
    print("前端页面:", run("curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/"))

client.close()
