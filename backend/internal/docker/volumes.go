package docker

import (
	"context"

	"github.com/docker/docker/api/types/volume"
)

// VolumeItem 是卷列表项
type VolumeItem struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	CreatedAt  string            `json:"created_at"`
	Labels     map[string]string `json:"labels"`
	Scope      string            `json:"scope"`
	// UsageData 可能为 nil(卷未被使用或无法统计)
	UsageData *VolumeUsage `json:"usage_data,omitempty"`
}

// VolumeUsage 卷空间使用信息
type VolumeUsage struct {
	Size     int64 `json:"size"`
	RefCount int64 `json:"ref_count"`
}

// ListVolumes 列出所有卷
func (c *Client) ListVolumes(ctx context.Context) ([]VolumeItem, error) {
	resp, err := c.cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, wrapDockerError(err)
	}
	out := make([]VolumeItem, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		item := VolumeItem{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			CreatedAt:  v.CreatedAt,
			Labels:     v.Labels,
			Scope:      v.Scope,
		}
		if v.UsageData != nil {
			item.UsageData = &VolumeUsage{Size: v.UsageData.Size, RefCount: v.UsageData.RefCount}
		}
		out = append(out, item)
	}
	return out, nil
}

// InspectVolume 查看卷详情
func (c *Client) InspectVolume(ctx context.Context, name string) (*volume.Volume, error) {
	v, err := c.cli.VolumeInspect(ctx, name)
	if err != nil {
		return nil, wrapDockerError(err)
	}
	return &v, nil
}

// RemoveVolume 删除卷;force 强制删除(即使被使用)
func (c *Client) RemoveVolume(ctx context.Context, name string, force bool) error {
	return wrapDockerError(c.cli.VolumeRemove(ctx, name, force))
}
