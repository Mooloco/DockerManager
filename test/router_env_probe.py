"""探测路由 OpenWrt 环境:包管理/docker 包名/端口/服务目录"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

cmds = [
    "cat /etc/openwrt_release 2>/dev/null | head -3",
    "which opkg apk 2>/dev/null",
    "opkg list-installed 2>/dev/null | grep -E '^(dockerd|docker|luci|docker-compose)' | head -8",
    "ss -tlnp | grep -E ':8080|:1808' | head -3",
    "ls /etc/init.d/ | grep -iE 'docker|dock' ",
    "grep -r 'START=' /etc/init.d/dockerd 2>/dev/null",
    "df -h /ext 2>/dev/null | tail -1; df -h /overlay 2>/dev/null | tail -1",
]
for c in cmds:
    stdin, stdout, stderr = client.exec_command(c, timeout=15)
    out = stdout.read().decode().strip()
    print(f"$ {c}\n{out}\n")

client.close()
