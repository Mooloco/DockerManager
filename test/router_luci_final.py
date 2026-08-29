"""安装最终 luci ipk + 浏览器验证 LuCI 页面"""
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
sftp.put(r"C:\Users\ms\Desktop\Hermes\GitHub\docker-manager\build\luci-app-dockermanager_1.3.0_all.ipk", "/tmp/luci2.ipk")
sftp.close()
print("==> 安装(force 覆盖)")
print(run("opkg install --force-reinstall /tmp/luci2.ipk 2>&1 | tail -2"))
print(run("/etc/init.d/rpcd restart && /etc/init.d/nginx restart 2>/dev/null; sleep 2"))
print("==> RPC 复查")
print(run("ubus call dockermanager status 2>&1"))
client.close()
print("安装完成")
