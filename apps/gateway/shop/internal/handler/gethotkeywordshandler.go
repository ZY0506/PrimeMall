// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 热搜词列表
func getHotKeywordsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetHotKeywordsLogic(r.Context(), svcCtx)
		resp, err := l.GetHotKeywords()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
