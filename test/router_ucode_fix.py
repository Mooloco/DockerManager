"""部署修复后的 ucode,重启 rpcd,验证 RPC + 启停控制"""
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
sftp.put(r"C:\Users\ms\Desktop\Hermes\GitHub\docker-manager\packaging\openwrt\luci\root\usr\share\rpcd\ucode\dockermanager.uc", "/usr/share/rpcd/ucode/dockermanager.uc")
sftp.close()
print(run("rm -f /usr/share/rpcd/ucode/dockermanager && /etc/init.d/rpcd restart && sleep 2"))
print("==> ubus 对象")
print(run("ubus list | grep -i docker"))
print("==> status")
print(run("ubus call dockermanager status 2>&1"))
print("==> 启停控制测试")
print(run("ubus call dockermanager stop 2>&1; sleep 2; ubus call dockermanager status 2>&1; ubus call dockermanager start 2>&1; sleep 3; ubus call dockermanager status 2>&1"))

client.close()
