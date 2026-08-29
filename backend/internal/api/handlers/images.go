package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Mooloco/docker-manager/backend/internal/api"
	"github.com/Mooloco/docker-manager/backend/internal/docker"
)

// ImageHandler 处理镜像列表/删除
type ImageHandler struct {
	docker *docker.Client
}

// NewImageHandler 创建镜像处理器
func NewImageHandler(dc *docker.Client) *ImageHandler {
	return &ImageHandler{docker: dc}
}

// List GET /api/v1/images
func (h *ImageHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.docker.ListImages(r.Context())
	if err != nil {
		api.Fail(w, "DOCKER_ERROR", "获取镜像列表失败: "+err.Error())
		return
	}
	api.OK(w, items)
}

// Remove DELETE /api/v1/images/{id}?force=true
func (h *ImageHandler) Remove(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	force := r.URL.Query().Get("force") == "true"
	if err := h.docker.RemoveImage(r.Context(), id, force); err != nil {
		api.Fail(w, "IMAGE_REMOVE_FAILED", "删除镜像失败: "+err.Error())
		return
	}
	api.OK(w, map[string]bool{"removed": true})
}
