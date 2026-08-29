package auth

import (
	"strings"
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("S3cret!pass")
	if err != nil {
		t.Fatalf("HashPassword 失败: %v", err)
	}
	// 格式: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("哈希格式异常: %s", hash)
	}
	if strings.Contains(hash, "S3cret!pass") {
		t.Error("哈希中不应包含明文密码")
	}

	// 正确密码验证通过
	ok, err := VerifyPassword("S3cret!pass", hash)
	if err != nil || !ok {
		t.Errorf("正确密码应验证通过, ok=%v err=%v", ok, err)
	}

	// 错误密码验证失败
	ok, _ = VerifyPassword("wrong-pass", hash)
	if ok {
		t.Error("错误密码不应验证通过")
	}
}

func TestHashIsUnique(t *testing.T) {
	h1, _ := HashPassword("same-password")
	h2, _ := HashPassword("same-password")
	if h1 == h2 {
		t.Error("相同密码两次哈希应不同(随机盐)")
	}
}

func TestVerifyMalformedHash(t *testing.T) {
	ok, err := VerifyPassword("x", "not-a-valid-hash")
	if err == nil {
		t.Error("畸形哈希应返回错误")
	}
	if ok {
		t.Error("畸形哈希不应验证通过")
	}
}

func TestNewToken(t *testing.T) {
	t1, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken 失败: %v", err)
	}
	t2, _ := NewToken()
	if len(t1) != 64 { // 32 字节 hex = 64 字符
		t.Errorf("token 长度应为 64,实际 %d", len(t1))
	}
	if t1 == t2 {
		t.Error("两次生成的 token 不应相同")
	}
}
