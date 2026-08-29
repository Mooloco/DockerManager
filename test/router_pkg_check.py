"""确认 iStoreOS dockermanager 0.1.1 的来源与构成"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

cmds = [
    "opkg info dockermanager",
    "ls -la /usr/bin/ | grep -i docker",
    "ls /apps/dockermanager/ 2>/dev/null",
    "ss -tlnp | grep 8192",
    "opkg files dockermanager 2>/dev/null | head -20",
]
for c in cmds:
    stdin, stdout, stderr = client.exec_command(c, timeout=15)
    out = stdout.read().decode().strip()
    print(f"$ {c}\n{out}\n")

client.close()
