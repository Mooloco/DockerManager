package auth

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/Mooloco/docker-manager/backend/internal/config"
	"github.com/Mooloco/docker-manager/backend/internal/database"
)

var (
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrWrongPassword      = errors.New("当前密码错误")
)

// Service 负责登录、session 生命周期与密码管理
type Service struct {
	db  *database.DB
	cfg *config.AuthConfig
}

// NewService 创建认证服务
func NewService(db *database.DB, cfg *config.AuthConfig) *Service {
	return &Service{db: db, cfg: cfg}
}

// EnsureAdmin 首次启动时创建初始管理员账号。
// 已有用户则跳过;密码来自配置指定的环境变量。
func (s *Service) EnsureAdmin() error {
	n, err := s.db.UserCount()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	password := s.cfg.Password()
	if password == "" {
		return fmt.Errorf("未设置初始管理员密码:请通过环境变量 %s 指定", s.cfg.PasswordEnv)
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	if err := s.db.CreateUser(s.cfg.Username, hash); err != nil {
		return err
	}
	slog.Info("已创建初始管理员账号", "username", s.cfg.Username)
	return nil
}

// Login 校验用户名密码,成功则创建 session 并返回 token
func (s *Service) Login(username, password string) (string, error) {
	u, err := s.db.GetUserByUsername(username)
	if errors.Is(err, database.ErrUserNotFound) {
		return "", ErrInvalidCredentials
	}
	if err != nil {
		return "", err
	}
	ok, err := VerifyPassword(password, u.PasswordHash)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrInvalidCredentials
	}
	token, err := s.db.CreateSession(u.ID, s.cfg.SessionTTLHours)
	if err != nil {
		return "", err
	}
	slog.Info("用户登录成功", "username", username)
	return token, nil
}

// Logout 登出,删除 session
func (s *Service) Logout(token string) error {
	return s.db.DeleteSession(token)
}

// UserByToken 通过 token 获取用户(用于中间件鉴权)
func (s *Service) UserByToken(token string) (*database.User, error) {
	return s.db.GetUserByToken(token)
}

// ChangePassword 修改密码:先校验旧密码,再写入新密码
func (s *Service) ChangePassword(userID int64, oldPassword, newPassword string) error {
	u, err := s.userByID(userID)
	if err != nil {
		return err
	}
	ok, err := VerifyPassword(oldPassword, u.PasswordHash)
	if err != nil {
		return err
	}
	if !ok {
		return ErrWrongPassword
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.db.UpdatePasswordHash(u.ID, hash); err != nil {
		return err
	}
	slog.Info("用户修改了密码", "username", u.Username)
	return nil
}

// userByID 按 ID 查询用户(通过 username 反查,保持存储层 API 最小)
func (s *Service) userByID(userID int64) (*database.User, error) {
	// database 层未提供按 ID 查询,这里用会话表反查;
	// 简化处理:遍历是不合理的,直接扩展数据库层查询
	return s.db.GetUserByID(userID)
}
