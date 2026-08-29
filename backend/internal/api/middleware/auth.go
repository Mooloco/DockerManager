package middleware

import (
	"context"
	"net/http"

	"github.com/Mooloco/docker-manager/backend/internal/api"
	"github.com/Mooloco/docker-manager/backend/internal/auth"
)

// SessionCookie 是 session cookie 名称
const SessionCookie = "dm_session"

// userKey 是 context 中用户信息的键
type userKey struct{}

// UserContext 是放入 context 的用户信息
type UserContext struct {
	ID       int64
	Username string
}

// RequireAuth 校验 session cookie,未认证返回 401
func RequireAuth(authSvc *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookie)
			if err != nil || cookie.Value == "" {
				api.Unauthorized(w, "")
				return
			}
			user, err := authSvc.UserByToken(cookie.Value)
			if err != nil {
				// 清除无效 cookie
				http.SetCookie(w, &http.Cookie{
					Name:     SessionCookie,
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
				api.Unauthorized(w, "登录已过期,请重新登录")
				return
			}
			ctx := context.WithValue(r.Context(), userKey{}, &UserContext{
				ID:       user.ID,
				Username: user.Username,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext 从 context 取当前用户(仅限已认证路由)
func UserFromContext(ctx context.Context) *UserContext {
	if u, ok := ctx.Value(userKey{}).(*UserContext); ok {
		return u
	}
	return nil
}
