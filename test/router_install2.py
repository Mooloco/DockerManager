"""路由安装新 ipk(opkg),验证格式与功能"""
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

# 1. 上传
sftp = client.open_sftp()
sftp.put(r"C:\Users\ms\Desktop\Hermes\GitHub\docker-manager\build\dockermanager_1.3.0_x86_64.ipk", "/tmp/dm.ipk")
sftp.close()
print("上传完成")

# 2. 移除旧 iStoreOS 包(连带其 LuCI/依赖)
print("==> 移除旧包(含依赖)")
print(run("opkg remove dockermanager --force-removal-of-dependent-packages 2>&1 | tail -4"))

# 3. 安装
print("==> 安装")
print(run("opkg install /tmp/dm.ipk 2>&1 | tail -4"))

# 4. 配置
print("==> 配置 UCI")
for opt, val in [("enabled", "1"), ("port", "8080"), ("data_dir", "/ext/dockermanager"), ("admin_password", "test1234")]:
    print(run(f"uci set dockermanager.main.{opt}='{val}'"))
print(run("uci commit dockermanager"))
print(run("cat /etc/config/dockermanager"))

# 5. 启动
print("==> 启动")
print(run("/etc/init.d/dockermanager enable && /etc/init.d/dockermanager start"))
time.sleep(3)
print(run("/etc/init.d/dockermanager status"))

# 6. 验证
print("==> 验证")
TOKEN = run("curl -s -D - -o /dev/null -X POST http://localhost:8080/api/v1/auth/login "
            "-H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"test1234\"}' "
            "| grep -i set-cookie | sed 's/.*dm_session=\\([^;]*\\).*/dm_session=\\1/' | tr -d '\\r'")
print(f"登录: {TOKEN[:28]}...")
if not TOKEN.startswith("dm_session"):
    print("!! 登录失败,日志:")
    print(run("logread | grep -i dockermanager | tail -8"))
else:
    print("系统信息:", run(f"curl -s -H 'Cookie: {TOKEN}' http://localhost:8080/api/v1/system/info")[:180])
    projects = run(f"curl -s -H 'Cookie: {TOKEN}' http://localhost:8080/api/v1/projects")
    try:
        names = [p["name"] for p in json.loads(projects)["data"]]
        print(f"自动发现项目: {names}")
    except Exception:
        print("项目:", projects[:180])
    print("前端页面:", run("curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/"))

client.close()
