"""分离 ipk 三段,检查 control.tar.gz / data.tar.gz 内容"""
import gzip
import io
import subprocess
import paramiko

HOST, USER, PASS = "192.168.1.1", "root", "JDunix786"
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd, timeout=30):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    return stdout.read().decode().strip() + stderr.read().decode().strip()

sftp = client.open_sftp()
sftp.put(r"C:\Users\ms\Desktop\Hermes\GitHub\docker-manager\build\dockermanager_1.3.0_x86_64.ipk", "/tmp/dm.ipk")
sftp.close()

script = r'''
import gzip, io, tarfile
data = open('/tmp/dm.ipk','rb').read()
# 第一段 debian-binary:纯文本 "2.0\n"
assert data[:4] == b'2.0\n', f"debian-binary 不对: {data[:4]!r}"
pos = 4
# 逐段解析 gzip 流
for name in ['control.tar.gz', 'data.tar.gz']:
    # gzip 流结束位置
    buf = io.BytesIO(data[pos:])
    try:
        with gzip.GzipFile(fileobj=buf) as f:
            content = f.read()
    except Exception as e:
        print(f"{name}: 解析失败 {e}")
        break
    consumed = buf.tell()
    print(f"{name}: {len(content)} 字节(压缩流 {consumed} 字节)")
    tf = tarfile.open(fileobj=io.BytesIO(content))
    for m in tf.getmembers():
        print(f"   {m.name}  mode={oct(m.mode)}  size={m.size}")
    pos += consumed
print("三段解析完成,总偏移:", pos, "/", len(data))
'''
out = run(f"python3 - << 'PY'\n{script}\nPY")
print(out)
client.close()
