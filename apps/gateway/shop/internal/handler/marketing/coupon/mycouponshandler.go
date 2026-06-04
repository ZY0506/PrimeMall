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

// 我的优惠券：查看个人卡券包
func MyCouponsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MyCouponListReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ClientError(r.Context(), w, response.RequestError, err.Error())
			return
		}

		l := coupon.NewMyCouponsLogic(r.Context(), svcCtx)
		resp, err := l.MyCoupons(&req)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, resp)
		}
	}
}
