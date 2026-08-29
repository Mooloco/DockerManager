package websocket

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Upgrader 是 WebSocket 升级器(统一配置)
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	// 允许跨域(管理工具可能从其他来源访问)
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Conn 包装 WebSocket 连接,提供线程安全的写操作
type Conn struct {
	ws   *websocket.Conn
	mu   sync.Mutex
	done chan struct{}
}

// NewConn 包装一个已升级的连接
func NewConn(ws *websocket.Conn) *Conn {
	return &Conn{ws: ws, done: make(chan struct{})}
}

// WriteJSON 线程安全地写 JSON 消息
func (c *Conn) WriteJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	select {
	case <-c.done:
		return websocket.ErrCloseSent
	default:
	}
	_ = c.ws.SetWriteDeadline(time.Now().Add(15 * time.Second))
	return c.ws.WriteJSON(v)
}

// Close 关闭连接(幂等)
func (c *Conn) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	select {
	case <-c.done:
		return
	default:
		close(c.done)
		_ = c.ws.Close()
	}
}

// Done 返回连接关闭信号
func (c *Conn) Done() <-chan struct{} {
	return c.done
}

// WSMessage 是所有 WS 消息的统一信封
type WSMessage struct {
	Type    string      `json:"type"` // log | stats | pull | error | end
	Stream  string      `json:"stream,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Log 记录 WebSocket 事件
func Log(action string, detail string) {
	slog.Debug("websocket", "action", action, "detail", detail)
}
