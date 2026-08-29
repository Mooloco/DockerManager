package docker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/docker/docker/client"
)

// Client 是对 Docker SDK 的薄封装,
// 业务代码统一通过本包的 Service 层访问,不直接触碰 SDK。
type Client struct {
	cli *client.Client
}

// NewClient 创建 Docker 客户端。
// host 支持 unix:///var/run/docker.sock 或 tcp://host:port 等标准 DOCKER_HOST 格式。
func NewClient(host string) (*Client, error) {
	opts := []client.Opt{
		client.WithAPIVersionNegotiation(),
	}
	if host != "" {
		opts = append(opts, client.WithHost(host))
	}
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("创建 Docker 客户端失败: %w", err)
	}

	// 验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := cli.Ping(ctx); err != nil {
		return nil, fmt.Errorf("无法连接 Docker Engine (%s): %w", host, err)
	}

	slog.Info("Docker Engine 连接成功", "host", host)
	return &Client{cli: cli}, nil
}

// Close 关闭底层连接
func (c *Client) Close() error {
	return c.cli.Close()
}

// Raw 返回底层 SDK 客户端(仅限需要底层能力的内部场景)
func (c *Client) Raw() *client.Client {
	return c.cli
}
