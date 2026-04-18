// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package middleware

import (
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

type TokenBlacklistMiddleware struct {
	rdb *redis.Client
}

func NewTokenBlacklistMiddleware(rdb *redis.Client) *TokenBlacklistMiddleware {
	return &TokenBlacklistMiddleware{
		rdb: rdb,
	}
}

func (m *TokenBlacklistMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 从 ctx 里拿 jti
		jti, ok := ctx.Value("jti").(string)
		if !ok || jti == "" {
			httpx.WriteJsonCtx(ctx, w, http.StatusUnauthorized, response.Response{
				Code: response.ErrCodeTokenInvalid,
				Msg:  "Token 无效",
			})
			return
		}

		key := constants.ATOKEN_BLACKLIST_KEY + jti

		// 查 Redis
		exists, err := m.rdb.Exists(ctx, key).Result()
		if err != nil {
			httpx.WriteJsonCtx(ctx, w, http.StatusInternalServerError, response.Response{
				Code: response.InternalError,
				Msg:  "Internal Error",
				Data: nil,
			})
			return
		}

		// 在黑名单
		if exists == 1 {
			httpx.WriteJsonCtx(ctx, w, http.StatusUnauthorized, response.Response{
				Code: response.ErrCodeUnauthorized,
				Msg:  "请重新登录",
			})
			return
		}
		next(w, r)
	}
}
