"""对比 luci.upnp return 结构与 rpcd ucode 加载机制"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

print("==> luci.upnp 完整内容")
print(run("cat /usr/share/rpcd/ucode/luci.upnp | tail -40"))
print("==> ucode.so 是否被 rpcd 加载")
print(run("ls -la /usr/lib/rpcd/ucode.so; ldd /usr/lib/rpcd/ucode.so 2>/dev/null | head -3 || echo '无依赖'"))

client.close()
