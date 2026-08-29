"""查看无后缀 ucode 文件格式,测试两种文件名"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

print("==> luci(无后缀)文件内容")
print(run("head -12 /usr/share/rpcd/ucode/luci"))
print("==> 试复制无后缀名并重启 rpcd")
print(run("cp /usr/share/rpcd/ucode/dockermanager.uc /usr/share/rpcd/ucode/dockermanager && /etc/init.d/rpcd restart && sleep 2 && ubus list | grep -iE 'docker|manager'"))

client.close()
