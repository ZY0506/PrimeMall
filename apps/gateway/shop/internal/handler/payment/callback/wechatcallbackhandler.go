// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package callback

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/payment/callback"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/common/response"
	"net/http"
)

// 微信支付异步回调（服务端自行幂等）
func WechatCallbackHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := callback.NewWechatCallbackLogic(r.Context(), svcCtx)
		resp, err := l.WechatCallback()
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, resp)
		}
	}
}
