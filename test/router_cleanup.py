"""清理路由上的测试资源(只删除 dm-test 相关,不动任何用户设置)"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=15, look_for_keys=False, allow_agent=False)

def run(cmd, timeout=60):
    _, so, se = client.exec_command(cmd, timeout=timeout)
    return (so.read().decode(errors="replace") + se.read().decode(errors="replace")).strip()

print("停止并删除测试容器:", run("docker rm -f dm-test"))
print("删除测试镜像:", run("docker rmi docker-manager:dev"))
print("清理临时文件:", run("rm -f /tmp/dm-image.tar && rm -rf /tmp/dm-data && echo done"))
print("确认:", run("docker ps --filter name=dm-test --format '{{.Names}}' | wc -l"))
client.close()
print("=== 路由清理完成,用户设置未动 ===")
