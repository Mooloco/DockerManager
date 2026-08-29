"""检查路由 Docker 数据目录与端口占用"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10, look_for_keys=False, allow_agent=False)

def run(cmd):
    _, stdout, stderr = client.exec_command(cmd, timeout=30)
    return (stdout.read().decode(errors="replace") + stderr.read().decode(errors="replace")).strip()

print("=== Docker Root Dir ===")
print(run("docker info --format '{{.DockerRootDir}}'"))
print("\n=== Root Dir 所在分区 ===")
print(run("docker info --format '{{.DockerRootDir}}' | xargs df -h | tail -1"))
print("\n=== 镜像数/占用 ===")
print(run("docker system df"))
print("\n=== 常用端口占用 ===")
print(run("netstat -tlnp 2>/dev/null | grep -E ':(8080|18080|18081|18082|18083)' || echo '目标端口均空闲'"))
client.close()
