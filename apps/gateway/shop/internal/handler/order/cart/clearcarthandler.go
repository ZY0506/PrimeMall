// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package cart

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/order/cart"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/common/response"
	"net/http"
)

// 清空购物车
func ClearCartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := cart.NewClearCartLogic(r.Context(), svcCtx)
		err := l.ClearCart()
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, nil)
		}
	}
}
