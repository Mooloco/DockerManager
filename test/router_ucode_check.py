"""检查 ucode 语法与 rpcd 加载错误"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

print("==> 文件与权限")
print(run("ls -la /usr/share/rpcd/ucode/dockermanager.uc"))
print("==> ucode 语法检查")
print(run("which ucode; ucode -c /usr/share/rpcd/ucode/dockermanager.uc 2>&1 && echo SYNTAX_OK || echo SYNTAX_ERR"))
print("==> rpcd 日志")
print(run("logread | grep -i rpcd | tail -8"))
print("==> 其他 ucode 服务对比")
print(run("ls /usr/share/rpcd/ucode/ | head -8"))
print("==> 直接执行测试(popen 是否可用)")
print(run("ucode -e 'const fs = require(\"fs\"); print(fs.popen(\"uci -q get dockermanager.@main[0].port\").read(\"line\"));' 2>&1"))

client.close()
