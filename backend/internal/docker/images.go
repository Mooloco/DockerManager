package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/docker/docker/api/types/image"
)

// ImageItem 是镜像列表项
type ImageItem struct {
	ID          string   `json:"id"`
	ShortID     string   `json:"short_id"`
	RepoTags    []string `json:"repo_tags"`
	RepoDigests []string `json:"repo_digests"`
	Created     int64    `json:"created"`
	Size        int64    `json:"size"`
	SharedSize  int64    `json:"shared_size"`
	VirtualSize int64    `json:"virtual_size"`
	Containers  int      `json:"containers"`
}

// ListImages 列出所有镜像
func (c *Client) ListImages(ctx context.Context) ([]ImageItem, error) {
	summaries, err := c.cli.ImageList(ctx, image.ListOptions{All: true})
	if err != nil {
		return nil, wrapDockerError(err)
	}
	out := make([]ImageItem, 0, len(summaries))
	for _, s := range summaries {
		out = append(out, ImageItem{
			ID:          s.ID,
			ShortID:     shortImageID(s.ID),
			RepoTags:    s.RepoTags,
			RepoDigests: s.RepoDigests,
			Created:     s.Created,
			Size:        s.Size,
			SharedSize:  s.SharedSize,
			VirtualSize: s.VirtualSize,
			Containers:  int(s.Containers),
		})
	}
	return out, nil
}

// RemoveImage 删除镜像;force 强制删除
func (c *Client) RemoveImage(ctx context.Context, id string, force bool) error {
	_, err := c.cli.ImageRemove(ctx, id, image.RemoveOptions{Force: force})
	return wrapDockerError(err)
}

// PullImage 拉取镜像,返回 JSON 进度流(调用方负责关闭)。
// 流的每一行是一个 Docker 进度对象,如 {"status":"Pulling from nginx",...}。
func (c *Client) PullImage(ctx context.Context, ref string) (io.ReadCloser, error) {
	// 规范化引用:不带 tag 时补 :latest
	if !strings.Contains(ref, ":") && !strings.Contains(ref, "@") {
		ref = ref + ":latest"
	}
	reader, err := c.cli.ImagePull(ctx, ref, image.PullOptions{})
	if err != nil {
		return nil, wrapDockerError(err)
	}
	return reader, nil
}

// PullEvent 是拉取进度事件(UI 友好的扁平结构)
type PullEvent struct {
	Status         string `json:"status"`
	Progress       string `json:"progress"`
	ProgressDetail struct {
		Current int64 `json:"current"`
		Total   int64 `json:"total"`
	} `json:"progress_detail"`
	ID       string `json:"id"`
	Error    string `json:"error"`
	Stream   string `json:"stream"`
	Complete bool   `json:"complete"`
}

// ReadPullEvents 读取拉取进度流并逐条解码为事件。
// fn 返回 false 时停止读取(客户端断开)。
func ReadPullEvents(r io.ReadCloser, fn func(PullEvent) bool) error {
	defer r.Close()
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var raw struct {
			Status         string `json:"status"`
			Progress       string `json:"progress"`
			ProgressDetail struct {
				Current int64 `json:"current"`
				Total   int64 `json:"total"`
			} `json:"progressDetail"`
			ID     string `json:"id"`
			Error  string `json:"error"`
			Stream string `json:"stream"`
		}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue // 跳过无法解析的行
		}
		ev := PullEvent{
			Status:   raw.Status,
			Progress: raw.Progress,
			ID:       raw.ID,
			Error:    raw.Error,
			Stream:   raw.Stream,
		}
		ev.ProgressDetail = raw.ProgressDetail
		if !fn(ev) {
			return nil
		}
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func shortImageID(id string) string {
	if strings.HasPrefix(id, "sha256:") {
		id = strings.TrimPrefix(id, "sha256:")
	}
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
