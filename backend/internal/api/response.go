package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Response 是统一 API 响应格式
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError 是错误结构
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OK 返回成功响应
func OK(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, Response{Success: true, Data: data})
}

// Created 返回创建成功响应
func Created(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusCreated, Response{Success: true, Data: data})
}

// NoContent 返回 204
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Fail 返回业务错误(HTTP 200 + success:false,前端统一处理)
func Fail(w http.ResponseWriter, code string, message string) {
	writeJSON(w, http.StatusOK, Response{
		Success: false,
		Error:   &APIError{Code: code, Message: message},
	})
}

// Error 返回 HTTP 错误(用于鉴权失败等)
func Error(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, Response{
		Success: false,
		Error:   &APIError{Code: code, Message: message},
	})
}

// Unauthorized 401
func Unauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "未登录或登录已过期"
	}
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden 403
func Forbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "没有权限执行该操作"
	}
	Error(w, http.StatusForbidden, "FORBIDDEN", message)
}

// writeJSON 写出 JSON 响应
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("写 JSON 响应失败", "error", err)
	}
}
