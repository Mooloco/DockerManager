#!/bin/bash
# 纯 ssh 调试 ipk 组装
set -e
ssh root@192.168.1.24 << 'EOF'
set -x
rm -rf /tmp/ipkd && mkdir -p /tmp/ipkd/src/usr/bin /tmp/ipkd/src/etc/init.d /tmp/ipkd/src/etc/config
cp /tmp/ipkdbg/usr/bin/dockermanager /tmp/ipkd/src/usr/bin/ 2>/dev/null || echo "no ipkdbg file"
ls -la /tmp/ipkd/src/usr/bin/
# 上传真实二进制
EOF
scp /c/Users/ms/Desktop/Hermes/GitHub/docker-manager/build/docker-manager-linux-amd64 root@192.168.1.24:/tmp/ipkd/src/usr/bin/dockermanager
ssh root@192.168.1.24 << 'EOF'
set -ex
chmod 755 /tmp/ipkd/src/usr/bin/dockermanager
ls -la /tmp/ipkd/src/usr/bin/
rm -rf /tmp/ipkb && mkdir -p /tmp/ipkb/control /tmp/ipkb/data
cp -a /tmp/ipkd/src/usr /tmp/ipkd/src/etc /tmp/ipkb/data/
echo "=== data 目录 ==="
find /tmp/ipkb/data -type f | head
echo "=== data.tar.gz 生成 ==="
cd /tmp/ipkb
echo "2.0" > debian-binary
tar czf data.tar.gz -C data .
echo "data.tar.gz: $(stat -c %s data.tar.gz)"
tar tzf data.tar.gz | head -6
echo "=== control ==="
echo "Package: dockermanager" > control/control
tar czf control.tar.gz -C control .
echo "control.tar.gz: $(stat -c %s control.tar.gz)"
tar tzf control.tar.gz
cat debian-binary control.tar.gz data.tar.gz > /tmp/test2.ipk
echo "ipk: $(stat -c %s /tmp/test2.ipk)"
EOF