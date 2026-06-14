// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 查询搜索历史
func getSearchHistoryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetSearchHistoryLogic(r.Context(), svcCtx)
		resp, err := l.GetSearchHistory()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
