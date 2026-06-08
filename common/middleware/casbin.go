package middleware

import (
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/casbin/casbin/v2"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"strings"
)

type CasbinEnforcerMiddleware struct {
	enforcer *casbin.Enforcer
}

func NewCasbinEnforcerMiddleware(modelPath, policyPath string) (*CasbinEnforcerMiddleware, error) {
	enforcer, err := casbin.NewEnforcer(modelPath, policyPath)
	if err != nil {
		return nil, err
	}
	return &CasbinEnforcerMiddleware{enforcer: enforcer}, nil
}

// Handle Casbin 权限校验中间件
// 从 JWT 中提取 role，与请求路径和 HTTP 方法进行鉴权
func (m *CasbinEnforcerMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 从 JWT claims 中提取角色（go-zero 自动注入到 context）
		role, ok := ctx.Value("role").(string)
		if !ok || role == "" {
			httpx.WriteJsonCtx(ctx, w, http.StatusForbidden, response.Response{
				Code: response.ErrCodePermissionDenied,
				Msg:  "权限不足",
			})
			return
		}

		// 获取请求路径和方法
		path := r.URL.Path
		method := r.Method

		// 通过 Casbin 鉴权
		allowed, err := m.enforcer.Enforce(role, path, method)
		if err != nil {
			httpx.WriteJsonCtx(ctx, w, http.StatusInternalServerError, response.Response{
				Code: response.InternalError,
				Msg:  "权限校验失败: " + err.Error(),
			})
			return
		}

		if !allowed {
			httpx.WriteJsonCtx(ctx, w, http.StatusForbidden, response.Response{
				Code: response.ErrCodePermissionDenied,
				Msg:  "权限不足",
			})
			return
		}

		next(w, r)
	}
}

// GetCurrentRole 从 context 中提取当前用户的角色
func GetCurrentRole(ctx interface{ Value(key any) any }) string {
	role, ok := ctx.Value("role").(string)
	if !ok || role == "" {
		return ""
	}
	return role
}

// HasPermission 判断当前请求是否有权限（供 handler/logic 层内部使用）
func (m *CasbinEnforcerMiddleware) HasPermission(role, path, method string) bool {
	// 标准化 path：去掉 query 参数
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}
	allowed, err := m.enforcer.Enforce(role, path, method)
	if err != nil {
		return false
	}
	return allowed
}
