package callback

import (
	"io"
	"net/http"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/payment/callback"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/common/response"
)

// 支付宝异步回调（服务端自行幂等）
// 接收支付宝的异步回调通知，解析请求体后调用 Logic 层处理
func AlipayCallbackHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 读取请求体
		body, err := io.ReadAll(r.Body)
		if err != nil {
			response.LogicError(r.Context(), w, response.NewBizError(response.ErrCodeInvalidParam, "读取回调数据失败"))
			return
		}
		defer r.Body.Close()

		l := callback.NewAlipayCallbackLogic(r.Context(), svcCtx)
		resp, err := l.AlipayCallback(body)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, resp)
		}
	}
}
