"""对比 luci.upnp 的 ucode 格式,查完整 ubus"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

print("==> ubus 完整列表(rpcd 相关)")
print(run("ubus list | grep -E 'luci|rpcd' | head -12"))
print("==> luci.upnp.uc 格式(前 25 行)")
print(run("head -25 /usr/share/rpcd/ucode/luci.upnp.uc"))
print("==> rpcd ucode 插件配置")
print(run("grep -r 'ucode' /etc/config/rpcd /etc/init.d/rpcd 2>/dev/null | head -5"))
print("==> 我们的文件完整内容")
print(run("head -20 /usr/share/rpcd/ucode/dockermanager.uc"))

client.close()
