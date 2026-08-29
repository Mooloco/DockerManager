package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Mooloco/docker-manager/backend/internal/api"
	"github.com/Mooloco/docker-manager/backend/internal/docker"
)

// ContainerHandler 处理容器列表/操作/详情
type ContainerHandler struct {
	docker *docker.Client
}

// NewContainerHandler 创建容器处理器
func NewContainerHandler(dc *docker.Client) *ContainerHandler {
	return &ContainerHandler{docker: dc}
}

// List GET /api/v1/containers?all=true
func (h *ContainerHandler) List(w http.ResponseWriter, r *http.Request) {
	all := true // 默认显示全部
	if v := r.URL.Query().Get("all"); v != "" {
		all, _ = strconv.ParseBool(v)
	}
	items, err := h.docker.ListContainers(r.Context(), all)
	if err != nil {
		api.Fail(w, "DOCKER_ERROR", "获取容器列表失败: "+err.Error())
		return
	}
	api.OK(w, items)
}

// Action POST /api/v1/containers/{id}/{action}
// 支持 start/stop/restart/pause/unpause/kill/remove
func (h *ContainerHandler) Action(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	action := docker.ContainerAction(chi.URLParam(r, "action"))

	// 解析可选参数
	opts := map[string]interface{}{}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&opts)
	}
	// query 参数也支持(force/timeout)
	if v := r.URL.Query().Get("force"); v != "" {
		opts["force"], _ = strconv.ParseBool(v)
	}
	if v := r.URL.Query().Get("timeout"); v != "" {
		opts["timeout"], _ = strconv.ParseFloat(v, 64)
	}

	state, err := h.docker.PerformContainerAction(r.Context(), id, action, opts)
	if err != nil {
		api.Fail(w, "CONTAINER_ACTION_FAILED", "容器操作失败: "+err.Error())
		return
	}
	api.OK(w, map[string]interface{}{
		"action": action,
		"state":  state,
	})
}

// Inspect GET /api/v1/containers/{id}/inspect
func (h *ContainerHandler) Inspect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	insp, err := h.docker.InspectContainer(r.Context(), id)
	if err != nil {
		api.Fail(w, "CONTAINER_NOT_FOUND", "获取容器详情失败: "+err.Error())
		return
	}
	api.OK(w, insp)
}

// Overview GET /api/v1/containers/{id}/overview
// 返回详情页 Overview 所需的扁平数据
func (h *ContainerHandler) Overview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	insp, err := h.docker.InspectContainer(r.Context(), id)
	if err != nil {
		api.Fail(w, "CONTAINER_NOT_FOUND", "获取容器详情失败: "+err.Error())
		return
	}

	overview := map[string]interface{}{
		"id":            insp.ID,
		"name":          strings.TrimPrefix(insp.Name, "/"),
		"image":         insp.Config.Image,
		"created":       insp.Created,
		"started_at":    insp.State.StartedAt,
		"finished_at":   insp.State.FinishedAt,
		"status":        insp.State.Status,
		"state":         insp.State,
		"restart_policy": insp.HostConfig.RestartPolicy.Name,
		"command":       insp.Config.Cmd,
		"entrypoint":    insp.Config.Entrypoint,
		"working_dir":   insp.Config.WorkingDir,
		"tty":           insp.Config.Tty,
		"ports":         insp.NetworkSettings.Ports,
		"mounts":        insp.Mounts,
		"env":           insp.Config.Env,
		"labels":        insp.Config.Labels,
		"network_mode":  insp.HostConfig.NetworkMode,
		"restart_count": insp.RestartCount,
	}
	api.OK(w, overview)
}
