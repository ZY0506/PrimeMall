// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package profile

import (
	"github.com/ZY0506/PrimeMall/common/response"
	"net/http"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/user/profile"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 退出登录
func LogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RefreshTokenReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ClientError(r.Context(), w, response.ErrCodeInvalidParam, err.Error())
			return
		}

		l := profile.NewLogoutLogic(r.Context(), svcCtx)
		err := l.Logout(&req)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, nil)
		}
	}
}
