package middleware

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
)

// 定义私有类型，防止外部冲突
type ctxKey string

const (
	ContextKeyIP        ctxKey = "client_ip"
	ContextKeyUserAgent ctxKey = "user_agent"
	ContextKeyDeviceId  ctxKey = "device_id"
)

type Middleware struct{}

func NewMiddleware() *Middleware {
	return &Middleware{}
}

func (m *Middleware) ClientInfoHandle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		userAgent := r.Header.Get("User-Agent")

		deviceId := r.Header.Get("X-Device-Id")
		if deviceId == "" {
			deviceId = generateDeviceFingerprint(clientIP, userAgent)
		}

		// 批量注入 Context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyIP, clientIP)
		ctx = context.WithValue(ctx, ContextKeyUserAgent, userAgent)
		ctx = context.WithValue(ctx, ContextKeyDeviceId, deviceId)

		next(w, r.WithContext(ctx))
	}
}

// --- 辅助工具函数 ---
func getClientIP(r *http.Request) string {
	// GoZero 部署通常会有 Nginx 转发
	// X-Forwarded-For 比较可靠，但要注意第一个 IP 才是原始客户端
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

func generateDeviceFingerprint(ip, userAgent string) string {
	raw := fmt.Sprintf("%s|%s", ip, userAgent)
	return fmt.Sprintf("%x", md5.Sum([]byte(raw)))
}
