package database

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCreateAndGetUser(t *testing.T) {
	db := newTestDB(t)

	if err := db.CreateUser("admin", "fakehash"); err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	u, err := db.GetUserByUsername("admin")
	if err != nil {
		t.Fatalf("查询用户失败: %v", err)
	}
	if u.Username != "admin" || u.PasswordHash != "fakehash" {
		t.Errorf("用户数据不符: %+v", u)
	}

	// 重复创建应报错(UNIQUE)
	if err := db.CreateUser("admin", "another"); err == nil {
		t.Error("重复用户名应报错")
	}

	// 不存在的用户
	if _, err := db.GetUserByUsername("nobody"); !errors.Is(err, ErrUserNotFound) {
		t.Errorf("应返回 ErrUserNotFound,实际: %v", err)
	}
}

func TestUpdatePasswordHash(t *testing.T) {
	db := newTestDB(t)
	_ = db.CreateUser("admin", "oldhash")
	u, _ := db.GetUserByUsername("admin")

	if err := db.UpdatePasswordHash(u.ID, "newhash"); err != nil {
		t.Fatalf("更新密码失败: %v", err)
	}
	u2, _ := db.GetUserByUsername("admin")
	if u2.PasswordHash != "newhash" {
		t.Error("密码哈希未更新")
	}

	// 不存在的用户
	if err := db.UpdatePasswordHash(9999, "x"); !errors.Is(err, ErrUserNotFound) {
		t.Errorf("应返回 ErrUserNotFound,实际: %v", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	db := newTestDB(t)
	_ = db.CreateUser("admin", "hash")
	u, _ := db.GetUserByUsername("admin")

	token, err := db.CreateSession(u.ID, 24)
	if err != nil {
		t.Fatalf("创建 session 失败: %v", err)
	}
	if len(token) != 64 {
		t.Errorf("token 长度异常: %d", len(token))
	}

	// 通过 token 取回用户
	user, err := db.GetUserByToken(token)
	if err != nil {
		t.Fatalf("GetUserByToken 失败: %v", err)
	}
	if user.Username != "admin" {
		t.Errorf("session 用户不符: %s", user.Username)
	}

	// 删除后失效
	if err := db.DeleteSession(token); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetUserByToken(token); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("删除后应返回 ErrSessionNotFound,实际: %v", err)
	}
}

func TestSessionExpiry(t *testing.T) {
	db := newTestDB(t)
	_ = db.CreateUser("admin", "hash")
	u, _ := db.GetUserByUsername("admin")

	// TTL 0 小时 → 立即过期
	token, err := db.CreateSession(u.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetUserByToken(token); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("过期 session 应返回 ErrSessionNotFound,实际: %v", err)
	}
}

func TestDeleteExpiredSessions(t *testing.T) {
	db := newTestDB(t)
	_ = db.CreateUser("admin", "hash")
	u, _ := db.GetUserByUsername("admin")

	_, _ = db.CreateSession(u.ID, 24)  // 有效

	// 手动插入一条过期记录(1 小时前)确保清理逻辑覆盖
	expired := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	_, _ = db.sql.Exec("INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)", "expired-token", u.ID, expired)

	n, err := db.DeleteExpiredSessions()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("应清理 1 条过期 session,实际 %d", n)
	}

	// 有效 session 仍在
	var count int
	_ = db.sql.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count)
	if count != 1 {
		t.Errorf("应剩 1 条有效 session,实际 %d", count)
	}
}
