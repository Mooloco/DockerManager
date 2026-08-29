package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// User 表示一个登录用户
type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

var ErrUserNotFound = errors.New("用户不存在")

// CreateUser 创建用户(密码需已哈希)
func (d *DB) CreateUser(username, passwordHash string) error {
	_, err := d.sql.Exec(
		"INSERT INTO users (username, password_hash) VALUES (?, ?)",
		username, passwordHash,
	)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}
	return nil
}

// UserCount 返回用户总数
func (d *DB) UserCount() (int, error) {
	var n int
	err := d.sql.QueryRow("SELECT COUNT(*) FROM users").Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("统计用户失败: %w", err)
	}
	return n, nil
}

// GetUserByUsername 按用户名查询用户
func (d *DB) GetUserByUsername(username string) (*User, error) {
	u := &User{}
	err := d.sql.QueryRow(
		"SELECT id, username, password_hash, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return u, nil
}

// UpdatePasswordHash 更新用户密码哈希
func (d *DB) UpdatePasswordHash(userID int64, newHash string) error {
	res, err := d.sql.Exec(
		"UPDATE users SET password_hash = ?, updated_at = datetime('now') WHERE id = ?",
		newHash, userID,
	)
	if err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrUserNotFound
	}
	return nil
}
