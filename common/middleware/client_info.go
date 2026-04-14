package middleware

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
	"strings"
)

const ContextKeyClientInfo = "client_info"

type ClientInfo struct {
	IP        string
	UserAgent string
	DeviceID  string
}

type Middleware struct{}

func NewMiddleware() *Middleware {
	return &Middleware{}
}

// ClientInfoHandle 获取客户端信息
func (m *Middleware) ClientInfoHandle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		userAgent := r.Header.Get("User-Agent")

		deviceId := r.Header.Get("X-Device-Id")
		if deviceId == "" {
			deviceId = generateDeviceFingerprint(clientIP, userAgent)
		}

		// 注入 Context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyClientInfo, &ClientInfo{
			IP:        clientIP,
			UserAgent: userAgent,
			DeviceID:  deviceId,
		})

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

// GetClientInfo 获取 ClientInfo
func GetClientInfo(ctx context.Context) *ClientInfo {
	val, ok := ctx.Value(ContextKeyClientInfo).(*ClientInfo)
	if !ok {
		logx.WithContext(ctx).Error("上下文信息缺失")
		panic("missing client info")
	}
	return val
}
