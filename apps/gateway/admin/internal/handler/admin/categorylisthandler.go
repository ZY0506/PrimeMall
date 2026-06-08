// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package admin

import (
	"net/http"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/logic/admin"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func CategoryListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := admin.NewCategoryListLogic(r.Context(), svcCtx)
		resp, err := l.CategoryList()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
