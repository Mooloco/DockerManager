# Docker Manager — 多阶段构建镜像
# 阶段 1:前端构建(Node.js)
FROM node:22-alpine AS frontend
WORKDIR /build/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install --no-audit --no-fund
COPY frontend/ ./
RUN npm run build

# 阶段 2:Go 后端构建
FROM golang:1.27-alpine AS backend
WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
# 嵌入前端构建产物
COPY --from=frontend /build/frontend/dist ./internal/web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/docker-manager ./cmd/server

# 阶段 3:最小运行镜像
FROM alpine:3.21
# docker-cli + compose 插件:通过挂载的 docker.sock 执行 docker compose 命令
RUN apk add --no-cache ca-certificates tzdata docker-cli docker-cli-compose
WORKDIR /app
COPY --from=backend /out/docker-manager /app/docker-manager
COPY backend/config.example.yaml /app/config.yaml
# 注意:默认以 root 运行,确保能访问宿主机 /var/run/docker.sock。
# 生产加固可取消注释下面两行,并将宿主机 docker 组 gid 映射到容器用户:
#   RUN addgroup -S -g 999 docker-manager && adduser -S -G docker-manager docker-manager \
#       && chown docker-manager:docker-manager /data
#   USER docker-manager
RUN mkdir -p /data
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/app/docker-manager"]
CMD ["-config", "/app/config.yaml"]
