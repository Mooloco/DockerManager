#!/usr/bin/env bash
# Docker Manager OpenWrt ipk 打包(x86_64)
# 用法: bash packaging/openwrt/build-ipk.sh [版本号]
# 组装脚本独立 scp 执行,避免嵌套 heredoc 解析问题
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
VERSION="${1:-1.3.0}"
OUT="$ROOT/build/dockermanager_${VERSION}_x86_64.ipk"
BIN="$ROOT/build/docker-manager-linux-amd64"

[ -f "$BIN" ] || { echo "缺少 $BIN,请先交叉编译" >&2; exit 1; }

echo "==> [1/3] 上传文件到测试机"
ssh root@192.168.1.24 "rm -rf /tmp/ipk-src && mkdir -p /tmp/ipk-src/usr/bin /tmp/ipk-src/etc/init.d /tmp/ipk-src/etc/config"
scp "$BIN" root@192.168.1.24:/tmp/ipk-src/usr/bin/dockermanager >/dev/null
scp "$ROOT/packaging/openwrt/root/etc/init.d/dockermanager" root@192.168.1.24:/tmp/ipk-src/etc/init.d/dockermanager >/dev/null
scp "$ROOT/packaging/openwrt/root/etc/config/dockermanager" root@192.168.1.24:/tmp/ipk-src/etc/config/dockermanager >/dev/null
scp "$ROOT/packaging/openwrt/assemble-ipk.sh" root@192.168.1.24:/tmp/assemble-ipk.sh >/dev/null

echo "==> [2/3] 测试机组装"
ssh root@192.168.1.24 "sh /tmp/assemble-ipk.sh '$VERSION' /tmp/dockermanager.ipk"

echo "==> [3/3] 取回 ipk"
scp root@192.168.1.24:/tmp/dockermanager.ipk "$OUT" >/dev/null
ssh root@192.168.1.24 "rm -rf /tmp/ipk-src /tmp/ipk-build /tmp/assemble-ipk.sh /tmp/dockermanager.ipk"
echo ""
echo "完成: $OUT"
ls -lh "$OUT"
