package search

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"
	"github.com/zeromicro/go-zero/core/logx"
)

type SearchSuggestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchSuggestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchSuggestLogic {
	return &SearchSuggestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchSuggestLogic) SearchSuggest(req *types.SearchSuggestReq) (resp *types.SearchSuggestResp, err error) {
	rpcResp, err := l.svcCtx.SearchRpc.SuggestProducts(l.ctx, &search.SuggestReq{
		Keyword: req.Keyword,
		Size:    int32(req.Size),
	})
	if err != nil {
		l.Logger.Errorf("SearchSuggest RPC error: %v", err)
		return nil, err
	}

	if rpcResp.Suggestions == nil {
		rpcResp.Suggestions = []string{}
	}

	return &types.SearchSuggestResp{
		Suggestions: rpcResp.Suggestions,
	}, nil
}
