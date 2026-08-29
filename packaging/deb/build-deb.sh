#!/usr/bin/env bash
# Docker Manager deb 打包:
#   本机(Windows/git-bash):编译二进制 + 组装 deb 目录
#   测试机(Linux):dpkg-deb 打包(本机无 dpkg-deb)
# 用法: bash packaging/deb/build-deb.sh [版本号]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
VERSION="${1:-0.1.0}"
PKG_NAME="docker-manager"
ARCH="amd64"
STAGE="$ROOT/build/deb-stage"
PKG_FILE="$ROOT/build/${PKG_NAME}_${VERSION}_${ARCH}.deb"
BIN="$ROOT/build/docker-manager-linux-amd64"

echo "==> [1/4] 构建前端并交叉编译 Linux amd64"
(cd "$ROOT/frontend" && npm run build >/dev/null 2>&1)
(cd "$ROOT/backend" && rm -rf internal/web/dist && cp -r "$ROOT/frontend/dist" internal/web/dist)
(cd "$ROOT/backend" && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o "$BIN" ./cmd/server)
ls -lh "$BIN"

echo "==> [2/4] 组装 deb 目录"
rm -rf "$STAGE"
mkdir -p "$STAGE/DEBIAN" \
         "$STAGE/usr/bin" \
         "$STAGE/usr/lib/systemd/system" \
         "$STAGE/etc/docker-manager" \
         "$STAGE/var/lib/docker-manager"

cp "$BIN" "$STAGE/usr/bin/docker-manager"
chmod 755 "$STAGE/usr/bin/docker-manager"   # 确保执行位(cp 会保留源文件 644 权限)
cp "$ROOT/packaging/deb/docker-manager.service" "$STAGE/usr/lib/systemd/system/"

cat > "$STAGE/etc/docker-manager/config.yaml" <<'YAML'
# Docker Manager 配置(默认值;可用同名环境变量覆盖)
server:
  host: 0.0.0.0
  port: 8080

docker:
  host: unix:///var/run/docker.sock

database:
  path: /var/lib/docker-manager/docker-manager.db

compose:
  # Compose 项目 YAML 存放目录(须在 systemd ProtectSystem 的可写区内)
  projects_dir: /var/lib/docker-manager/projects

logging:
  level: info

auth:
  username: admin
  password_env: DM_ADMIN_PASSWORD
  session_ttl_hours: 24
YAML
echo "/etc/docker-manager/config.yaml" > "$STAGE/DEBIAN/conffiles"

cat > "$STAGE/DEBIAN/control" <<CTRL
Package: $PKG_NAME
Version: $VERSION
Section: admin
Priority: optional
Architecture: $ARCH
Depends: docker.io | docker-ce | docker-engine, systemd
Maintainer: Mooloco <mooloco@outlook.com>
Homepage: https://github.com/Mooloco/docker-manager
Description: Lightweight Docker web management console
 Single-binary web console for managing a single Docker Engine node:
 containers, images, networks, volumes, real-time logs and stats.
CTRL

cat > "$STAGE/DEBIAN/postinst" <<'POSTINST'
#!/bin/sh
set -e
if ! id docker-manager >/dev/null 2>&1; then
    useradd --system --home /var/lib/docker-manager --shell /usr/sbin/nologin docker-manager
    usermod -aG docker docker-manager
fi
chmod 755 /usr/bin/docker-manager
chown -R docker-manager:docker /var/lib/docker-manager
chmod 750 /var/lib/docker-manager
chown -R docker-manager:docker /etc/docker-manager 2>/dev/null || true
chmod 640 /etc/docker-manager/config.yaml
if [ "$1" = "configure" ] || [ "$1" = "install" ]; then
    systemctl daemon-reload >/dev/null 2>&1 || true
    systemctl enable docker-manager >/dev/null 2>&1 || true
    echo "Docker Manager 已安装。首次启动前设置密码:"
    echo "  systemctl edit docker-manager   # 添加 Environment=DM_ADMIN_PASSWORD=你的密码"
    echo "  systemctl start docker-manager"
fi
exit 0
POSTINST
chmod 755 "$STAGE/DEBIAN/postinst"

cat > "$STAGE/DEBIAN/prerm" <<'PRERM'
#!/bin/sh
set -e
if [ "$1" = "remove" ] || [ "$1" = "purge" ]; then
    systemctl stop docker-manager >/dev/null 2>&1 || true
    systemctl disable docker-manager >/dev/null 2>&1 || true
fi
exit 0
PRERM
chmod 755 "$STAGE/DEBIAN/prerm"

echo "==> [3/4] 传输到测试机打包"
ssh root@192.168.1.24 "rm -rf /tmp/deb-stage && mkdir -p /tmp/deb-stage"
tar czf - -C "$ROOT/build" deb-stage | ssh root@192.168.1.24 "tar xzf - -C /tmp"
# Windows 上 chmod 无效,须在测试机(Linux)设置权限
ssh root@192.168.1.24 "chmod 755 /tmp/deb-stage/usr/bin/docker-manager /tmp/deb-stage/DEBIAN/postinst /tmp/deb-stage/DEBIAN/prerm && dpkg-deb --build --root-owner-group /tmp/deb-stage /tmp/${PKG_NAME}_${VERSION}_${ARCH}.deb"

echo "==> [4/4] 取回 deb"
scp root@192.168.1.24:/tmp/${PKG_NAME}_${VERSION}_${ARCH}.deb "$PKG_FILE" >/dev/null 2>&1
ssh root@192.168.1.24 "rm -rf /tmp/deb-stage /tmp/${PKG_NAME}_${VERSION}_${ARCH}.deb"

echo ""
echo "完成: $PKG_FILE"
ls -lh "$PKG_FILE"
