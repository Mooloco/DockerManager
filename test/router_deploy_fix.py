"""部署修复后的 init 脚本并完整验证"""
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

sftp = client.open_sftp()
sftp.put(r"C:\Users\ms\Desktop\Hermes\GitHub\docker-manager\packaging\openwrt\root\etc\init.d\dockermanager", "/etc/init.d/dockermanager")
sftp.close()
print(run("chmod 755 /etc/init.d/dockermanager && /etc/init.d/dockermanager stop 2>&1; sleep 1; /etc/init.d/dockermanager start 2>&1; echo rc=$?"))
time.sleep(4)
print("status:", run("/etc/init.d/dockermanager status"))
print("进程:", run("ps | grep dockermanager | grep -v grep | head -2"))

TOKEN = run("curl -s -D - -o /dev/null -X POST http://localhost:8080/api/v1/auth/login "
            "-H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"test1234\"}' "
            "| grep -i set-cookie | sed 's/.*dm_session=\\([^;]*\\).*/dm_session=\\1/' | tr -d '\\r'")
print(f"登录: {TOKEN[:28]}...")
if not TOKEN.startswith("dm_session"):
    print("!! 失败:", run("logread | grep -i 'dockermanager' | tail -6"))
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
