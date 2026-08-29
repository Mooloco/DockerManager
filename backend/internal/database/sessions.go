package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrSessionNotFound = errors.New("session 不存在或已过期")

// CreateSession 创建一条 session,返回随机 token
func (d *DB) CreateSession(userID int64, ttlHours int) (string, error) {
	token, err := newSessionToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(time.Duration(ttlHours) * time.Hour).UTC().Format(time.RFC3339)
	_, err = d.sql.Exec(
		"INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)",
		token, userID, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("创建 session 失败: %w", err)
	}
	return token, nil
}

// GetUserByToken 校验 token 并返回对应用户;过期或不存在返回 ErrSessionNotFound
func (d *DB) GetUserByToken(token string) (*User, error) {
	u := &User{}
	var expiresAt string
	err := d.sql.QueryRow(
		`SELECT u.id, u.username, u.password_hash, u.created_at, u.updated_at, s.expires_at
		 FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.token = ?`,
		token,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询 session 失败: %w", err)
	}

	exp, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil || time.Now().After(exp) {
		// 过期 session 顺手清理
		_, _ = d.sql.Exec("DELETE FROM sessions WHERE token = ?", token)
		return nil, ErrSessionNotFound
	}

	// 更新最后活跃时间(忽略失败)
	_, _ = d.sql.Exec("UPDATE sessions SET last_seen_at = datetime('now') WHERE token = ?", token)
	return u, nil
}

// DeleteSession 删除一条 session(登出)
func (d *DB) DeleteSession(token string) error {
	_, err := d.sql.Exec("DELETE FROM sessions WHERE token = ?", token)
	if err != nil {
		return fmt.Errorf("删除 session 失败: %w", err)
	}
	return nil
}

// DeleteExpiredSessions 清理所有过期 session
func (d *DB) DeleteExpiredSessions() (int64, error) {
	res, err := d.sql.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return 0, fmt.Errorf("清理过期 session 失败: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}
