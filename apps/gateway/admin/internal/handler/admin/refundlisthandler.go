// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package admin

import (
	"github.com/ZY0506/PrimeMall/common/response"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/logic/admin"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
)

func RefundListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminRefundListReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ClientError(r.Context(), w, response.ErrCodeInvalidParam, err.Error())
			return
		}

		l := admin.NewRefundListLogic(r.Context(), svcCtx)
		resp, err := l.RefundList(&req)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, resp)
		}
	}
}
