"""手动调用 start_service 调试"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

print("==> start 命令详细输出")
print(run("cd /etc/init.d && ./dockermanager start; echo \"exit=$?\""))
print("==> 手动执行 start_service 逻辑")
print(run("""cd /etc/init.d
. /etc/rc.common
start_service() {
	echo "start_service called"
	config_load dockermanager
	config_get enabled main enabled 1
	config_get port main port 8080
	config_get data_dir main data_dir /ext/dockermanager
	config_get admin_password main admin_password ""
	echo "enabled=$enabled port=$port data_dir=$data_dir pw=${admin_password:+SET}"
}
start_service"""))
print("==> procd 实例状态")
print(run("ubus call service list 2>/dev/null | grep -A 6 dockermanager | head -12"))
print("==> 完整 logread")
print(run("logread | grep -iE 'dockermanager|procd.*instance' | tail -12"))

client.close()
