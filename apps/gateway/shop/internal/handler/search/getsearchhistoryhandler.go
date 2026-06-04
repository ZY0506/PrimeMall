package search

import (
	"net/http"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/search"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/common/response"
)

func GetSearchHistoryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := search.NewGetSearchHistoryLogic(r.Context(), svcCtx)
		resp, err := l.GetSearchHistory()
		if err != nil {
			response.LogicError(r.Context(), w, err)
		} else {
			response.Success(w, r, resp)
		}
	}
}
