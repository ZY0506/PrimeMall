// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package admin

import (
	"net/http"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/logic/admin"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func BlacklistUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BlackUserReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := admin.NewBlacklistUserLogic(r.Context(), svcCtx)
		err := l.BlacklistUser(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
