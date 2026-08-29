"""确认 rpcd 是否真的重启 + 前台调试加载"""
import paramiko
import time

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=60):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

print("==> rpcd 进程")
print(run("ps | grep rpcd | grep -v grep"))
print("==> 重启前后 PID")
print(run("pidof rpcd; /etc/init.d/rpcd restart; sleep 2; pidof rpcd"))
print("==> 重启后 ubus")
print(run("ubus list | grep -iE 'docker|manager|luci' | head -8"))
print("==> rpcd 前台调试(2 秒)")
print(run("(rpcd -d -f > /tmp/rpcd.log 2>&1 &) ; sleep 3; grep -iE 'ucode|docker|error' /tmp/rpcd.log | head -10; kill $(pidof rpcd) 2>/dev/null; echo done"))

client.close()
