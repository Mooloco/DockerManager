"""测试 ucode uci cursor 读取匿名 section + 诊断 status 错误"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

print("==> cursor 读匿名 section 测试")
print(run("""ucode -e '
const { cursor } = require("uci");
const c = cursor();
print("at-main:", c.get("dockermanager", "@main[0]", "port") || "EMPTY");
print("main:", c.get("dockermanager", "main", "port") || "EMPTY");
print("first:", c.get("dockermanager", "dockermanager", "port") || "EMPTY");
' 2>&1"""))
print("==> rpcd 前台日志(status 调用时)")
print(run("""(rpcd -d -f > /tmp/rpcd2.log 2>&1 &) ; sleep 2
ubus call dockermanager status 2>&1
sleep 1
kill $(pidof rpcd) 2>/dev/null
grep -iE 'error|dockermanager' /tmp/rpcd2.log | head -8"""))

client.close()
