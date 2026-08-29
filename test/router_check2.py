"""路由 192.168.1.1:检查 dm-test 容器状态"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=15, look_for_keys=False, allow_agent=False)

def run(cmd, timeout=60):
    _, so, se = client.exec_command(cmd, timeout=timeout)
    return (so.read().decode(errors="replace") + se.read().decode(errors="replace")).strip()

print("容器状态:", run("docker ps -a --filter name=dm-test --format '{{.Status}} {{.Ports}}'"))
print("镜像:", run("docker images docker-manager:dev --format '{{.Repository}}:{{.Tag}} {{.Size}}'"))
print("启动日志:", run("docker logs dm-test 2>&1 | tail -5"))
print("端口监听:", run("netstat -tlnp 2>/dev/null | grep 18080 || echo '未监听'"))
client.close()
