// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package core

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/order/core"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	"net/http"
)

// 取消订单（仅待支付状态可取消）
func CancelOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CancelOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ClientError(r.Context(), w, response.ErrCodeInvalidParam, err.Error())
			return
		}

		// 尝试从 path 参数获取 order_sn（httpx.Parse 在某些版本中可能不会自动映射 path 参数到 path tag）
		if req.OrderSn == "" {
			vars := pathvar.Vars(r)
			if sn, ok := vars["order_sn"]; ok {
				req.OrderSn = sn
			}
		}

		if req.OrderSn == "" {
			response.ClientError(r.Context(), w, response.ErrCodeInvalidParam, "订单号不能为空")
			return
		}

		l := core.NewCancelOrderLogic(r.Context(), svcCtx)
		err := l.CancelOrder(&req)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, nil)
		}
	}
}
