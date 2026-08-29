package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 是应用的顶层配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Docker   DockerConfig   `yaml:"docker"`
	Database DatabaseConfig `yaml:"database"`
	Logging  LoggingConfig  `yaml:"logging"`
	Auth     AuthConfig     `yaml:"auth"`
	Compose  ComposeConfig  `yaml:"compose"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DockerConfig struct {
	Host string `yaml:"host"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type AuthConfig struct {
	// Username 是初始管理员用户名(仅首次启动创建账号时使用)
	Username string `yaml:"username"`
	// PasswordEnv 指定存放初始管理员密码的环境变量名
	PasswordEnv string `yaml:"password_env"`
	// SessionTTLHours 是登录 session 的有效期(小时)
	SessionTTLHours int `yaml:"session_ttl_hours"`
}

type ComposeConfig struct {
	// ProjectsDir 是 compose 项目 YAML 的存储目录
	ProjectsDir string `yaml:"projects_dir"`
}

// Default 返回默认配置
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Docker: DockerConfig{
			Host: "unix:///var/run/docker.sock",
		},
		Database: DatabaseConfig{
			Path: "./data/docker-manager.db",
		},
		Logging: LoggingConfig{
			Level: "info",
		},
		Auth: AuthConfig{
			Username:        "admin",
			PasswordEnv:     "DM_ADMIN_PASSWORD",
			SessionTTLHours: 24,
		},
		Compose: ComposeConfig{
			ProjectsDir: "./data/projects",
		},
	}
}

// Load 加载配置:默认值 -> YAML 文件 -> 环境变量覆盖
func Load(path string) (*Config, error) {
	cfg := Default()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("读取配置文件失败: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("解析配置文件 %s 失败: %w", path, err)
			}
		}
	}

	// 环境变量覆盖
	applyEnv(cfg)
	return cfg, nil
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = n
		}
	}
	if v := os.Getenv("DOCKER_HOST"); v != "" {
		cfg.Docker.Host = v
	}
	if v := os.Getenv("DATABASE_PATH"); v != "" {
		cfg.Database.Path = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("AUTH_USERNAME"); v != "" {
		cfg.Auth.Username = v
	}
	if v := os.Getenv("AUTH_PASSWORD_ENV"); v != "" {
		cfg.Auth.PasswordEnv = v
	}
	if v := os.Getenv("AUTH_SESSION_TTL_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Auth.SessionTTLHours = n
		}
	}
	if v := os.Getenv("DM_PROJECTS_DIR"); v != "" {
		cfg.Compose.ProjectsDir = v
	}
}

// Password 返回初始管理员密码(从环境变量读取)
func (c *AuthConfig) Password() string {
	return os.Getenv(c.PasswordEnv)
}

// Level 返回规范化的小写日志级别
func (c *LoggingConfig) LevelValue() string {
	return strings.ToLower(c.Level)
}
