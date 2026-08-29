"""修复路由 dockermanager 配置并启动验证"""
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

print("==> 包状态")
print(run("opkg list-installed | grep -E 'dockermanager|luci-app-dockermanager|app-meta-dockermanager'"))
print("==> 重置 UCI 配置")
print(run("rm -f /etc/config/dockermanager"))
print(run("""cat > /etc/config/dockermanager <<'EOF'
config main 'dockermanager'
	option enabled '1'
	option port '8080'
	option bind_addr '0.0.0.0'
	option data_dir '/ext/dockermanager'
	option projects_dir '/ext/dockermanager/projects'
	option admin_password 'test1234'
EOF
uci commit dockermanager
cat /etc/config/dockermanager"""))
print("==> 启动")
print(run("/etc/init.d/dockermanager enable && /etc/init.d/dockermanager start"))
time.sleep(4)
print(run("/etc/init.d/dockermanager status"))
print("==> 验证")
TOKEN = run("curl -s -D - -o /dev/null -X POST http://localhost:8080/api/v1/auth/login "
            "-H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"test1234\"}' "
            "| grep -i set-cookie | sed 's/.*dm_session=\\([^;]*\\).*/dm_session=\\1/' | tr -d '\\r'")
print(f"登录: {TOKEN[:28]}...")
if not TOKEN.startswith("dm_session"):
    print("!! 失败,日志:")
    print(run("logread | grep -i dockermanager | tail -10"))
else:
    print("系统信息:", run(f"curl -s -H 'Cookie: {TOKEN}' http://localhost:8080/api/v1/system/info")[:160])
    projects = run(f"curl -s -H 'Cookie: {TOKEN}' http://localhost:8080/api/v1/projects")
    try:
        names = [p["name"] for p in json.loads(projects)["data"]]
        print(f"自动发现项目: {names}")
    except Exception:
        print("项目:", projects[:160])
    print("前端页面:", run("curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/"))

client.close()
