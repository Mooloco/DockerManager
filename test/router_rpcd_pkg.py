"""查 rpcd ucode 支持与目录结构"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

print("==> rpcd 相关包")
print(run("opkg list-installed | grep -iE 'rpcd|ucode' | head -8"))
print("==> ucode 目录内容(带类型)")
print(run("ls -la /usr/share/rpcd/ucode/ | head -12"))
print("==> 其他 rpcd 插件目录")
print(run("ls /usr/lib/rpcd/ 2>/dev/null | head -12"))
print("==> rpcd 配置")
print(run("cat /etc/config/rpcd 2>/dev/null | head -15"))

client.close()
