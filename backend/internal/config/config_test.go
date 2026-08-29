package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.Server.Port != 8080 {
		t.Errorf("默认端口应为 8080,实际 %d", cfg.Server.Port)
	}
	if cfg.Docker.Host != "unix:///var/run/docker.sock" {
		t.Errorf("默认 docker host 错误: %s", cfg.Docker.Host)
	}
	if cfg.Auth.PasswordEnv != "DM_ADMIN_PASSWORD" {
		t.Errorf("默认密码环境变量名错误: %s", cfg.Auth.PasswordEnv)
	}
}

func TestLoadFromYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
server:
  host: 127.0.0.1
  port: 9999
docker:
  host: tcp://10.0.0.1:2375
database:
  path: /tmp/test.db
auth:
  username: root
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load 失败: %v", err)
	}
	if cfg.Server.Host != "127.0.0.1" || cfg.Server.Port != 9999 {
		t.Errorf("yaml server 配置未生效: %+v", cfg.Server)
	}
	if cfg.Docker.Host != "tcp://10.0.0.1:2375" {
		t.Errorf("yaml docker 配置未生效: %s", cfg.Docker.Host)
	}
	if cfg.Auth.Username != "root" {
		t.Errorf("yaml auth 配置未生效: %s", cfg.Auth.Username)
	}
}

func TestEnvOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	_ = os.WriteFile(path, []byte("server:\n  port: 1111\n"), 0o644)

	t.Setenv("SERVER_PORT", "2222")
	t.Setenv("DOCKER_HOST", "tcp://1.2.3.4:2376")
	t.Setenv("AUTH_USERNAME", "envuser")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != 2222 {
		t.Errorf("环境变量 SERVER_PORT 应覆盖 yaml: %d", cfg.Server.Port)
	}
	if cfg.Docker.Host != "tcp://1.2.3.4:2376" {
		t.Errorf("环境变量 DOCKER_HOST 未生效: %s", cfg.Docker.Host)
	}
	if cfg.Auth.Username != "envuser" {
		t.Errorf("环境变量 AUTH_USERNAME 未生效: %s", cfg.Auth.Username)
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("配置文件不存在应使用默认值,而不是报错: %v", err)
	}
	if cfg == nil {
		t.Fatal("cfg 不应为 nil")
	}
}

func TestPasswordFromEnv(t *testing.T) {
	t.Setenv("DM_ADMIN_PASSWORD", "secret123")
	cfg := Default()
	if cfg.Auth.Password() != "secret123" {
		t.Error("Password() 应读取环境变量")
	}
}
