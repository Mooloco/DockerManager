package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
)

// ContainerAction 是容器生命周期操作
type ContainerAction string

const (
	ActionStart   ContainerAction = "start"
	ActionStop    ContainerAction = "stop"
	ActionRestart ContainerAction = "restart"
	ActionPause   ContainerAction = "pause"
	ActionUnpause ContainerAction = "unpause"
	ActionKill    ContainerAction = "kill"
	ActionRemove  ContainerAction = "remove"
)

// PerformContainerAction 执行容器操作并返回操作后的最新状态
func (c *Client) PerformContainerAction(ctx context.Context, id string, action ContainerAction, opts map[string]interface{}) (string, error) {
	var err error
	switch action {
	case ActionStart:
		err = c.cli.ContainerStart(ctx, id, container.StartOptions{})
	case ActionStop:
		timeout := 10
		if v, ok := opts["timeout"].(float64); ok {
			timeout = int(v)
		}
		err = c.cli.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout})
	case ActionRestart:
		timeout := 10
		if v, ok := opts["timeout"].(float64); ok {
			timeout = int(v)
		}
		err = c.cli.ContainerRestart(ctx, id, container.StopOptions{Timeout: &timeout})
	case ActionPause:
		err = c.cli.ContainerPause(ctx, id)
	case ActionUnpause:
		err = c.cli.ContainerUnpause(ctx, id)
	case ActionKill:
		signal := "SIGKILL"
		if v, ok := opts["signal"].(string); ok && v != "" {
			signal = v
		}
		err = c.cli.ContainerKill(ctx, id, signal)
	case ActionRemove:
		force := false
		if v, ok := opts["force"].(bool); ok {
			force = v
		}
		removeVolumes := false
		if v, ok := opts["remove_volumes"].(bool); ok {
			removeVolumes = v
		}
		err = c.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: force, RemoveVolumes: removeVolumes})
	default:
		return "", fmt.Errorf("不支持的容器操作: %s", action)
	}
	if err != nil {
		return "", wrapDockerError(err)
	}

	// 操作后查询最新状态
	insp, err := c.cli.ContainerInspect(ctx, id)
	if err != nil {
		// 容器可能已被删除
		if errdefs.IsNotFound(err) {
			return "removed", nil
		}
		return "", wrapDockerError(err)
	}
	return insp.State.Status, nil
}

// InspectContainer 返回容器完整 inspect 数据
func (c *Client) InspectContainer(ctx context.Context, id string) (*container.InspectResponse, error) {
	insp, err := c.cli.ContainerInspect(ctx, id)
	if err != nil {
		return nil, wrapDockerError(err)
	}
	return &insp, nil
}

// wrapDockerError 将 Docker SDK 错误转换为友好的错误信息
func wrapDockerError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errdefs.IsNotFound(err):
		return fmt.Errorf("资源不存在(可能已被删除)")
	case errdefs.IsConflict(err):
		return fmt.Errorf("操作冲突(容器状态不允许该操作)")
	case errdefs.IsForbidden(err):
		return fmt.Errorf("操作被拒绝(权限不足)")
	case errdefs.IsUnauthorized(err):
		return fmt.Errorf("Docker 认证失败")
	default:
		return fmt.Errorf("%s", strings.TrimSpace(err.Error()))
	}
}
