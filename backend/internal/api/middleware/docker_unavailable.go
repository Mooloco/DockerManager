package middleware

import (
	"net/http"

	"github.com/Mooloco/docker-manager/backend/internal/api"
)

// DockerUnavailable 当 Docker 客户端不可用时,
// 所有 API 请求统一返回明确错误(避免 handler 空指针)。
func DockerUnavailable(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.Fail(w, "DOCKER_UNAVAILABLE", "Docker Engine 不可用,请检查 docker.sock 配置与 Docker 服务状态")
	})
}
