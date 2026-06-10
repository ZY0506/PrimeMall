// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package cart

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/order/cart"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

// 批量删除购物车商品
func RemoveCartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SkuIdsReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ClientError(r.Context(), w, response.ErrCodeInvalidParam, err.Error())
			return
		}

		l := cart.NewRemoveCartLogic(r.Context(), svcCtx)
		err := l.RemoveCart(&req)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, nil)
		}
	}
}
