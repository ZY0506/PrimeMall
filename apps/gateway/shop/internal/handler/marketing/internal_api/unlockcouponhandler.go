// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package internal_api

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/marketing/internal_api"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

// 订单服务：解锁优惠券（订单取消时回滚）
func UnlockCouponHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UnlockCouponReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ClientError(r.Context(), w, response.RequestError, err.Error())
			return
		}

		l := internal_api.NewUnlockCouponLogic(r.Context(), svcCtx)
		resp, err := l.UnlockCoupon(&req)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, resp)
		}
	}
}
