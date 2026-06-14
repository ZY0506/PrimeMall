// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchSuggestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// ES搜索建议
func NewSearchSuggestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchSuggestLogic {
	return &SearchSuggestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchSuggestLogic) SearchSuggest(req *types.SearchSuggestReq) (resp *types.SearchSuggestResp, err error) {
	// todo: add your logic here and delete this line

	return
}
