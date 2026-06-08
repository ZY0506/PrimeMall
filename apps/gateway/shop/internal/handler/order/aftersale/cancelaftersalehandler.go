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

// 取消售后申请（仅待审核状态可取消）
func CancelAfterSaleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.IdPathReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ClientError(r.Context(), w, response.RequestError, err.Error())
			return
		}

		l := aftersale.NewCancelAfterSaleLogic(r.Context(), svcCtx)
		err := l.CancelAfterSale(&req)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, nil)
		}
	}
}
