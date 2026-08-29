package database

import (
	"crypto/rand"
	"fmt"
)

// newSessionToken 生成 32 字节随机 hex token
func newSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("生成 session token 失败: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}
