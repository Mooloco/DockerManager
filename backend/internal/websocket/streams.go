package websocket

import (
	"context"
	"log/slog"
	"time"

	"github.com/Mooloco/docker-manager/backend/internal/docker"
)

// StreamLogs 把容器日志流转发到 WebSocket。
// 阻塞直到日志流结束或客户端断开。
func StreamLogs(ctx context.Context, conn *Conn, dc *docker.Client, id string, opts docker.LogStream) {
	logCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 客户端断开时取消日志流
	go func() {
		select {
		case <-conn.Done():
			cancel()
		case <-logCtx.Done():
		}
	}()

	err := dc.StreamContainerLogs(logCtx, id, opts, func(line docker.LogLine) bool {
		select {
		case <-logCtx.Done():
			return false
		default:
		}
		return conn.WriteJSON(WSMessage{Type: "log", Stream: line.Stream, Data: line.Data}) == nil
	})

	// 发送结束标记
	end := WSMessage{Type: "end"}
	if err != nil {
		end.Message = err.Error()
	}
	_ = conn.WriteJSON(end)
	conn.Close()
}

// StreamStats 定时推送容器实时 stats。
// interval 为推送间隔(秒);阻塞直到客户端断开。
func StreamStats(ctx context.Context, conn *Conn, dc *docker.Client, id string, interval time.Duration) {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 先推一次
	push := func() {
		st, err := dc.GetContainerStats(ctx, id)
		if err != nil {
			_ = conn.WriteJSON(WSMessage{Type: "error", Message: err.Error()})
			return
		}
		if conn.WriteJSON(WSMessage{Type: "stats", Data: st}) != nil {
			cancelAndClose(conn)
		}
	}
	push()

	for {
		select {
		case <-conn.Done():
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			push()
		}
	}
}

// StreamPull 把镜像拉取进度流转发到 WebSocket。
// 阻塞直到拉取完成或客户端断开。
func StreamPull(ctx context.Context, conn *Conn, dc *docker.Client, ref string) {
	pullCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-conn.Done():
			cancel()
		case <-pullCtx.Done():
		}
	}()

	reader, err := dc.PullImage(pullCtx, ref)
	if err != nil {
		_ = conn.WriteJSON(WSMessage{Type: "error", Message: err.Error()})
		conn.Close()
		return
	}

	streamDone := make(chan error, 1)
	go func() {
		streamDone <- docker.ReadPullEvents(reader, func(ev docker.PullEvent) bool {
			select {
			case <-pullCtx.Done():
				return false
			default:
			}
			return conn.WriteJSON(WSMessage{Type: "pull", Data: ev}) == nil
		})
	}()

	select {
	case err := <-streamDone:
		if err != nil {
			_ = conn.WriteJSON(WSMessage{Type: "error", Message: err.Error()})
		} else {
			_ = conn.WriteJSON(WSMessage{Type: "end", Data: map[string]string{"ref": ref}})
		}
	case <-conn.Done():
		cancel()
		<-streamDone
	case <-pullCtx.Done():
		<-streamDone
	}
	conn.Close()
}

// cancelAndClose 断开连接并取消
func cancelAndClose(conn *Conn) {
	conn.Close()
}

// logWS 记录日志流事件
func logWS(action, detail string) {
	slog.Debug("ws-stream", "action", action, "detail", detail)
}
