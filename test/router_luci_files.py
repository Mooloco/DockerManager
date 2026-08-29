"""检查路由 LuCI 文件与菜单注册"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

print("==> 包内容")
print(run("opkg files luci-app-dockermanager 2>&1 | head -10"))
print("==> 视图文件")
print(run("ls -la /www/luci-static/resources/view/dockermanager/ 2>&1"))
print("==> 菜单 JSON")
print(run("cat /usr/share/luci/menu.d/luci-app-dockermanager.json 2>&1"))
print("==> LuCI 菜单注册检查")
print(run("ls /usr/share/luci/menu.d/ | wc -l"))
client.close()
