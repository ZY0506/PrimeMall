// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package profile

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/user/profile"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

// 注销账号
func DeleteAccountHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteAccountReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ClientError(r.Context(), w, response.RequestError, err.Error())
			return
		}

		l := profile.NewDeleteAccountLogic(r.Context(), svcCtx)
		err := l.DeleteAccount(&req)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, nil)
		}
	}
}
