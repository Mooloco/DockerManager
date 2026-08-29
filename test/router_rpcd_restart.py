"""重启 rpcd 并验证 RPC + LuCI 页面"""
import paramiko
import time

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=60):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

print("==> 重启 rpcd")
print(run("/etc/init.d/rpcd restart && sleep 2 && ubus list | grep -i docker"))
print("==> RPC status")
print(run("ubus call dockermanager status 2>&1"))
print("==> RPC start/stop 测试")
print(run("ubus call dockermanager stop 2>&1; sleep 2; ubus call dockermanager status 2>&1; ubus call dockermanager start 2>&1; sleep 3; ubus call dockermanager status 2>&1"))

client.close()
