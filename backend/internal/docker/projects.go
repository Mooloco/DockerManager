package docker

import (
	"context"
	"sort"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

// Compose 标签(compose v2 为每个容器打标)
const (
	labelProject     = "com.docker.compose.project"
	labelWorkingDir  = "com.docker.compose.project.working_dir"
	labelConfigFiles = "com.docker.compose.project.config_files"
	labelService     = "com.docker.compose.service"
	labelOneoff      = "com.docker.compose.oneoff"
)

// Project 是 compose 项目列表项
type Project struct {
	Name        string   `json:"name"`
	Source      string   `json:"source"` // discovered(已有) | managed(本工具创建)
	ConfigFiles []string `json:"config_files"`
	WorkingDir  string   `json:"working_dir"`
	Containers  int      `json:"containers"`
	Running     int      `json:"running"`
	Services    []string `json:"services"`
	// Up 后是否还在运行(managed 项目 down 后无容器,靠此标记)
	HasContainers bool `json:"has_containers"`
}

// ProjectContainer 是项目内单个容器
type ProjectContainer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	State  string `json:"state"`
	Status string `json:"status"`
	Ports  []Port `json:"ports"`
	// IP 是容器在项目网络中的主 IP
	IP string `json:"ip,omitempty"`
}

// ProjectNetwork 是项目使用的网络及连接信息
type ProjectNetwork struct {
	Name      string `json:"name"`
	Driver    string `json:"driver"`
	Internal  bool   `json:"internal"`
	Containers []NetworkMember `json:"containers"`
}

// NetworkMember 是网络中的一个容器
type NetworkMember struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

// ProjectService 是项目内一个服务(含其容器)
type ProjectService struct {
	Name       string             `json:"name"`
	Containers []ProjectContainer `json:"containers"`
}

// ProjectVolume 是项目用到的卷/挂载
type ProjectVolume struct {
	// Type: volume | bind | tmpfs | named_pipe
	Type string `json:"type"`
	// Name: volume 名为卷名;bind 为源路径
	Name string `json:"name"`
	// Destination 容器内挂载点
	Destination string `json:"destination"`
	// Mountpoint 仅 volume 类型:宿主机上的实际位置
	Mountpoint string `json:"mountpoint,omitempty"`
	// Service 属于哪个服务
	Service string `json:"service"`
	RW      bool   `json:"rw"`
}

// ProjectDetail 是项目详情
type ProjectDetail struct {
	Project
	Services []ProjectService `json:"services"`
	Volumes  []ProjectVolume  `json:"volumes"`
	Networks []ProjectNetwork `json:"networks"`
}

// ListProjects 自动发现 compose 项目:扫描容器,按 compose 标签分组
func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	summaries, err := c.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, wrapDockerError(err)
	}

	// 项目名 -> 聚合信息
	type agg struct {
		project     Project
		serviceSet  map[string]struct{}
	}
	groups := make(map[string]*agg)
	var order []string

	for _, s := range summaries {
		proj := s.Labels[labelProject]
		if proj == "" {
			continue // 非 compose 容器
		}
		a, ok := groups[proj]
		if !ok {
			a = &agg{serviceSet: map[string]struct{}{}}
			a.project = Project{
				Name:          proj,
				Source:        "discovered",
				WorkingDir:    s.Labels[labelWorkingDir],
				ConfigFiles:   splitCSV(s.Labels[labelConfigFiles]),
				HasContainers: true,
			}
			groups[proj] = a
			order = append(order, proj)
		}
		a.project.Containers++
		if s.State == "running" {
			a.project.Running++
		}
		if svc := s.Labels[labelService]; svc != "" && s.Labels[labelOneoff] != "True" {
			a.serviceSet[svc] = struct{}{}
		}
	}

	projects := make([]Project, 0, len(order))
	for _, name := range order {
		a := groups[name]
		svcs := make([]string, 0, len(a.serviceSet))
		for svc := range a.serviceSet {
			svcs = append(svcs, svc)
		}
		sort.Strings(svcs)
		a.project.Services = svcs
		projects = append(projects, a.project)
	}
	sort.Slice(projects, func(i, j int) bool { return projects[i].Name < projects[j].Name })
	return projects, nil
}

// GetProjectDetail 获取项目详情:服务、容器、卷信息
func (c *Client) GetProjectDetail(ctx context.Context, name string) (*ProjectDetail, error) {
	summaries, err := c.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, wrapDockerError(err)
	}

	detail := &ProjectDetail{
		Project: Project{Name: name, Source: "discovered", HasContainers: false},
	}
	volSeen := map[string]struct{}{}

	for _, s := range summaries {
		if s.Labels[labelProject] != name {
			continue
		}
		detail.HasContainers = true
		detail.Containers++
		if s.State == "running" {
			detail.Running++
		}
		detail.ConfigFiles = splitCSV(s.Labels[labelConfigFiles])
		if wd := s.Labels[labelWorkingDir]; wd != "" {
			detail.WorkingDir = wd
		}
		svcName := s.Labels[labelService]
		if svcName == "" {
			svcName = "(无服务名)"
		}
		pc := ProjectContainer{
			ID:     s.ID,
			Name:   containerName(s.Names),
			Image:  s.Image,
			State:  s.State,
			Status: s.Status,
			Ports:  portsToUI(s.Ports),
		}

		// 找到或创建服务条目
		idx := -1
		for i := range detail.Services {
			if detail.Services[i].Name == svcName {
				idx = i
				break
			}
		}
		if idx == -1 {
			detail.Services = append(detail.Services, ProjectService{Name: svcName})
			idx = len(detail.Services) - 1
		}
		detail.Services[idx].Containers = append(detail.Services[idx].Containers, pc)

		// 容器挂载与网络信息(Inspect 每个容器)
		if insp, err := c.cli.ContainerInspect(ctx, s.ID); err == nil {
			if insp.Mounts != nil {
				for _, m := range insp.Mounts {
					v := ProjectVolume{
						Type:        string(m.Type),
						Destination: m.Destination,
						Service:     svcName,
						RW:          m.RW,
					}
					switch m.Type {
					case "volume":
						v.Name = m.Name
						v.Mountpoint = m.Source
					default:
						v.Name = m.Source
					}
					key := v.Type + "|" + v.Name + "|" + v.Destination
					if _, seen := volSeen[key]; !seen {
						volSeen[key] = struct{}{}
						detail.Volumes = append(detail.Volumes, v)
					}
				}
			}
			// 网络聚合
			if insp.NetworkSettings != nil && insp.NetworkSettings.Networks != nil {
				ctName := containerName(s.Names)
				for netName, net := range insp.NetworkSettings.Networks {
					idx := -1
					for i := range detail.Networks {
						if detail.Networks[i].Name == netName {
							idx = i
							break
						}
					}
					if idx == -1 {
						detail.Networks = append(detail.Networks, ProjectNetwork{Name: netName})
						idx = len(detail.Networks) - 1
					}
					detail.Networks[idx].Containers = append(detail.Networks[idx].Containers, NetworkMember{
						Name: ctName,
						IP:   net.IPAddress,
					})
					// 容器主 IP(取第一个非空)
					if pc.IP == "" && net.IPAddress != "" {
						pc.IP = net.IPAddress
					}
				}
				// 回写容器 IP 到服务列表
				for i := range detail.Services {
					for j := range detail.Services[i].Containers {
						if detail.Services[i].Containers[j].ID == pc.ID {
							detail.Services[i].Containers[j].IP = pc.IP
						}
					}
				}
			}
		}
	}

	// 补充网络 driver 信息(项目网络数量少,逐个查询可接受)
	for i := range detail.Networks {
		if net, err := c.cli.NetworkInspect(ctx, detail.Networks[i].Name, network.InspectOptions{}); err == nil {
			detail.Networks[i].Driver = net.Driver
			detail.Networks[i].Internal = net.Internal
		}
	}

	// 服务排序
	sort.Slice(detail.Services, func(i, j int) bool { return detail.Services[i].Name < detail.Services[j].Name })
	return detail, nil
}

// splitCSV 拆分逗号分隔的标签值
func splitCSV(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
