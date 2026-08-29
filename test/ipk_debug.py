"""在测试机手动组装 ipk,逐步检查每步结果"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

# 先建目录再上传
print(run("rm -rf /tmp/ipkdbg && mkdir -p /tmp/ipkdbg/usr/bin /tmp/ipkdbg/etc/init.d /tmp/ipkdbg/etc/config"))
sftp = client.open_sftp()
sftp.put(r"C:\Users\ms\Desktop\Hermes\GitHub\docker-manager\build\docker-manager-linux-amd64", "/tmp/ipkdbg/usr/bin/dockermanager")
sftp.close()

print(run("""
set -e
rm -rf /tmp/ipkdbg && mkdir -p /tmp/ipkdbg/usr/bin /tmp/ipkdbg/etc/init.d /tmp/ipkdbg/etc/config
# 上传的二进制在 /tmp/ipkdbg/usr/bin/dockermanager(上面 sftp)
echo "=== 文件就位 ==="
ls -la /tmp/ipkdbg/usr/bin/
# 写 init 脚本和 config
cat > /tmp/ipkdbg/etc/init.d/dockermanager <<'EOF'
#!/bin/sh /etc/rc.common
USE_PROCD=1
START=96
EOF
echo 'config main' > /tmp/ipkdbg/etc/config/dockermanager

chmod 755 /tmp/ipkdbg/usr/bin/dockermanager /tmp/ipkdbg/etc/init.d/dockermanager

echo "=== 组装 ==="
rm -rf /tmp/ipk2 && mkdir -p /tmp/ipk2/control /tmp/ipk2/data
cp -a /tmp/ipkdbg/usr /tmp/ipkdbg/etc /tmp/ipk2/data/
ls -laR /tmp/ipk2/data | head -12
cat > /tmp/ipk2/control/control <<'CTRL'
Package: dockermanager
Version: 1.3.0
Architecture: x86_64
CTRL
cd /tmp/ipk2
echo "2.0" > debian-binary
tar czf control.tar.gz -C control .
echo "control.tar.gz 大小: $(stat -c %s control.tar.gz)"
tar czf data.tar.gz -C data .
echo "data.tar.gz 大小: $(stat -c %s data.tar.gz)"
echo "=== 解压验证 control ==="
tar tzf control.tar.gz
echo "=== 解压验证 data ==="
tar tzf data.tar.gz | head -8
cat debian-binary control.tar.gz data.tar.gz > /tmp/test.ipk
echo "=== ipk 大小: $(stat -c %s /tmp/test.ipk)"
"""))

client.close()
