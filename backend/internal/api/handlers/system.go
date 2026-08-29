package handlers

import (
	"net/http"

	"github.com/Mooloco/docker-manager/backend/internal/api"
	"github.com/Mooloco/docker-manager/backend/internal/docker"
)

// SystemHandler 提供 Dashboard 系统信息
type SystemHandler struct {
	docker *docker.Client
}

// NewSystemHandler 创建系统信息处理器
func NewSystemHandler(dc *docker.Client) *SystemHandler {
	return &SystemHandler{docker: dc}
}

// Info GET /api/v1/system/info
func (h *SystemHandler) Info(w http.ResponseWriter, r *http.Request) {
	info, err := h.docker.GetSystemInfo(r.Context())
	if err != nil {
		api.Fail(w, "DOCKER_UNAVAILABLE", "无法获取 Docker 信息: "+err.Error())
		return
	}
	api.OK(w, info)
}

// Ping GET /api/v1/system/ping
func (h *SystemHandler) Ping(w http.ResponseWriter, r *http.Request) {
	_, err := h.docker.Raw().Ping(r.Context())
	if err != nil {
		api.Fail(w, "DOCKER_UNAVAILABLE", "Docker Engine 不可达")
		return
	}
	api.OK(w, map[string]bool{"ok": true})
}
