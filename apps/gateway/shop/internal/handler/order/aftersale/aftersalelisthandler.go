// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package aftersale

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/order/aftersale"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

// 获取我的售后单列表（分页）
func AfterSaleListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AfterSaleListReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ClientError(r.Context(), w, response.ErrCodeInvalidParam, err.Error())
			return
		}

		l := aftersale.NewAfterSaleListLogic(r.Context(), svcCtx)
		resp, err := l.AfterSaleList(&req)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, resp)
		}
	}
}
