package web

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

// distFS 嵌入前端构建产物(由 build 阶段生成)
//go:embed all:dist
var distFS embed.FS

// Handler 返回前端静态资源处理器(SPA)。
// 支持 history 路由:未知路径回退到 index.html。
func Handler() (http.Handler, error) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, err
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 只处理 GET/HEAD
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		filePath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if filePath == "" || filePath == "." {
			filePath = "index.html"
		}

		data, err := fs.ReadFile(sub, filePath)
		if err != nil {
			// 静态资源缺失直接 404;其他路径走 SPA 回退
			if strings.HasPrefix(filePath, "assets/") || strings.Contains(filePath, ".") {
				http.NotFound(w, r)
				return
			}
			data, err = fs.ReadFile(sub, "index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else {
			ct := mime.TypeByExtension(path.Ext(filePath))
			if ct == "" {
				ct = "application/octet-stream"
			}
			w.Header().Set("Content-Type", ct)
		}

		w.Header().Set("Cache-Control", cacheControl(filePath))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}), nil
}

// cacheControl 给带 hash 的静态资源设置长缓存,index.html 不缓存
func cacheControl(filePath string) string {
	if filePath == "index.html" {
		return "no-cache"
	}
	return "public, max-age=86400"
}
