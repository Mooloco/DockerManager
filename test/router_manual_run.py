"""手动运行 dockermanager 二进制,捕获崩溃原因"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

print("==> 手动运行(前台 5 秒)")
print(run("timeout 5 /usr/bin/dockermanager 2>&1 | head -20; echo exit=$?"))
print("==> 带环境变量运行")
print(run("timeout 5 env DM_ADMIN_PASSWORD=test1234 SERVER_PORT=18099 DATABASE_PATH=/tmp/dmtest.db DM_PROJECTS_DIR=/tmp/dmproj /usr/bin/dockermanager 2>&1 | head -20; echo exit=$?"))
print("==> 文件检查")
print(run("ls -la /usr/bin/dockermanager /ext/dockermanager/ 2>&1; file /usr/bin/dockermanager 2>&1 | head -1"))
print("==> procd 日志")
print(run("logread | grep -A 3 'crash loop\\|dockermanager' | tail -15"))

client.close()
