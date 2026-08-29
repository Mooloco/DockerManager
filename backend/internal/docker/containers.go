package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
)

// ContainerItem 是容器列表项(含实时 CPU/内存)
type ContainerItem struct {
	ID      string   `json:"id"`
	ShortID string   `json:"short_id"`
	Name    string   `json:"name"`
	Image   string   `json:"image"`
	ImageID string   `json:"image_id"`
	State   string   `json:"state"`
	Status  string   `json:"status"`
	Created int64    `json:"created"`
	Ports   []Port   `json:"ports"`
	CPU     float64  `json:"cpu_percent"`
	Memory  float64  `json:"memory_bytes"`
	MemPct  float64  `json:"memory_percent"`
	PIDs    int64    `json:"pids"`
	NetIO   []uint64 `json:"net_io"` // [rx, tx]
	BlockIO []uint64 `json:"block_io"` // [read, write]
}

// Port 是端口映射的 UI 友好结构
type Port struct {
	IP          string `json:"ip"`
	PrivatePort uint16 `json:"private_port"`
	PublicPort  uint16 `json:"public_port"`
	Type        string `json:"type"`
}

// ListContainers 列出容器(含一次性实时 stats)
func (c *Client) ListContainers(ctx context.Context, all bool) ([]ContainerItem, error) {
	summaries, err := c.cli.ContainerList(ctx, container.ListOptions{All: all})
	if err != nil {
		return nil, err
	}

	items := make([]ContainerItem, 0, len(summaries))
	for _, s := range summaries {
		item := ContainerItem{
			ID:      s.ID,
			ShortID: shortID(s.ID),
			Name:    containerName(s.Names),
			Image:   s.Image,
			ImageID: s.ImageID,
			State:   s.State,
			Status:  s.Status,
			Created: s.Created,
			Ports:   portsToUI(s.Ports),
		}
		items = append(items, item)
	}

	// 批量拉取实时 stats(单个容器失败不影响整体列表)
	if stats, err := c.statsBatch(ctx, items); err == nil {
		for i := range items {
			if st, ok := stats[items[i].ID]; ok {
				items[i].CPU = st.CPU
				items[i].Memory = st.Memory
				items[i].MemPct = st.MemPct
				items[i].PIDs = st.PIDs
				items[i].NetIO = st.NetIO
				items[i].BlockIO = st.BlockIO
			}
		}
	}

	// 按创建时间倒序
	sort.Slice(items, func(i, j int) bool { return items[i].Created > items[j].Created })
	return items, nil
}

// statsEntry 是一次性 stats 的解析结果
type statsEntry struct {
	CPU     float64
	Memory  float64
	MemPct  float64
	PIDs    int64
	NetIO   []uint64
	BlockIO []uint64
}

// statsBatch 并发获取多个容器的实时 stats
func (c *Client) statsBatch(ctx context.Context, items []ContainerItem) (map[string]statsEntry, error) {
	type result struct {
		id  string
		st  statsEntry
		err error
	}
	ch := make(chan result, len(items))
	sem := make(chan struct{}, 8) // 限制并发

	for _, it := range items {
		go func(id string) {
			sem <- struct{}{}
			defer func() { <-sem }()
			// 每个容器独立超时,避免一个卡死拖垮全部
			sctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			st, err := c.containerStatsOnce(sctx, id)
			ch <- result{id: id, st: st, err: err}
		}(it.ID)
	}

	out := make(map[string]statsEntry, len(items))
	for range items {
		r := <-ch
		if r.err == nil {
			out[r.id] = r.st
		}
	}
	close(ch)
	return out, nil
}

// containerStatsOnce 获取单个容器的一次性 stats 并解析
func (c *Client) containerStatsOnce(ctx context.Context, id string) (statsEntry, error) {
	resp, err := c.cli.ContainerStatsOneShot(ctx, id)
	if err != nil {
		return statsEntry{}, err
	}
	defer resp.Body.Close()

	var st container.StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		return statsEntry{}, err
	}
	return parseStats(&st)
}

// containerName 从 Names 数组中取可读名称(去掉前导 /)
func containerName(names []string) string {
	if len(names) == 0 {
		return ""
	}
	n := names[0]
	if strings.HasPrefix(n, "/") {
		n = strings.TrimPrefix(n, "/")
	}
	return n
}

func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// portsToUI 转换 SDK 端口结构
func portsToUI(ports []container.Port) []Port {
	out := make([]Port, 0, len(ports))
	for _, p := range ports {
		out = append(out, Port{
			IP:          p.IP,
			PrivatePort: p.PrivatePort,
			PublicPort:  p.PublicPort,
			Type:        p.Type,
		})
	}
	return out
}

// ContainerStats 是单容器实时 stats(详情页使用)
type ContainerStats struct {
	CPU    float64 `json:"cpu_percent"`
	Memory float64 `json:"memory_bytes"`
	MemPct float64 `json:"memory_percent"`
	MemLimit float64 `json:"memory_limit"`
	PIDs   int64   `json:"pids"`
	NetIO  []uint64 `json:"net_io"`
	BlockIO []uint64 `json:"block_io"`
}

// GetContainerStats 获取单容器一次性 stats
func (c *Client) GetContainerStats(ctx context.Context, id string) (*ContainerStats, error) {
	resp, err := c.cli.ContainerStatsOneShot(ctx, id)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var st container.StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		return nil, err
	}
	parsed, err := parseStats(&st)
	if err != nil {
		return nil, err
	}
	limit := float64(st.MemoryStats.Limit)
	if limit == 0 {
		limit = parsed.Memory
	}
	return &ContainerStats{
		CPU:      parsed.CPU,
		Memory:   parsed.Memory,
		MemPct:   parsed.MemPct,
		MemLimit: limit,
		PIDs:     parsed.PIDs,
		NetIO:    parsed.NetIO,
		BlockIO:  parsed.BlockIO,
	}, nil
}

// parseStats 从 Docker stats 原始数据计算 CPU/内存等指标
func parseStats(s *container.StatsResponse) (statsEntry, error) {
	if s == nil || s.CPUStats.CPUUsage.TotalUsage == 0 || s.MemoryStats.Usage == 0 {
		return statsEntry{}, fmt.Errorf("无效的 stats 数据")
	}

	// CPU 使用率: (cpu_delta / system_delta) * online_cpus * 100
	var cpuPct float64
	if s.CPUStats.SystemUsage > 0 {
		cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage - s.PreCPUStats.CPUUsage.TotalUsage)
		sysDelta := float64(s.CPUStats.SystemUsage - s.PreCPUStats.SystemUsage)
		online := uint32(0)
		if s.CPUStats.OnlineCPUs > 0 {
			online = s.CPUStats.OnlineCPUs
		} else {
			online = uint32(len(s.CPUStats.CPUUsage.PercpuUsage))
		}
		if online == 0 {
			online = 1
		}
		if sysDelta > 0 && cpuDelta > 0 {
			cpuPct = (cpuDelta / sysDelta) * float64(online) * 100
		}
	}

	memUsage := float64(s.MemoryStats.Usage)
	// 扣除 cache 部分(与 docker stats 显示一致)
	if cache, ok := s.MemoryStats.Stats["cache"]; ok {
		if c := float64(cache); c > 0 && memUsage > c {
			memUsage -= c
		}
	}
	var memPct float64
	limit := float64(s.MemoryStats.Limit)
	if limit > 0 {
		memPct = memUsage / limit * 100
	}

	netIO := []uint64{0, 0}
	if s.Networks != nil {
		for _, n := range s.Networks {
			netIO[0] += n.RxBytes
			netIO[1] += n.TxBytes
		}
	}
	blockIO := []uint64{0, 0}
	if len(s.BlkioStats.IoServiceBytesRecursive) > 0 {
		for _, b := range s.BlkioStats.IoServiceBytesRecursive {
			switch b.Op {
			case "Read":
				blockIO[0] += b.Value
			case "Write":
				blockIO[1] += b.Value
			}
		}
	}

	return statsEntry{
		CPU:     round2(cpuPct),
		Memory:  memUsage,
		MemPct:  round2(memPct),
		PIDs:    int64(s.PidsStats.Current),
		NetIO:   netIO,
		BlockIO: blockIO,
	}, nil
}

func round2(f float64) float64 {
	return float64(int64(f*100+0.5)) / 100
}
