package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mooloco/docker-manager/backend/internal/api/routes"
	"github.com/Mooloco/docker-manager/backend/internal/compose"
	"github.com/Mooloco/docker-manager/backend/internal/auth"
	"github.com/Mooloco/docker-manager/backend/internal/config"
	"github.com/Mooloco/docker-manager/backend/internal/database"
	"github.com/Mooloco/docker-manager/backend/internal/docker"
	"github.com/Mooloco/docker-manager/backend/internal/web"
)

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	setupLogger(cfg.Logging.LevelValue())

	// 初始化数据库
	db, err := database.Open(cfg.Database.Path)
	if err != nil {
		slog.Error("初始化数据库失败", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 认证服务 + 初始管理员
	authSvc := auth.NewService(db, &cfg.Auth)
	if err := authSvc.EnsureAdmin(); err != nil {
		slog.Error("初始化管理员账号失败", "error", err)
		os.Exit(1)
	}

	// Docker 客户端(失败不退出:Docker 不可用时服务仍可启动,API 返回明确错误)
	var dc *docker.Client
	if c, err := docker.NewClient(cfg.Docker.Host); err != nil {
		slog.Error("初始化 Docker 客户端失败,服务将在 Docker 不可用状态下运行", "error", err)
	} else {
		dc = c
		defer dc.Close()
	}

	// 前端静态资源(开发阶段 dist 为空时自动降级)
	frontend, err := web.Handler()
	if err != nil {
		slog.Warn("前端资源不可用(开发模式)", "error", err)
	}

	// Compose 项目管理器(仅 Docker 可用时启用)
	var cm *compose.Manager
	if dc != nil {
		if m, err := compose.NewManager(cfg.Compose.ProjectsDir); err != nil {
			slog.Warn("compose 项目管理不可用", "error", err)
		} else {
			cm = m
		}
	}

	// 组装路由
	handler := routes.Router(authSvc, dc, cm, frontend)

	// 定期清理过期 session
	go sessionCleaner(db)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 优雅关闭
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("Docker Manager 已启动", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP 服务异常退出", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("收到退出信号,正在关闭...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	slog.Info("已退出")
}

// setupLogger 按级别配置 slog(JSON 输出,便于日志采集)
func setupLogger(level string) {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	slog.SetDefault(slog.New(handler))
}

// sessionCleaner 每小时清理过期 session
func sessionCleaner(db *database.DB) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		n, err := db.DeleteExpiredSessions()
		if err != nil {
			slog.Warn("清理过期 session 失败", "error", err)
			continue
		}
		if n > 0 {
			slog.Info("已清理过期 session", "count", n)
		}
	}
}
