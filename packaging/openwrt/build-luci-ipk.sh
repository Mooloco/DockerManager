#!/usr/bin/env bash
# luci-app-dockermanager ipk 打包
# 用法: bash packaging/openwrt/build-luci-ipk.sh [版本号]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
VERSION="${1:-1.3.0}"
OUT="$ROOT/build/luci-app-dockermanager_${VERSION}_all.ipk"
LUCI="$ROOT/packaging/openwrt/luci"

echo "==> [1/3] 上传文件到测试机"
ssh root@192.168.1.24 "rm -rf /tmp/luci-src && mkdir -p /tmp/luci-src"
scp -r "$LUCI/root/." "$LUCI/htdocs/." root@192.168.1.24:/tmp/luci-src/ >/dev/null

echo "==> [2/3] 测试机组装"
ssh root@192.168.1.24 "sh -s" << 'EOF'
set -e
cd /tmp/luci-src
rm -rf /tmp/luci-build && mkdir -p /tmp/luci-build/control /tmp/luci-build/data
cp -a usr /tmp/luci-build/data/
mkdir -p /tmp/luci-build/data/www
cp -a luci-static /tmp/luci-build/data/www/

printf 'Package: luci-app-dockermanager\nVersion: %s\nDepends: luci-base, dockermanager\nSection: luci\nPriority: optional\nArchitecture: all\nMaintainer: Mooloco <mooloco@outlook.com>\nSource: https://github.com/Mooloco/DockerManager\nDescription: LuCI entry for Docker Manager\n Provide a menu entry and status/control page for the Docker Manager\n service (dockermanager).\n' "$VERSION" > /tmp/luci-build/control/control

cd /tmp/luci-build
echo "2.0" > debian-binary
tar czf control.tar.gz -C control .
tar czf data.tar.gz -C data .
tar czf /tmp/luci-app-dockermanager.ipk debian-binary control.tar.gz data.tar.gz
echo "组装完成: $(stat -c %s /tmp/luci-app-dockermanager.ipk) 字节"
EOF

echo "==> [3/3] 取回 ipk"
scp root@192.168.1.24:/tmp/luci-app-dockermanager.ipk "$OUT" >/dev/null
ssh root@192.168.1.24 "rm -rf /tmp/luci-src /tmp/luci-build /tmp/luci-app-dockermanager.ipk"
echo ""
echo "完成: $OUT"
ls -lh "$OUT"
