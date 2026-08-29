package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmbedContent(t *testing.T) {
	entries, err := fs.ReadDir(distFS, "dist")
	if err != nil {
		t.Fatalf("读取 embed dist 失败: %v", err)
	}
	for _, e := range entries {
		t.Logf("dist 内容: %s (dir=%v)", e.Name(), e.IsDir())
	}
	data, err := fs.ReadFile(distFS, "dist/index.html")
	if err != nil {
		t.Fatalf("读取 index.html 失败: %v", err)
	}
	if !strings.Contains(string(data), "<div id=\"app\">") {
		t.Errorf("index.html 内容异常: %s", string(data[:200]))
	}
}

func TestHandlerServesIndex(t *testing.T) {
	h, err := Handler()
	if err != nil {
		t.Fatalf("Handler() 失败: %v", err)
	}
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	t.Logf("GET / → code=%d headers=%v body=%q", rec.Code, rec.Header(), rec.Body.String()[:min(len(rec.Body.String()), 150)])
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / 返回 %d,期望 200", rec.Code)
	}

	// SPA 回退
	req2 := httptest.NewRequest("GET", "/containers", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	t.Logf("GET /containers → code=%d", rec2.Code)
	if rec2.Code != http.StatusOK {
		t.Fatalf("GET /containers 返回 %d,期望 200", rec2.Code)
	}
}
