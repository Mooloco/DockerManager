package docker

import (
	"testing"

	"github.com/docker/docker/api/types/container"
)

// buildStats 构造一个最小可用的 StatsResponse
func buildStats(totalUsage, preTotal, sysUsage, preSys uint64, online uint32, memUsage, memLimit uint64) *container.StatsResponse {
	return &container.StatsResponse{
		CPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{TotalUsage: totalUsage, PercpuUsage: []uint64{totalUsage}},
			SystemUsage: sysUsage,
			OnlineCPUs: online,
		},
		PreCPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{TotalUsage: preTotal},
			SystemUsage: preSys,
		},
		MemoryStats: container.MemoryStats{
			Usage: memUsage,
			Limit: memLimit,
			Stats: map[string]uint64{"cache": 0},
		},
	}
}

func TestParseStatsCPU(t *testing.T) {
	// 1 核,2 秒内 CPU 用了 1 秒(1e9 ns),system 用了 2e9 ns
	// cpu% = (1e9/2e9) * 1 * 100 = 50%
	s := buildStats(2e9, 1e9, 4e9, 2e9, 1, 1<<30, 2<<30)
	st, err := parseStats(s)
	if err != nil {
		t.Fatalf("parseStats 失败: %v", err)
	}
	if st.CPU < 49.9 || st.CPU > 50.1 {
		t.Errorf("CPU 应为 ~50%%,实际 %.2f", st.CPU)
	}
}

func TestParseStatsMemory(t *testing.T) {
	s := buildStats(1e9, 0, 2e9, 1e9, 2, 500<<20, 1<<30)
	st, err := parseStats(s)
	if err != nil {
		t.Fatal(err)
	}
	if st.Memory != 500<<20 {
		t.Errorf("内存应为 500MB,实际 %.0f", st.Memory)
	}
	if st.MemPct < 48.7 || st.MemPct > 48.9 {
		t.Errorf("内存百分比应为 ~48.8%%,实际 %.2f", st.MemPct)
	}
}

func TestParseStatsInvalid(t *testing.T) {
	// 全零数据应报错
	_, err := parseStats(&container.StatsResponse{})
	if err == nil {
		t.Error("空 stats 应报错")
	}
}

func TestParseStatsCacheDeduction(t *testing.T) {
	s := buildStats(1e9, 0, 2e9, 1e9, 1, 100<<20, 1<<30)
	s.MemoryStats.Stats["cache"] = 30 << 20 // 30MB cache
	st, err := parseStats(s)
	if err != nil {
		t.Fatal(err)
	}
	// 100MB usage - 30MB cache = 70MB
	if st.Memory != 70<<20 {
		t.Errorf("应扣除 cache,期望 70MB,实际 %.0f", st.Memory)
	}
}

func TestContainerName(t *testing.T) {
	if containerName([]string{"/my-app"}) != "my-app" {
		t.Error("应去掉前导斜杠")
	}
	if containerName(nil) != "" {
		t.Error("空 Names 应返回空")
	}
}

func TestShortID(t *testing.T) {
	id := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if shortID(id) != "0123456789ab" {
		t.Errorf("shortID 应为前 12 位,实际 %s", shortID(id))
	}
	if shortID("abc") != "abc" {
		t.Error("短 ID 不应截断")
	}
}
