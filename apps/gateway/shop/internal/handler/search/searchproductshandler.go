package search

import (
	"net/http"
	"strings"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/search"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	commonjwt "github.com/ZY0506/PrimeMall/common/jwt"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func SearchProductsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SearchProductsReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ClientError(r.Context(), w, response.ErrCodeInvalidParam, err.Error())
			return
		}

		// 参数校验
		if req.Page < 1 {
			response.ClientError(r.Context(), w, response.ErrCodeInvalidPage, "页码必须大于0")
			return
		}
		if req.Size < 1 || req.Size > 100 {
			response.ClientError(r.Context(), w, response.ErrCodeParamRange, "每页数量必须在1-100之间")
			return
		}
		if req.SortBy != "" && req.SortBy != "price" && req.SortBy != "sales" {
			response.ClientError(r.Context(), w, response.ErrCodeInvalidParam, "排序字段仅支持 price 或 sales")
			return
		}
		if req.SortType != "" && req.SortType != "asc" && req.SortType != "desc" {
			response.ClientError(r.Context(), w, response.ErrCodeInvalidParam, "排序方向仅支持 asc 或 desc")
			return
		}

		// 可选鉴权：如果有token则提取userId用于记录搜索历史
		ctx := r.Context()
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr := authHeader[7:]
			claims, err := commonjwt.ParseToken(tokenStr, []byte(svcCtx.Config.JwtAuth.AccessSecret))
			if err == nil && claims != nil && claims.UserId > 0 {
				if mdCtx, mdErr := ctxdata.PutUserIdToCtx(ctx, claims.UserId); mdErr == nil {
					ctx = mdCtx
				}
			}
		}

		l := search.NewSearchProductsLogic(ctx, svcCtx)
		resp, err := l.SearchProducts(&req)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, resp)
		}
	}
}
