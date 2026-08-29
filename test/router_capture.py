"""捕获 dockermanager 手动运行的输出"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

print("==> 手动运行(默认配置)")
print(run("""rm -f /tmp/dm.log /tmp/dmtest.db
( /usr/bin/dockermanager > /tmp/dm.log 2>&1 & echo \$! > /tmp/dm.pid )
sleep 4
echo "--- 日志 ---"
cat /tmp/dm.log
echo "--- 进程 ---"
ps | grep dockermanager | grep -v grep
kill \$(cat /tmp/dm.pid) 2>/dev/null
"""))
print("==> procd 启动崩溃详情(用 strace?没有。查 dmesg)")
print(run("dmesg | grep -i 'dockermanager\\|segfault\\|killed' | tail -5"))

client.close()
