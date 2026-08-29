package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

// NetworkItem 是网络列表项
type NetworkItem struct {
	ID      string            `json:"id"`
	ShortID string            `json:"short_id"`
	Name    string            `json:"name"`
	Driver  string            `json:"driver"`
	Scope   string            `json:"scope"`
	Attachable bool           `json:"attachable"`
	Internal   bool           `json:"internal"`
	IPv6    bool              `json:"ipv6"`
	Labels  map[string]string `json:"labels"`
}

// ListNetworks 列出所有网络
func (c *Client) ListNetworks(ctx context.Context) ([]NetworkItem, error) {
	nets, err := c.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, wrapDockerError(err)
	}
	out := make([]NetworkItem, 0, len(nets))
	for _, n := range nets {
		out = append(out, NetworkItem{
			ID:        n.ID,
			ShortID:   shortID(n.ID),
			Name:      n.Name,
			Driver:    n.Driver,
			Scope:     n.Scope,
			Attachable: n.Attachable,
			Internal:  n.Internal,
			IPv6:      n.EnableIPv6,
			Labels:    n.Labels,
		})
	}
	return out, nil
}

// InspectNetwork 查看网络详情(含 IPAM 配置)
func (c *Client) InspectNetwork(ctx context.Context, id string) (*network.Inspect, error) {
	insp, err := c.cli.NetworkInspect(ctx, id, network.InspectOptions{})
	if err != nil {
		return nil, wrapDockerError(err)
	}
	return &insp, nil
}

// RemoveNetwork 删除网络。
// 若网络正被运行中的容器引用,拒绝删除并返回明确提示。
func (c *Client) RemoveNetwork(ctx context.Context, id string) error {
	insp, err := c.cli.NetworkInspect(ctx, id, network.InspectOptions{})
	if err != nil {
		return wrapDockerError(err)
	}
	if len(insp.Containers) > 0 {
		// 一次拉取全部容器,建立 ID → 状态映射
		containers, err := c.cli.ContainerList(ctx, container.ListOptions{All: true})
		if err != nil {
			return wrapDockerError(err)
		}
		stateByID := make(map[string]string, len(containers))
		for _, ct := range containers {
			stateByID[ct.ID] = ct.State
		}
		var running []string
		for cid, ep := range insp.Containers {
			if stateByID[cid] == "running" {
				name := ep.Name
				if name == "" {
					name = shortID(cid)
				}
				running = append(running, name)
			}
		}
		if len(running) > 0 {
			return fmt.Errorf("网络 %s 正被运行中的容器 %s 引用,无法删除;请先停止相关容器", insp.Name, strings.Join(running, "、"))
		}
	}
	return wrapDockerError(c.cli.NetworkRemove(ctx, id))
}
