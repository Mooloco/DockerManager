package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mooloco/docker-manager/backend/internal/api"
	"github.com/Mooloco/docker-manager/backend/internal/api/middleware"
	"github.com/Mooloco/docker-manager/backend/internal/auth"
)

// AuthHandler 处理登录/登出/密码管理
type AuthHandler struct {
	auth *auth.Service
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authSvc *auth.Service) *AuthHandler {
	return &AuthHandler{auth: authSvc}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Fail(w, "INVALID_REQUEST", "请求格式错误")
		return
	}
	if req.Username == "" || req.Password == "" {
		api.Fail(w, "INVALID_REQUEST", "用户名和密码不能为空")
		return
	}

	token, err := h.auth.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			api.Fail(w, "INVALID_CREDENTIALS", "用户名或密码错误")
			return
		}
		api.Fail(w, "LOGIN_FAILED", "登录失败,请稍后重试")
		return
	}

	// 设置 HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// 未设置 MaxAge:session 有效期由数据库控制
	})
	api.OK(w, map[string]string{"username": req.Username})
}

// Logout POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(middleware.SessionCookie)
	if err == nil && cookie.Value != "" {
		_ = h.auth.Logout(cookie.Value)
	}
	// 清除 cookie
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	api.OK(w, map[string]bool{"logged_out": true})
}

// Me GET /api/v1/auth/me 返回当前登录用户
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	u := middleware.UserFromContext(r.Context())
	if u == nil {
		api.Unauthorized(w, "")
		return
	}
	api.OK(w, map[string]string{"username": u.Username})
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ChangePassword POST /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	u := middleware.UserFromContext(r.Context())
	if u == nil {
		api.Unauthorized(w, "")
		return
	}
	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Fail(w, "INVALID_REQUEST", "请求格式错误")
		return
	}
	if len(req.NewPassword) < 8 {
		api.Fail(w, "WEAK_PASSWORD", "新密码长度至少 8 位")
		return
	}

	err := h.auth.ChangePassword(u.ID, req.OldPassword, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrWrongPassword):
			api.Fail(w, "WRONG_PASSWORD", "当前密码不正确")
		case errors.Is(err, auth.ErrInvalidCredentials):
			api.Fail(w, "WRONG_PASSWORD", "当前密码不正确")
		default:
			api.Fail(w, "CHANGE_PASSWORD_FAILED", "修改密码失败")
		}
		return
	}
	api.OK(w, map[string]bool{"changed": true})
}
