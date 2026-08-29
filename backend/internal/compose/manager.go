package compose

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Manager 负责 compose 项目的文件存储与 CLI 执行
type Manager struct {
	projectsDir string
}

// NewManager 创建 compose 管理器
func NewManager(projectsDir string) (*Manager, error) {
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建项目目录失败: %w", err)
	}
	return &Manager{projectsDir: projectsDir}, nil
}

// Dir 返回项目根目录
func (m *Manager) Dir() string {
	return m.projectsDir
}

// ProjectDir 返回单个项目的目录
func (m *Manager) ProjectDir(name string) string {
	return filepath.Join(m.projectsDir, sanitizeName(name))
}

// ComposeFile 返回项目默认的 compose 文件路径
func (m *Manager) ComposeFile(name string) string {
	return filepath.Join(m.ProjectDir(name), "docker-compose.yml")
}

// Save 保存项目 YAML(先写临时文件校验,再原子替换)
func (m *Manager) Save(name, yaml string) error {
	name = sanitizeName(name)
	if name == "" {
		return fmt.Errorf("项目名不能为空")
	}
	dir := filepath.Join(m.projectsDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建项目目录失败: %w", err)
	}
	final := filepath.Join(dir, "docker-compose.yml")
	tmp := final + ".tmp"

	if err := os.WriteFile(tmp, []byte(yaml), 0o644); err != nil {
		return fmt.Errorf("写入项目文件失败: %w", err)
	}
	if err := os.Rename(tmp, final); err != nil {
		return fmt.Errorf("保存项目文件失败: %w", err)
	}
	slog.Info("compose 项目已保存", "project", name, "file", final)
	return nil
}

// Read 读取项目 YAML 内容
func (m *Manager) Read(name string) (string, error) {
	data, err := os.ReadFile(m.ComposeFile(name))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("项目文件不存在(可能尚未创建或无法访问)")
		}
		return "", fmt.Errorf("读取项目文件失败: %w", err)
	}
	return string(data), nil
}

// ReadFile 读取任意路径的 compose 文件(用于编辑已有项目)
func (m *Manager) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取 compose 文件失败: %w", err)
	}
	return string(data), nil
}

// Remove 删除项目目录(仅本工具管理的项目)
func (m *Manager) Remove(name string) error {
	return os.RemoveAll(m.ProjectDir(name))
}

// ListManaged 列出本工具管理的项目名(projects_dir 下含 compose 文件的目录)
func (m *Manager) ListManaged() []string {
	entries, err := os.ReadDir(m.projectsDir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(m.projectsDir, e.Name(), "docker-compose.yml")); err == nil {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}

// ComposeAvailable 检测 docker compose CLI 是否可用
func ComposeAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "compose", "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		slog.Warn("docker compose CLI 不可用,compose 项目写操作将降级", "error", err, "output", strings.TrimSpace(out.String()))
		return false
	}
	return true
}

// Result 是 compose CLI 执行结果
type Result struct {
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
}

// Up 执行 docker compose up -d(files 支持多 compose 文件)
func (m *Manager) Up(name string, files []string, extraArgs ...string) (*Result, error) {
	return m.runCompose(name, files, append([]string{"up", "-d"}, extraArgs...)...)
}

// Down 执行 docker compose down(默认不删卷)
func (m *Manager) Down(name string, files []string, removeVolumes bool) (*Result, error) {
	args := []string{"down"}
	if removeVolumes {
		args = append(args, "-v")
	}
	return m.runCompose(name, files, args...)
}

// Restart 执行 docker compose restart
func (m *Manager) Restart(name string, files []string) (*Result, error) {
	return m.runCompose(name, files, "restart")
}

// Stop 执行 docker compose stop(停止容器但保留容器与网络,可用 up/start 恢复)
func (m *Manager) Stop(name string, files []string) (*Result, error) {
	return m.runCompose(name, files, "stop")
}

// Rebuild 执行 down → 等待 1 秒 → up -d,返回合并输出
func (m *Manager) Rebuild(name string, files []string) (*Result, error) {
	down, err := m.Down(name, files, false)
	if err != nil {
		return down, fmt.Errorf("重建失败(down 阶段): %s", down.Output)
	}
	time.Sleep(1 * time.Second)
	up, err := m.Up(name, files)
	if err != nil {
		return up, fmt.Errorf("重建失败(up 阶段): %s", up.Output)
	}
	output := strings.TrimSpace(down.Output)
	if output != "" {
		output += "\n"
	}
	output += strings.TrimSpace(up.Output)
	return &Result{ExitCode: 0, Output: output}, nil
}

// runCompose 执行 docker compose,compose 文件所在目录作为工作目录(相对路径正确)
func (m *Manager) runCompose(name string, files []string, args ...string) (*Result, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("未指定 compose 文件")
	}
	cmdArgs := []string{"compose"}
	for _, f := range files {
		cmdArgs = append(cmdArgs, "-f", f)
	}
	cmdArgs = append(cmdArgs, "-p", sanitizeName(name))
	cmdArgs = append(cmdArgs, args...)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)
	cmd.Dir = filepath.Dir(files[0])
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	slog.Info("执行 docker compose", "args", strings.Join(cmdArgs, " "), "dir", cmd.Dir)
	err := cmd.Run()

	res := &Result{Output: strings.TrimSpace(out.String())}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.ExitCode = ee.ExitCode()
		} else {
			res.ExitCode = -1
		}
		return res, fmt.Errorf("docker compose 执行失败: %s", res.Output)
	}
	return res, nil
}

// sanitizeName 项目名只允许安全字符(防止路径穿越)
func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		return ""
	}
	for _, c := range name {
		ok := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.'
		if !ok {
			return ""
		}
	}
	return name
}
