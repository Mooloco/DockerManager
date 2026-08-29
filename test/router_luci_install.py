"""安装 luci-app 并验证 RPC + 菜单"""
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
sftp.put(r"C:\Users\ms\Desktop\Hermes\GitHub\docker-manager\build\luci-app-dockermanager_1.3.0_all.ipk", "/tmp/luci.ipk")
sftp.close()
print("==> 安装")
print(run("opkg install /tmp/luci.ipk 2>&1 | tail -3"))
print("==> RPC 测试")
print(run("ubus call dockermanager status 2>&1"))
print(run("ubus call dockermanager status 2>&1 | head -1"))
print("==> 菜单文件")
print(run("ls /usr/share/luci/menu.d/ | grep docker"))
print("==> 视图文件")
print(run("ls /www/luci-static/resources/view/dockermanager/ 2>&1"))

client.close()
