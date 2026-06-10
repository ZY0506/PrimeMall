// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package address

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/user/address"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

// 设置默认收货地址
func SetDefaultAddressHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.IdPathReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ClientError(r.Context(), w, response.ErrCodeInvalidParam, err.Error())
			return
		}

		l := address.NewSetDefaultAddressLogic(r.Context(), svcCtx)
		err := l.SetDefaultAddress(&req)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, nil)
		}
	}
}
