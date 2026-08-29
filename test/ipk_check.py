"""检查 ipk 包结构(opkg 报 malformed)"""
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

# 把 ipk 传到路由后检查
sftp = client.open_sftp()
sftp.put(r"C:\Users\ms\Desktop\Hermes\GitHub\docker-manager\build\dockermanager_1.3.0_x86_64.ipk", "/tmp/dm.ipk")
sftp.close()

print("==> 文件结构(前 4 字节 = gzip 头?)")
print(run("xxd /tmp/dm.ipk | head -2"))
print("==> 解包检查")
print(run("""cd /tmp && rm -rf dmx && mkdir dmx && cd dmx && \
dd if=../dm.ipk bs=1 skip=0 of=deb.bin count=$( (head -c 3 ../dm.ipk | wc -c) ) 2>/dev/null; \
# 简单方式:用 tar 测试各段
python3 - << 'PY'
import gzip, io
data = open('/tmp/dm.ipk','rb').read()
print('总大小:', len(data))
# debian-binary 是纯文本 "2.0\\n"(3 字节)
print('前 3 字节:', data[:3])
# 后面是 gzip 流(control.tar.gz),解出看内容
pos = 3
def read_gzip(d, pos):
    # 找到 gzip 魔数
    import zlib
    # 简单:尝试用 gzip 从 pos 开始读
    f = gzip.GzipFile(fileobj=io.BytesIO(d[pos:]))
    return f.read(), pos + len(f.fileobj.raw._fp.getvalue()) if hasattr(f.fileobj, 'raw') else None
# 用更简单的方法验证:分别尝试
import subprocess
r = subprocess.run(['tar','tzf','/tmp/dm.ipk'], capture_output=True, text=True)
print('tar tzf 直接读:', r.returncode, r.stdout[:200], r.stderr[:100])
PY
"""))
print("==> 尝试 opkg 装到临时目录")
print(run("opkg install /tmp/dm.ipk 2>&1 | head -5"))

client.close()
