// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package category

import (
	"github.com/ZY0506/PrimeMall/common/response"
	"net/http"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/logic/category"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
)

// 分类列表
func CategoryListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := category.NewCategoryListLogic(r.Context(), svcCtx)
		resp, err := l.CategoryList()
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, resp)
		}
	}
}
