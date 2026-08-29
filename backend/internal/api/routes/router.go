package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Mooloco/docker-manager/backend/internal/api/handlers"
	"github.com/Mooloco/docker-manager/backend/internal/api/middleware"
	"github.com/Mooloco/docker-manager/backend/internal/auth"
	"github.com/Mooloco/docker-manager/backend/internal/compose"
	"github.com/Mooloco/docker-manager/backend/internal/docker"
)

// Router 组装全部 HTTP 路由
// cm 为 nil 时禁用 compose 项目管理功能
func Router(authSvc *auth.Service, dc *docker.Client, cm *compose.Manager, frontend http.Handler) http.Handler {
	r := chi.NewRouter()
	// 全局中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recover)

	// 公开路由
	r.Route("/api/v1/auth", func(r chi.Router) {
		ah := handlers.NewAuthHandler(authSvc)
		r.Post("/login", ah.Login)
		r.With(middleware.RequireAuth(authSvc)).Post("/logout", ah.Logout)
		r.With(middleware.RequireAuth(authSvc)).Get("/me", ah.Me)
		r.With(middleware.RequireAuth(authSvc)).Post("/change-password", ah.ChangePassword)
	})

	// 受保护 API(全部需要登录)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.RequireAuth(authSvc))

		// Docker 不可用时统一返回明确错误
		if dc == nil {
			r.Use(middleware.DockerUnavailable)
			// 占位通配路由:让中间件链生效,DockerUnavailable 会拦截所有请求
			r.Get("/*", noop)
			r.Post("/*", noop)
			r.Delete("/*", noop)
			r.Put("/*", noop)
			return
		}

		sh := handlers.NewSystemHandler(dc)
		r.Get("/system/info", sh.Info)
		r.Get("/system/ping", sh.Ping)

		ch := handlers.NewContainerHandler(dc)
		r.Get("/containers", ch.List)
		r.Post("/containers/{id}/{action}", ch.Action)
		r.Get("/containers/{id}/inspect", ch.Inspect)
		r.Get("/containers/{id}/overview", ch.Overview)

		ih := handlers.NewImageHandler(dc)
		r.Get("/images", ih.List)
		r.Delete("/images/{id}", ih.Remove)

		nh := handlers.NewNetworkHandler(dc)
		r.Get("/networks", nh.List)
		r.Get("/networks/{id}", nh.Inspect)
		r.Delete("/networks/{id}", nh.Remove)

		vh := handlers.NewVolumeHandler(dc)
		r.Get("/volumes", vh.List)
		r.Get("/volumes/{name}", vh.Inspect)
		r.Delete("/volumes/{name}", vh.Remove)

		// Compose 项目
		if cm != nil {
			ph := handlers.NewProjectHandler(dc, cm)
			r.Get("/projects", ph.List)
			r.Get("/projects/{name}", ph.Get)
			r.Get("/projects/{name}/yaml", ph.GetYAML)
			r.Post("/projects", ph.Create)
			r.Put("/projects/{name}", ph.Update)
			r.Post("/projects/{name}/up", ph.Up)
			r.Post("/projects/{name}/stop", ph.Stop)
			r.Post("/projects/{name}/down", ph.Down)
			r.Post("/projects/{name}/restart", ph.Restart)
			r.Post("/projects/{name}/rebuild", ph.Rebuild)
			r.Delete("/projects/{name}", ph.Remove)
			r.Post("/compose/convert", ph.Convert)
		}

		// WebSocket
		wh := handlers.NewWSHandler(dc)
		r.Get("/ws/containers/{id}/logs", wh.Logs)
		r.Get("/ws/containers/{id}/stats", wh.Stats)
		r.Get("/ws/images/pull", wh.Pull)
	})

	// API 404
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 8 && r.URL.Path[:8] == "/api/v1/" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"success":false,"error":{"code":"NOT_FOUND","message":"接口不存在"}}`))
			return
		}
		// 非 API 路径交给前端(SPA)
		if frontend != nil {
			frontend.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})

	return r
}

// noop 是占位 handler(被 DockerUnavailable 中间件拦截,不会真正执行)
func noop(w http.ResponseWriter, _ *http.Request) {}
