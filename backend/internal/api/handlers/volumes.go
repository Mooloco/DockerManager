package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Mooloco/docker-manager/backend/internal/api"
	"github.com/Mooloco/docker-manager/backend/internal/docker"
)

// VolumeHandler 处理卷列表/详情/删除
type VolumeHandler struct {
	docker *docker.Client
}

// NewVolumeHandler 创建卷处理器
func NewVolumeHandler(dc *docker.Client) *VolumeHandler {
	return &VolumeHandler{docker: dc}
}

// List GET /api/v1/volumes
func (h *VolumeHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.docker.ListVolumes(r.Context())
	if err != nil {
		api.Fail(w, "DOCKER_ERROR", "获取卷列表失败: "+err.Error())
		return
	}
	api.OK(w, items)
}

// Inspect GET /api/v1/volumes/{name}
func (h *VolumeHandler) Inspect(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	vol, err := h.docker.InspectVolume(r.Context(), name)
	if err != nil {
		api.Fail(w, "VOLUME_NOT_FOUND", "获取卷详情失败: "+err.Error())
		return
	}
	api.OK(w, vol)
}

// Remove DELETE /api/v1/volumes/{name}?force=true
func (h *VolumeHandler) Remove(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	force := r.URL.Query().Get("force") == "true"
	if err := h.docker.RemoveVolume(r.Context(), name, force); err != nil {
		api.Fail(w, "VOLUME_REMOVE_FAILED", "删除卷失败: "+err.Error())
		return
	}
	api.OK(w, map[string]bool{"removed": true})
}
