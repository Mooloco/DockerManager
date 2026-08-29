package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

// LogStream 是容器日志流选项
type LogStream struct {
	Follow    bool
	Tail      string
	Timestamps bool
	Since     string
	ShowStdout bool
	ShowStderr bool
}

// LogLine 是单条日志(带流类型)
type LogLine struct {
	Stream string `json:"stream"` // stdout | stderr
	Data   string `json:"data"`
}

// StreamContainerLogs 流式读取容器日志。
// fn 返回 false 时停止(客户端断开);阻塞直到日志流结束。
func (c *Client) StreamContainerLogs(ctx context.Context, id string, opts LogStream, fn func(LogLine) bool) error {
	showStdout := true
	showStderr := true
	if !opts.ShowStdout && !opts.ShowStderr {
		showStdout, showStderr = true, true
	}
	if !opts.ShowStdout {
		showStdout = false
	}
	if !opts.ShowStderr {
		showStderr = false
	}

	reader, err := c.cli.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: opts.ShowStdout,
		ShowStderr: opts.ShowStderr,
		Follow:     opts.Follow,
		Tail:       opts.Tail,
		Timestamps: opts.Timestamps,
		Since:      opts.Since,
	})
	if err != nil {
		return wrapDockerError(err)
	}
	defer reader.Close()

	// 检查容器是否 TTY:TTY 容器日志不是多路复用格式,直接当 stdout 读
	insp, err := c.cli.ContainerInspect(ctx, id)
	isTTY := err == nil && insp.Config != nil && insp.Config.Tty

	if isTTY {
		buf := make([]byte, 32*1024)
		for {
			n, rerr := reader.Read(buf)
			if n > 0 {
				if !fn(LogLine{Stream: "stdout", Data: string(buf[:n])}) {
					return nil
				}
			}
			if rerr == io.EOF {
				return nil
			}
			if rerr != nil {
				return rerr
			}
		}
	}

	// 标准多路复用格式:用 StdCopy 分离 stdout/stderr
	if !showStdout {
		// 只要 stderr:StdCopy 需要两个 writer,用一个丢弃 writer
		_, err = stdcopy.StdCopy(io.Discard, &streamWriter{stream: "stderr", fn: fn}, reader)
		return err
	}
	if !showStderr {
		_, err = stdcopy.StdCopy(&streamWriter{stream: "stdout", fn: fn}, io.Discard, reader)
		return err
	}
	_, err = stdcopy.StdCopy(
		&streamWriter{stream: "stdout", fn: fn},
		&streamWriter{stream: "stderr", fn: fn},
		reader,
	)
	return err
}

// streamWriter 把 stdcopy 的输出转换成 LogLine 回调
type streamWriter struct {
	stream string
	fn     func(LogLine) bool
}

func (w *streamWriter) Write(p []byte) (int, error) {
	if !w.fn(LogLine{Stream: w.stream, Data: string(p)}) {
		return 0, io.EOF // 中断
	}
	return len(p), nil
}
