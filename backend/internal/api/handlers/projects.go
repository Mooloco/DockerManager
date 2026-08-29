package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Mooloco/docker-manager/backend/internal/api"
	"github.com/Mooloco/docker-manager/backend/internal/compose"
	"github.com/Mooloco/docker-manager/backend/internal/docker"
)

// ProjectHandler 处理 compose 项目管理
type ProjectHandler struct {
	docker  *docker.Client
	compose *compose.Manager
	cliOK   bool
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler(dc *docker.Client, cm *compose.Manager) *ProjectHandler {
	return &ProjectHandler{
		docker:  dc,
		compose: cm,
		cliOK:   compose.ComposeAvailable(),
	}
}

// projectView 是合并 discovered 与 managed 后的列表视图
type projectView struct {
	docker.Project
	Description string `json:"description,omitempty"`
	ComposeFile string `json:"compose_file,omitempty"`
}

// List GET /api/v1/projects
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	discovered, err := h.docker.ListProjects(r.Context())
	if err != nil {
		api.Fail(w, "DOCKER_ERROR", "获取项目列表失败: "+err.Error())
		return
	}

	// managed 项目(本工具创建)
	managed := h.compose.ListManaged()
	managedSet := map[string]bool{}
	for _, name := range managed {
		managedSet[name] = true
	}

	// discovered 直接作为视图
	views := make([]projectView, 0, len(discovered)+len(managed))
	seen := map[string]bool{}
	for _, p := range discovered {
		// discovered 中已存在的项目,若也在 managed 里,标记为 managed 来源
		if managedSet[p.Name] {
			p.Source = "managed"
		}
		views = append(views, projectView{Project: p})
		seen[p.Name] = true
	}
	// managed 但当前无容器(down 状态)的项目
	for _, name := range managed {
		if !seen[name] {
			views = append(views, projectView{
				Project: docker.Project{
					Name:          name,
					Source:        "managed",
					ConfigFiles:   []string{h.compose.ComposeFile(name)},
					HasContainers: false,
				},
			})
		}
	}

	api.OK(w, views)
}

// Get GET /api/v1/projects/{name} 项目详情
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	detail, err := h.docker.GetProjectDetail(r.Context(), name)
	if err != nil {
		api.Fail(w, "PROJECT_ERROR", "获取项目详情失败: "+err.Error())
		return
	}

	// managed 补充:compose 文件路径与描述
	if _, err := os.Stat(h.compose.ComposeFile(name)); err == nil {
		detail.Source = "managed"
		detail.ConfigFiles = []string{h.compose.ComposeFile(name)}
	}
	if !detail.HasContainers && detail.Source == "" {
		// 无容器且不是 managed → 项目不存在
		api.Fail(w, "PROJECT_NOT_FOUND", "项目不存在或已停止")
		return
	}

	api.OK(w, detail)
}

// GetYAML GET /api/v1/projects/{name}/yaml 返回 compose 文件内容与路径
func (h *ProjectHandler) GetYAML(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	// 优先本工具管理的文件
	if data, err := h.compose.Read(name); err == nil {
		api.OK(w, map[string]interface{}{
			"name":         name,
			"yaml":         data,
			"compose_file": h.compose.ComposeFile(name),
		})
		return
	}

	// 尝试 discovered 项目:从容器标签拿 compose 文件路径并读取
	detail, err := h.docker.GetProjectDetail(r.Context(), name)
	if err != nil || len(detail.ConfigFiles) == 0 {
		api.Fail(w, "PROJECT_NOT_FOUND", "项目不存在或无法定位 compose 文件")
		return
	}
	file := detail.ConfigFiles[0]
	data, err := h.compose.ReadFile(file)
	if err != nil {
		api.Fail(w, "FILE_UNREADABLE", "无法读取 compose 文件("+file+"):"+err.Error())
		return
	}
	api.OK(w, map[string]interface{}{
		"name":         name,
		"yaml":         data,
		"compose_file": file,
	})
}

type createProjectRequest struct {
	Name        string `json:"name"`
	YAML        string `json:"yaml"`
	Description string `json:"description"`
	Start       bool   `json:"start"`
}

// Create POST /api/v1/projects 新建项目(保存 YAML,可选立即 up)
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Fail(w, "INVALID_REQUEST", "请求格式错误")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		api.Fail(w, "INVALID_REQUEST", "项目名不能为空")
		return
	}
	if strings.TrimSpace(req.YAML) == "" {
		api.Fail(w, "INVALID_REQUEST", "compose YAML 不能为空")
		return
	}
	if !h.cliOK {
		api.Fail(w, "COMPOSE_UNAVAILABLE", "docker compose CLI 不可用,无法执行 up/down 操作")
		return
	}

	if err := h.compose.Save(req.Name, req.YAML); err != nil {
		api.Fail(w, "SAVE_FAILED", err.Error())
		return
	}

	result := map[string]interface{}{
		"name":        req.Name,
		"compose_file": h.compose.ComposeFile(req.Name),
		"started":     false,
	}
	if req.Start {
		res, err := h.compose.Up(req.Name, []string{h.compose.ComposeFile(req.Name)})
		if err != nil {
			api.Fail(w, "UP_FAILED", "项目已保存但启动失败: "+err.Error())
			return
		}
		result["started"] = true
		result["output"] = res.Output
	}
	api.Created(w, result)
}

type updateProjectRequest struct {
	YAML        string `json:"yaml"`
	Description string `json:"description"`
}

// Update PUT /api/v1/projects/{name} 更新 compose YAML
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var req updateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Fail(w, "INVALID_REQUEST", "请求格式错误")
		return
	}
	if strings.TrimSpace(req.YAML) == "" {
		api.Fail(w, "INVALID_REQUEST", "compose YAML 不能为空")
		return
	}
	if err := h.compose.Save(name, req.YAML); err != nil {
		api.Fail(w, "SAVE_FAILED", err.Error())
		return
	}
	api.OK(w, map[string]bool{"updated": true})
}

// Up POST /api/v1/projects/{name}/up
func (h *ProjectHandler) Up(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	files, err := h.resolveComposeFiles(r.Context(), name)
	if err != nil {
		api.Fail(w, "PROJECT_ERROR", err.Error())
		return
	}
	if !h.cliOK {
		api.Fail(w, "COMPOSE_UNAVAILABLE", "docker compose CLI 不可用")
		return
	}
	res, err := h.compose.Up(name, files)
	if err != nil {
		api.Fail(w, "UP_FAILED", err.Error())
		return
	}
	api.OK(w, map[string]interface{}{"output": res.Output, "exit_code": res.ExitCode})
}

// Down POST /api/v1/projects/{name}/down
func (h *ProjectHandler) Down(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	files, err := h.resolveComposeFiles(r.Context(), name)
	if err != nil {
		api.Fail(w, "PROJECT_ERROR", err.Error())
		return
	}
	if !h.cliOK {
		api.Fail(w, "COMPOSE_UNAVAILABLE", "docker compose CLI 不可用")
		return
	}
	removeVolumes := false
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
		if v, ok := body["remove_volumes"].(bool); ok {
			removeVolumes = v
		}
	}
	res, err := h.compose.Down(name, files, removeVolumes)
	if err != nil {
		api.Fail(w, "DOWN_FAILED", err.Error())
		return
	}
	api.OK(w, map[string]interface{}{"output": res.Output, "exit_code": res.ExitCode})
}

// Stop POST /api/v1/projects/{name}/stop
// 停止容器但保留容器与网络(compose stop),可用 up 恢复
func (h *ProjectHandler) Stop(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	files, err := h.resolveComposeFiles(r.Context(), name)
	if err != nil {
		api.Fail(w, "PROJECT_ERROR", err.Error())
		return
	}
	if !h.cliOK {
		api.Fail(w, "COMPOSE_UNAVAILABLE", "docker compose CLI 不可用")
		return
	}
	res, err := h.compose.Stop(name, files)
	if err != nil {
		api.Fail(w, "STOP_FAILED", err.Error())
		return
	}
	api.OK(w, map[string]interface{}{"output": res.Output, "exit_code": res.ExitCode})
}

// Restart POST /api/v1/projects/{name}/restart
func (h *ProjectHandler) Restart(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	files, err := h.resolveComposeFiles(r.Context(), name)
	if err != nil {
		api.Fail(w, "PROJECT_ERROR", err.Error())
		return
	}
	if !h.cliOK {
		api.Fail(w, "COMPOSE_UNAVAILABLE", "docker compose CLI 不可用")
		return
	}
	res, err := h.compose.Restart(name, files)
	if err != nil {
		api.Fail(w, "RESTART_FAILED", err.Error())
		return
	}
	api.OK(w, map[string]interface{}{"output": res.Output, "exit_code": res.ExitCode})
}

// Rebuild POST /api/v1/projects/{name}/rebuild
// 重建:compose down → 1 秒 → compose up -d,返回完整日志
func (h *ProjectHandler) Rebuild(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	files, err := h.resolveComposeFiles(r.Context(), name)
	if err != nil {
		api.Fail(w, "PROJECT_ERROR", err.Error())
		return
	}
	if !h.cliOK {
		api.Fail(w, "COMPOSE_UNAVAILABLE", "docker compose CLI 不可用")
		return
	}
	res, err := h.compose.Rebuild(name, files)
	if err != nil {
		api.Fail(w, "REBUILD_FAILED", err.Error())
		return
	}
	api.OK(w, map[string]interface{}{"output": res.Output, "exit_code": res.ExitCode})
}

// Remove DELETE /api/v1/projects/{name}
// 删除项目:执行 docker compose down(停止并删除容器、网络),并删除本工具保存的 compose 文件。
// 已有项目(discovered)只执行 down,不删除宿主机上的 compose 文件。
func (h *ProjectHandler) Remove(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	isManaged := false
	if _, err := os.Stat(h.compose.ComposeFile(name)); err == nil {
		isManaged = true
	}

	// 1. 执行 down(删容器/网络)
	files, err := h.resolveComposeFiles(r.Context(), name)
	if err == nil && h.cliOK {
		if _, err := h.compose.Down(name, files, false); err != nil {
			api.Fail(w, "DOWN_FAILED", "停止并删除容器失败: "+err.Error())
			return
		}
	} else if err != nil {
		api.Fail(w, "PROJECT_ERROR", err.Error())
		return
	}

	// 2. 删除本工具保存的文件(仅 managed)
	if isManaged {
		if err := h.compose.Remove(name); err != nil {
			api.Fail(w, "REMOVE_FAILED", "容器已删除,但移除项目文件失败: "+err.Error())
			return
		}
		api.OK(w, map[string]interface{}{
			"removed":  true,
			"message":  "项目已删除(容器/网络已移除,compose 文件已删除)",
			"file_gone": true,
		})
		return
	}

	// discovered:只 down,不碰宿主机文件
	api.OK(w, map[string]interface{}{
		"removed":   true,
		"message":   "项目容器已删除,宿主机上的 compose 文件保留(由外部管理)",
		"file_gone": false,
	})
}

// Convert POST /api/v1/compose/convert
// 把 docker run 命令转换为 compose YAML
func (h *ProjectHandler) Convert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Fail(w, "INVALID_REQUEST", "请求格式错误")
		return
	}
	if strings.TrimSpace(req.Command) == "" {
		api.Fail(w, "INVALID_REQUEST", "请输入 docker run 命令")
		return
	}
	yamlStr, service, err := compose.ConvertDockerRun(req.Command)
	if err != nil {
		api.Fail(w, "CONVERT_FAILED", err.Error())
		return
	}
	api.OK(w, map[string]interface{}{
		"yaml":    yamlStr,
		"service": service,
	})
}

// resolveComposeFiles 确定 up/down 使用的 compose 文件列表:
// managed 项目用自己的文件;discovered 项目用容器标签里的 config_files
func (h *ProjectHandler) resolveComposeFiles(ctx context.Context, name string) ([]string, error) {
	// 1. managed
	if _, err := os.Stat(h.compose.ComposeFile(name)); err == nil {
		return []string{h.compose.ComposeFile(name)}, nil
	}
	// 2. discovered(从容器标签)
	detail, err := h.docker.GetProjectDetail(ctx, name)
	if err != nil {
		return nil, err
	}
	if len(detail.ConfigFiles) == 0 {
		return nil, fmt.Errorf("无法定位项目的 compose 文件")
	}
	// 过滤不存在的文件
	var files []string
	for _, f := range detail.ConfigFiles {
		if _, err := os.Stat(f); err == nil {
			files = append(files, f)
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("compose 文件不可访问(%s),容器部署时需挂载对应宿主目录", strings.Join(detail.ConfigFiles, ", "))
	}
	return files, nil
}
