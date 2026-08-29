"""查 procd 完整 env 与 admin_password 读取"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

print("==> procd 完整 env")
print(run("ubus call service list 2>/dev/null | python3 -c \"import json,sys; d=json.load(sys.stdin); print(json.dumps(d.get('dockermanager',{}), indent=1))\" 2>/dev/null | head -30"))
print("==> 密码读取测试(引用方式)")
print(run("""cd /etc/init.d
. /etc/rc.common
config_load dockermanager
config_get admin_password main admin_password ""
printf 'pw_len=%s\\n' "${#admin_password}"
printf 'pw_val=%s\\n' "$admin_password" 2>&1 | head -1"""))

client.close()
