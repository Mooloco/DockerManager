"""检查路由上已存在的 dockermanager 包详情"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

cmds = [
    "opkg info dockermanager 2>/dev/null",
    "ls -la /usr/bin/dockermanager /etc/init.d/dockermanager 2>&1",
    "cat /etc/init.d/dockermanager 2>/dev/null | head -30",
    "ls /etc/config/dockermanager 2>&1; cat /etc/config/dockermanager 2>/dev/null | head -10",
    "ls /usr/lib/lua/luci/controller/ 2>/dev/null | grep -i docker; ls /www/luci-static/resources/view/ 2>/dev/null | grep -i docker",
]
for c in cmds:
    stdin, stdout, stderr = client.exec_command(c, timeout=15)
    out = stdout.read().decode().strip()
    print(f"$ {c}\n{out}\n")

client.close()
