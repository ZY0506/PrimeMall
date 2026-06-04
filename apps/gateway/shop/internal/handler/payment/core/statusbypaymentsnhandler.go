// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package core

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/payment/core"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

// 根据支付流水号查询支付状态
func StatusByPaymentSnHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PaymentSnPathReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ClientError(r.Context(), w, response.RequestError, err.Error())
			return
		}

		l := core.NewStatusByPaymentSnLogic(r.Context(), svcCtx)
		resp, err := l.StatusByPaymentSn(&req)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, resp)
		}
	}
}
