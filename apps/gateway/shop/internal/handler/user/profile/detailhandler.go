// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package profile

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/user/profile"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/common/response"
	"net/http"
)

// 获取个人资料
func DetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := profile.NewDetailLogic(r.Context(), svcCtx)
		resp, err := l.Detail()
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, resp)
		}
	}
}
