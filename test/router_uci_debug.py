"""调试 config_load/config_get 为何读不到 UCI"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

print("==> 直连 uci 命令")
print(run("uci -q get dockermanager.@main[0].admin_password; echo rc=$?"))
print("==> rc.common 环境测试")
print(run("""cd /etc/init.d
. /etc/rc.common
config_load dockermanager
echo "CONFIG=$CONFIG"
echo "UCI_GET=$(uci -q get $CONFIG.@main[0].admin_password)"
config_get test_pw main admin_password ""
echo "config_get test_pw=[$test_pw]"
config_get test_pw2 @main[0] admin_password ""
echo "config_get @main test_pw2=[$test_pw2]" """))

client.close()
