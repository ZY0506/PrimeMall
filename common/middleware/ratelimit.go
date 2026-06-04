package middleware

import (
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"time"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type TokenBucketMiddleware struct {
	mu       sync.RWMutex
	limiters map[string]*ipLimiter
	rate     rate.Limit
	burst    int
	stopCh   chan struct{}
}

// NewTokenBucketMiddleware 创建基于令牌桶算法的 IP 限流中间件
// speed: 每秒允许的请求数（rate）
// burst: 允许的突发请求数
func NewTokenBucketMiddleware(speed, burst int) *TokenBucketMiddleware {
	m := &TokenBucketMiddleware{
		limiters: make(map[string]*ipLimiter),
		rate:     rate.Limit(speed),
		burst:    burst,
		stopCh:   make(chan struct{}),
	}

	// 启动后台清理协程：每 5 分钟清理超过 1 分钟未活动的 IP
	go m.cleanupLoop(5 * time.Minute)

	return m
}

// Stop 停止后台清理协程（可在 server.Stop 时调用）
func (m *TokenBucketMiddleware) Stop() {
	close(m.stopCh)
}

// Handle 限流中间件处理函数
func (m *TokenBucketMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 获取客户端 IP
		clientIP := getClientIP(r)

		limiter := m.getLimiter(clientIP)

		if !limiter.Allow() {
			logx.WithContext(ctx).Infof("请求被限流, ip=%s, path=%s", clientIP, r.URL.Path)
			httpx.WriteJsonCtx(ctx, w, http.StatusTooManyRequests, response.Response{
				Code: response.ErrCodeTooFrequent,
				Msg:  "访问过于频繁，请稍后再试",
			})
			return
		}

		next(w, r)
	}
}

// getLimiter 获取或创建指定 IP 的令牌桶
func (m *TokenBucketMiddleware) getLimiter(ip string) *rate.Limiter {
	m.mu.RLock()
	entry, exists := m.limiters[ip]
	m.mu.RUnlock()

	if exists {
		entry.lastSeen = time.Now()
		return entry.limiter
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	if entry, exists = m.limiters[ip]; exists {
		entry.lastSeen = time.Now()
		return entry.limiter
	}

	limiter := rate.NewLimiter(m.rate, m.burst)
	m.limiters[ip] = &ipLimiter{
		limiter:  limiter,
		lastSeen: time.Now(),
	}
	return limiter
}

// cleanupLoop 定期清理长时间未活动的 IP 限流记录
func (m *TokenBucketMiddleware) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopCh:
			return
		}
	}
}

// cleanup 清理超过 1 分钟未活动的 IP 记录
func (m *TokenBucketMiddleware) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	threshold := time.Now().Add(-1 * time.Minute)
	for ip, entry := range m.limiters {
		if entry.lastSeen.Before(threshold) {
			delete(m.limiters, ip)
		}
	}
}
