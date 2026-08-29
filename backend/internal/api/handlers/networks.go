package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Mooloco/docker-manager/backend/internal/api"
	"github.com/Mooloco/docker-manager/backend/internal/docker"
)

// NetworkHandler 处理网络列表/详情/删除
type NetworkHandler struct {
	docker *docker.Client
}

// NewNetworkHandler 创建网络处理器
func NewNetworkHandler(dc *docker.Client) *NetworkHandler {
	return &NetworkHandler{docker: dc}
}

// List GET /api/v1/networks
func (h *NetworkHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.docker.ListNetworks(r.Context())
	if err != nil {
		api.Fail(w, "DOCKER_ERROR", "获取网络列表失败: "+err.Error())
		return
	}
	api.OK(w, items)
}

// Inspect GET /api/v1/networks/{id}
func (h *NetworkHandler) Inspect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	net, err := h.docker.InspectNetwork(r.Context(), id)
	if err != nil {
		api.Fail(w, "NETWORK_NOT_FOUND", "获取网络详情失败: "+err.Error())
		return
	}
	api.OK(w, net)
}

// Remove DELETE /api/v1/networks/{id}
func (h *NetworkHandler) Remove(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.docker.RemoveNetwork(r.Context(), id); err != nil {
		api.Fail(w, "NETWORK_REMOVE_FAILED", "删除网络失败: "+err.Error())
		return
	}
	api.OK(w, map[string]bool{"removed": true})
}
