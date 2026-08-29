package compose

import (
	"strings"
	"testing"
)

func TestConvertDockerRunBasic(t *testing.T) {
	cmd := `docker run -d --name nginx-web -p 8080:80 -v /data/html:/usr/share/nginx/html -e TZ=Asia/Shanghai --restart unless-stopped nginx:alpine`
	yamlStr, svc, err := ConvertDockerRun(cmd)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	if svc != "nginx-web" {
		t.Errorf("服务名应为 nginx-web: %s", svc)
	}
	for _, want := range []string{
		"image: nginx:alpine",
		"container_name: nginx-web",
		"8080:80",
		"/data/html:/usr/share/nginx/html",
		"TZ=Asia/Shanghai",
		"restart: unless-stopped",
	} {
		if !strings.Contains(yamlStr, want) {
			t.Errorf("YAML 缺少 %q:\n%s", want, yamlStr)
		}
	}
}

func TestConvertDockerRunCommandAndArgs(t *testing.T) {
	cmd := `docker run -d redis:7 --appendonly yes --maxmemory 100mb`
	yamlStr, _, err := ConvertDockerRun(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(yamlStr, "image: redis:7") {
		t.Errorf("镜像缺失:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "--appendonly") || !strings.Contains(yamlStr, "--maxmemory") {
		t.Errorf("command 参数缺失:\n%s", yamlStr)
	}
}

func TestConvertDockerRunQuotes(t *testing.T) {
	cmd := `docker run -d -e "MYSQL_ROOT_PASSWORD=pass word" -e FOO='bar baz' mysql:8`
	yamlStr, _, err := ConvertDockerRun(cmd)
	if err != nil {
		t.Fatalf("带引号命令解析失败: %v", err)
	}
	if !strings.Contains(yamlStr, "MYSQL_ROOT_PASSWORD=pass word") {
		t.Errorf("双引号参数错误:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "FOO=bar baz") {
		t.Errorf("单引号参数错误:\n%s", yamlStr)
	}
}

func TestConvertDockerRunTTY(t *testing.T) {
	cmd := `docker run -it --rm ubuntu:24.04 bash`
	yamlStr, _, err := ConvertDockerRun(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(yamlStr, "tty: true") || !strings.Contains(yamlStr, "stdin_open: true") {
		t.Errorf("-it 未转换:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "bash") {
		t.Errorf("command 缺失:\n%s", yamlStr)
	}
}

func TestConvertDockerRunPrivilegedAndCaps(t *testing.T) {
	cmd := `docker run -d --privileged --cap-add NET_ADMIN --cap-drop ALL portainer/portainer-ce`
	yamlStr, _, err := ConvertDockerRun(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(yamlStr, "privileged: true") {
		t.Errorf("privileged 缺失:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "NET_ADMIN") || !strings.Contains(yamlStr, "cap_drop") {
		t.Errorf("cap 配置缺失:\n%s", yamlStr)
	}
}

func TestConvertDockerRunNoName(t *testing.T) {
	cmd := `docker run -d -p 9000:9000 nginx:latest`
	yamlStr, svc, err := ConvertDockerRun(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if svc != "nginx" {
		t.Errorf("无 --name 时应从镜像推断服务名 nginx: %s", svc)
	}
	if !strings.Contains(yamlStr, "services:") {
		t.Errorf("结构错误:\n%s", yamlStr)
	}
}

func TestConvertDockerRunInvalid(t *testing.T) {
	if _, _, err := ConvertDockerRun(""); err == nil {
		t.Error("空命令应报错")
	}
	if _, _, err := ConvertDockerRun("docker run -d"); err == nil {
		t.Error("无镜像应报错")
	}
	if _, _, err := ConvertDockerRun("ls -la"); err == nil {
		t.Error("非 docker run 命令应报错")
	}
}

func TestShellSplit(t *testing.T) {
	args, err := shellSplit(`docker run -d --name "my app" -e 'A=1 2' plain`)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"docker", "run", "-d", "--name", "my app", "-e", "A=1 2", "plain"}
	if len(args) != len(want) {
		t.Fatalf("长度不符: %v", args)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Errorf("第 %d 项: got %q want %q", i, args[i], want[i])
		}
	}
	// 未闭合引号
	if _, err := shellSplit(`docker run -e "abc`); err == nil {
		t.Error("未闭合引号应报错")
	}
}

func TestDefaultServiceName(t *testing.T) {
	cases := map[string]string{
		"nginx:latest":                    "nginx",
		"registry.example.com/app:v1":     "app",
		"ghcr.io/org/tool:1.0":            "tool",
		"busybox":                         "busybox",
		"weird/name.with-dots:tag":        "name.with-dots",
	}
	for in, want := range cases {
		if got := defaultServiceName(in); got != want {
			t.Errorf("defaultServiceName(%q) = %q, want %q", in, got, want)
		}
	}
}
