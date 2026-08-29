"""路由安装 dockermanager ipk(替换 iStoreOS 自带),验证后保留服务"""
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

# 1. 上传 ipk
print("==> 上传 ipk")
sftp = client.open_sftp()
sftp.put(r"C:\Users\ms\Desktop\Hermes\GitHub\docker-manager\build\dockermanager_1.3.0_x86_64.ipk", "/tmp/dockermanager.ipk")
sftp.close()
print("上传完成")

# 2. 卸载 iStoreOS 自带
print("==> 卸载旧 dockermanager")
print(run("opkg remove dockermanager"))
print("==> 安装我们的")
print(run("opkg install /tmp/dockermanager.ipk"))

# 3. 配置并启动
print("==> 配置 UCI")
print(run("uci set dockermanager.main.enabled='1'"))
print(run("uci set dockermanager.main.port='8080'"))
print(run("uci set dockermanager.main.data_dir='/ext/dockermanager'"))
print(run("uci set dockermanager.main.admin_password='test1234'"))
print(run("uci commit dockermanager"))
print("==> 启动服务")
print(run("/etc/init.d/dockermanager enable && /etc/init.d/dockermanager start"))
time.sleep(3)
print(run("/etc/init.d/dockermanager status"))

# 4. 验证
print("==> 验证:登录 + 系统信息 + 项目")
TOKEN = run("curl -s -D - -o /dev/null -X POST http://localhost:8080/api/v1/auth/login "
            "-H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"test1234\"}' "
            "| grep -i set-cookie | sed 's/.*dm_session=\\([^;]*\\).*/dm_session=\\1/' | tr -d '\\r'")
print(f"登录: {TOKEN[:30]}...")
if not TOKEN.startswith("dm_session"):
    print("!! 登录失败")
    print(run("logread | grep -i dockermanager | tail -10"))
else:
    sysinfo = run(f"curl -s -H 'Cookie: {TOKEN}' http://localhost:8080/api/v1/system/info")
    print("系统信息:", sysinfo[:200])
    projects = run(f"curl -s -H 'Cookie: {TOKEN}' http://localhost:8080/api/v1/projects")
    try:
        data = json.loads(projects)["data"]
        names = [p["name"] for p in data]
        print(f"自动发现项目: {names}")
    except Exception:
        print("项目列表:", projects[:200])
    page = run("curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/")
    print(f"前端页面: HTTP {page}")

client.close()
