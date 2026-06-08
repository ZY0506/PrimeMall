// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/auth"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

// ResetPasswordHandler 忘记密码重置（幂等）
func ResetPasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ResetPasswordReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ClientError(r.Context(), w, response.RequestError, err.Error())
			return
		}

		l := auth.NewResetPasswordLogic(r.Context(), svcCtx)
		err := l.ResetPassword(&req)
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, nil)
		}
	}
}
