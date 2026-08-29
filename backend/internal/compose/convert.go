package compose

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConvertDockerRun 把 docker run 命令转换为 compose YAML。
// 返回生成的 YAML 与建议的服务名。
func ConvertDockerRun(command string) (string, string, error) {
	args, err := shellSplit(command)
	if err != nil {
		return "", "", fmt.Errorf("命令解析失败: %w", err)
	}
	if len(args) == 0 {
		return "", "", fmt.Errorf("命令不能为空")
	}

	// 去掉 docker 前缀:docker / docker container / container
	i := 0
	for i < len(args) && (args[i] == "docker" || args[i] == "container" || args[i] == "podman") {
		i++
	}
	// 必须是 run/create 子命令
	if i >= len(args) || (args[i] != "run" && args[i] != "create") {
		return "", "", fmt.Errorf("请提供 docker run 命令,示例: docker run -d --name nginx -p 8080:80 nginx:latest")
	}
	i++

	var name, image, restart, network string
	var ports, volumes, envs, caps, drops, cmd []string
	privileged := false
	tty := false
	stdin := false

	// 剩余参数
	rest := args[i:]
	for j := 0; j < len(rest); j++ {
		tok := rest[j]
		// 取 flag 值:支持 --name=value 或 --name value
		next := func() (string, bool) {
			if j+1 < len(rest) {
				j++
				return rest[j], true
			}
			return "", false
		}
		key, val, hasEq := strings.Cut(tok, "=")

		switch key {
		case "-d", "--detach", "--rm":
			// compose 默认后台,忽略
		case "-it", "-ti":
			tty = true
			stdin = true
		case "-i", "--interactive":
			stdin = true
		case "-t", "--tty":
			tty = true
		case "--name":
			v, ok := valIfMissing(val, hasEq, next)
			if !ok {
				return "", "", fmt.Errorf("--name 缺少参数")
			}
			name = v
		case "-p", "--publish":
			v, ok := valIfMissing(val, hasEq, next)
			if !ok {
				return "", "", fmt.Errorf("%s 缺少参数", key)
			}
			ports = append(ports, v)
		case "-v", "--volume":
			v, ok := valIfMissing(val, hasEq, next)
			if !ok {
				return "", "", fmt.Errorf("%s 缺少参数", key)
			}
			volumes = append(volumes, v)
		case "-e", "--env":
			v, ok := valIfMissing(val, hasEq, next)
			if !ok {
				return "", "", fmt.Errorf("%s 缺少参数", key)
			}
			envs = append(envs, v)
		case "--restart":
			v, ok := valIfMissing(val, hasEq, next)
			if !ok {
				return "", "", fmt.Errorf("--restart 缺少参数")
			}
			restart = v
		case "--network", "--net":
			v, ok := valIfMissing(val, hasEq, next)
			if !ok {
				return "", "", fmt.Errorf("%s 缺少参数", key)
			}
			network = v
		case "--cap-add":
			v, ok := valIfMissing(val, hasEq, next)
			if !ok {
				return "", "", fmt.Errorf("--cap-add 缺少参数")
			}
			caps = append(caps, v)
		case "--cap-drop":
			v, ok := valIfMissing(val, hasEq, next)
			if !ok {
				return "", "", fmt.Errorf("--cap-drop 缺少参数")
			}
			drops = append(drops, v)
		case "--privileged":
			privileged = true
		case "--hostname", "-h":
			// 可选转换,跳过 hostname
			if _, ok := valIfMissing(val, hasEq, next); !ok {
				return "", "", fmt.Errorf("%s 缺少参数", key)
			}
		case "-w", "--workdir", "--entrypoint", "--user", "-u", "--env-file", "--log-driver", "--log-opt", "--label", "-l", "--label-file", "--add-host", "--dns", "--ip", "--ip6", "--link", "--expose", "--health-cmd", "--health-interval", "--health-timeout", "--health-retries", "--stop-signal", "--stop-timeout", "--init", "--shm-size", "--memory", "-m", "--cpus", "--cpuset-cpus", "--pids-limit", "--ulimit", "--sysctl", "--tmpfs", "--device", "--security-opt", "--read-only", "--platform", "--pull", "--sig-proxy", "--pid", "--ipc", "--uts", "--userns", "--cgroupns", "--volume-driver", "--mount", "--gpus", "--runtime":
			// 暂不转换的参数:跳过其值
			if _, ok := valIfMissing(val, hasEq, next); !ok {
				return "", "", fmt.Errorf("%s 缺少参数", key)
			}
		default:
			// 镜像确定之后的参数(含 -xxx)都是容器命令参数
			if image != "" {
				cmd = append(cmd, tok)
				continue
			}
			if strings.HasPrefix(key, "-") {
				// 镜像前的未知 flag:尝试跳过其值(尽力转换)
				if _, ok := valIfMissing(val, hasEq, next); !ok {
					// 没有值,忽略
				}
				continue
			}
			// 第一个非 flag 参数是镜像
			image = tok
		}
	}

	if image == "" {
		return "", "", fmt.Errorf("未找到镜像名")
	}

	// 组装 compose 结构
	svc := map[string]interface{}{}
	svc["image"] = image
	if name != "" {
		svc["container_name"] = name
	}
	if len(ports) > 0 {
		svc["ports"] = ports
	}
	if len(volumes) > 0 {
		svc["volumes"] = volumes
	}
	if len(envs) > 0 {
		svc["environment"] = envs
	}
	if restart != "" {
		svc["restart"] = restart
	}
	if network != "" {
		svc["networks"] = []string{network}
	}
	if len(caps) > 0 {
		svc["cap_add"] = caps
	}
	if len(drops) > 0 {
		svc["cap_drop"] = drops
	}
	if privileged {
		svc["privileged"] = true
	}
	if tty || stdin {
		svc["tty"] = tty
		svc["stdin_open"] = stdin
	}
	if len(cmd) > 0 {
		svc["command"] = cmd
	}

	serviceName := name
	if serviceName == "" {
		serviceName = defaultServiceName(image)
	}
	compose := map[string]interface{}{
		"services": map[string]interface{}{
			serviceName: svc,
		},
	}
	data, err := yaml.Marshal(compose)
	if err != nil {
		return "", "", fmt.Errorf("生成 YAML 失败: %w", err)
	}
	return string(data), serviceName, nil
}

// valIfMissing 处理 --flag=value 与 --flag value 两种形式
func valIfMissing(val string, hasEq bool, next func() (string, bool)) (string, bool) {
	if hasEq {
		if val == "" {
			return "", false
		}
		return val, true
	}
	return next()
}

// defaultServiceName 从镜像名推断服务名:去掉 registry 与 tag
func defaultServiceName(image string) string {
	n := image
	// 去掉 tag
	if idx := strings.LastIndex(n, ":"); idx > strings.LastIndex(n, "/") {
		n = n[:idx]
	}
	// 去掉 registry 前缀
	if idx := strings.LastIndex(n, "/"); idx >= 0 {
		n = n[idx+1:]
	}
	// 非法字符替换
	var sb strings.Builder
	for _, c := range n {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' {
			sb.WriteRune(c)
		} else {
			sb.WriteRune('-')
		}
	}
	if sb.Len() == 0 {
		return "app"
	}
	return sb.String()
}

// shellSplit 简单的 shell 分词:支持单引号、双引号、反斜杠转义
func shellSplit(s string) ([]string, error) {
	var args []string
	var cur strings.Builder
	inSingle, inDouble := false, false
	escaped := false
	started := false

	flush := func() {
		if started {
			args = append(args, cur.String())
			cur.Reset()
			started = false
		}
	}

	for _, r := range s {
		if escaped {
			cur.WriteRune(r)
			started = true
			escaped = false
			continue
		}
		switch {
		case r == '\\' && !inSingle:
			escaped = true
			started = true
		case r == '\'' && !inDouble:
			inSingle = !inSingle
			started = true
		case r == '"' && !inSingle:
			inDouble = !inDouble
			started = true
		case (r == ' ' || r == '\t' || r == '\n') && !inSingle && !inDouble:
			flush()
		default:
			cur.WriteRune(r)
			started = true
		}
	}
	if escaped {
		return nil, fmt.Errorf("结尾存在未完成的转义")
	}
	if inSingle || inDouble {
		return nil, fmt.Errorf("引号未闭合")
	}
	flush()
	return args, nil
}
