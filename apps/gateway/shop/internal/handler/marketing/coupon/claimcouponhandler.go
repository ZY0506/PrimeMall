// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package coupon

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/marketing/coupon"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

// 用户领取优惠券（幂等）
func ClaimCouponHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ClaimCouponReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ClientError(r.Context(), w, response.RequestError, err.Error())
			return
		}

		l := coupon.NewClaimCouponLogic(r.Context(), svcCtx)
		resp, err := l.ClaimCoupon(&req)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, resp)
		}
	}
}
