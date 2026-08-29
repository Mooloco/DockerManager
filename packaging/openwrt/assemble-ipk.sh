#!/bin/sh
# 在目标机(Linux)上组装 ipk。用法: assemble-ipk.sh <版本号> <输出路径>
# 前置:/tmp/ipk-src/{usr/bin/dockermanager, etc/init.d/dockermanager, etc/config/dockermanager}
set -e

VERSION="$1"
OUT="$2"

cd /tmp/ipk-src
chmod 755 usr/bin/dockermanager etc/init.d/dockermanager
chmod 644 etc/config/dockermanager

rm -rf /tmp/ipk-build
mkdir -p /tmp/ipk-build/control /tmp/ipk-build/data
cp -a usr etc /tmp/ipk-build/data/

# control 文件(用 printf 避免 heredoc 嵌套)
printf 'Package: dockermanager\nVersion: %s\nDepends: dockerd, docker, docker-compose, libc\nSection: admin\nPriority: optional\nArchitecture: x86_64\nMaintainer: Mooloco <mooloco@outlook.com>\nSource: https://github.com/Mooloco/DockerManager\nDescription: Docker Manager - lightweight Docker web console\n Lightweight web console for managing Docker containers, images, networks,\n volumes and Docker Compose projects. Single static binary, no runtime deps.\n' "$VERSION" > /tmp/ipk-build/control/control

echo "/etc/config/dockermanager" > /tmp/ipk-build/control/conffiles

cd /tmp/ipk-build
echo "2.0" > debian-binary
tar czf control.tar.gz -C control .
tar czf data.tar.gz -C data .
# 现代 OpenWrt ipk = gzip 压缩的 tar,内含 debian-binary + control.tar.gz + data.tar.gz
tar czf "$OUT" debian-binary control.tar.gz data.tar.gz

echo "组装完成: $OUT ($(stat -c %s "$OUT") 字节)"
