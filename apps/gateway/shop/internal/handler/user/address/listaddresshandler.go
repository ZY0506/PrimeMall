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

// 获取地址列表（分页）
func ListAddressHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AddressListReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ClientError(r.Context(), w, response.ErrCodeInvalidParam, err.Error())
			return
		}

		l := address.NewListAddressLogic(r.Context(), svcCtx)
		resp, err := l.ListAddress(&req)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, resp)
		}
	}
}
