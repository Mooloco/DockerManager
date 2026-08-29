"""调试 init 脚本"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

print("==> 语法检查")
print(run("sh -n /etc/init.d/dockermanager && echo SYNTAX_OK || echo SYNTAX_ERR"))
print("==> 手动执行 start 输出")
print(run("/etc/init.d/dockermanager stop 2>&1; /etc/init.d/dockermanager start 2>&1; echo ---; sleep 2; ps | grep dockermanager | grep -v grep | head -2"))
print("==> 手动执行 start_service 逻辑(模拟)")
print(run("""config_load dockermanager
config_get port main port 8080
config_get data_dir main data_dir /ext/dockermanager
config_get admin_password main admin_password ""
echo "port=$port data_dir=$data_dir pw=${admin_password:+SET}"""))
print("==> UCI 实际内容")
print(run("uci show dockermanager"))

client.close()
