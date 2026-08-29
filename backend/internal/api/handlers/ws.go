package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Mooloco/docker-manager/backend/internal/api"
	"github.com/Mooloco/docker-manager/backend/internal/docker"
	"github.com/Mooloco/docker-manager/backend/internal/websocket"
)

// WSHandler 处理 WebSocket 端点
type WSHandler struct {
	docker *docker.Client
}

// NewWSHandler 创建 WebSocket 处理器
func NewWSHandler(dc *docker.Client) *WSHandler {
	return &WSHandler{docker: dc}
}

// Logs WS /api/v1/ws/containers/{id}/logs
// query: follow=true&tail=100&timestamps=true
func (h *WSHandler) Logs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	q := r.URL.Query()

	conn, err := upgrade(w, r)
	if err != nil {
		return
	}

	opts := docker.LogStream{
		Follow:     q.Get("follow") != "false",
		Tail:       q.Get("tail"),
		Timestamps: q.Get("timestamps") == "true",
		ShowStdout: q.Get("stdout") != "false",
		ShowStderr: q.Get("stderr") != "false",
	}
	if opts.Tail == "" {
		opts.Tail = "100"
	}

	websocket.StreamLogs(r.Context(), conn, h.docker, id, opts)
}

// Stats WS /api/v1/ws/containers/{id}/stats
// query: interval=2 (秒)
func (h *WSHandler) Stats(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	conn, err := upgrade(w, r)
	if err != nil {
		return
	}

	interval := 2 * time.Second
	if v := r.URL.Query().Get("interval"); v != "" {
		if sec, err := strconv.Atoi(v); err == nil && sec > 0 {
			interval = time.Duration(sec) * time.Second
		}
	}

	websocket.StreamStats(r.Context(), conn, h.docker, id, interval)
}

// Pull WS /api/v1/ws/images/pull?ref=nginx:latest
func (h *WSHandler) Pull(w http.ResponseWriter, r *http.Request) {
	ref := r.URL.Query().Get("ref")
	if ref == "" {
		api.Fail(w, "INVALID_REQUEST", "缺少 ref 参数(镜像名,如 nginx:latest)")
		return
	}

	conn, err := upgrade(w, r)
	if err != nil {
		return
	}

	websocket.StreamPull(r.Context(), conn, h.docker, ref)
}

// upgrade 升级 HTTP 为 WebSocket
func upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	ws, err := websocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade 失败时它已经写好了 HTTP 错误响应
		return nil, err
	}
	return websocket.NewConn(ws), nil
}
