"""通过 paramiko 探测主路由 192.168.1.1 的 Docker 环境(只读操作)"""
import sys
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"

def run(client, cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    out = stdout.read().decode(errors="replace").strip()
    err = stderr.read().decode(errors="replace").strip()
    return out, err

def main():
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(HOST, username=USER, password=PASS, timeout=10, look_for_keys=False, allow_agent=False)
    print("SSH 连接成功")

    for label, cmd in [
        ("系统", "cat /etc/openwrt_release 2>/dev/null | head -3 || uname -a"),
        ("架构", "uname -m"),
        ("Docker 版本", "docker version --format '{{.Server.Version}}' 2>&1 | head -2"),
        ("容器列表", "docker ps --format '{{.Names}}' 2>&1 | head -10"),
        ("磁盘", "df -h / /tmp 2>/dev/null | awk 'NR==1||NR==2||NR==3'"),
        ("内存", "free -m | head -2"),
        ("docker.sock 权限", "ls -la /var/run/docker.sock 2>&1"),
    ]:
        out, err = run(client, cmd)
        print(f"\n=== {label} ===")
        print(out if out else err)

    client.close()

if __name__ == "__main__":
    main()
