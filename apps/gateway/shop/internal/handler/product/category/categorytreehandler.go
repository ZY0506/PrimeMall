// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package category

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/product/category"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/common/response"
	"net/http"
)

// 获取商品全部分类树
func CategoryTreeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := category.NewCategoryTreeLogic(r.Context(), svcCtx)
		resp, err := l.CategoryTree()
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, resp)
		}
	}
}
