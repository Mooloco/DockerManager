package docker

import (
	"context"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
)

// SystemInfo 是 Dashboard 需要展示的 Docker 基本信息(适合 UI 的扁平结构)
type SystemInfo struct {
	ServerVersion    string `json:"server_version"`
	APIVersion       string `json:"api_version"`
	MinAPIVersion    string `json:"min_api_version"`
	OSType           string `json:"os_type"`
	OperatingSystem  string `json:"operating_system"`
	Architecture     string `json:"architecture"`
	KernelVersion    string `json:"kernel_version"`
	Driver           string `json:"driver"`
	RootDir          string `json:"root_dir"`
	MemoryLimit      bool   `json:"memory_limit"`
	SwapLimit        bool   `json:"swap_limit"`
	CPUs             int    `json:"cpus"`
	TotalMemory      int64  `json:"total_memory"`
	Name             string `json:"name"`
	Containers       int    `json:"containers"`
	ContainersRunning int   `json:"containers_running"`
	ContainersPaused int    `json:"containers_paused"`
	ContainersStopped int   `json:"containers_stopped"`
	Images           int    `json:"images"`
}

// GetSystemInfo 汇总 Docker 引擎信息与资源统计
func (c *Client) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	info, err := c.cli.Info(ctx)
	if err != nil {
		return nil, err
	}
	images, err := c.cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, err
	}
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	var running, paused, stopped int
	for _, ct := range containers {
		switch ct.State {
		case "running":
			running++
		case "paused":
			paused++
		default:
			stopped++
		}
	}

	return &SystemInfo{
		ServerVersion:      info.ServerVersion,
		APIVersion:         c.cli.ClientVersion(),
		MinAPIVersion:      "",
		OSType:             info.OSType,
		OperatingSystem:   info.OperatingSystem,
		Architecture:      info.Architecture,
		KernelVersion:     info.KernelVersion,
		Driver:            info.Driver,
		RootDir:           info.DockerRootDir,
		MemoryLimit:       info.MemoryLimit,
		SwapLimit:         info.SwapLimit,
		CPUs:              info.NCPU,
		TotalMemory:       info.MemTotal,
		Name:              info.Name,
		Containers:        info.Containers,
		ContainersRunning: info.ContainersRunning,
		ContainersPaused:  info.ContainersPaused,
		ContainersStopped: info.ContainersStopped,
		Images:            len(images),
	}, nil
}

// Events 订阅 Docker 事件流(供实时状态刷新使用)
func (c *Client) Events(ctx context.Context, options events.ListOptions) (<-chan events.Message, <-chan error) {
	return c.cli.Events(ctx, options)
}
