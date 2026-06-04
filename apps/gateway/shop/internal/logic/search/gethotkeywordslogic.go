package search

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetHotKeywordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetHotKeywordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHotKeywordsLogic {
	return &GetHotKeywordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHotKeywordsLogic) GetHotKeywords() (resp *types.HotKeywordsResp, err error) {
	rpcResp, err := l.svcCtx.SearchRpc.GetHotKeywords(l.ctx, &search.Empty{})
	if err != nil {
		l.Logger.Errorf("GetHotKeywords RPC error: %v", err)
		return nil, err
	}

	list := make([]types.HotKeywordItem, 0, len(rpcResp.List))
	for _, item := range rpcResp.List {
		list = append(list, types.HotKeywordItem{
			Id:          item.Id,
			Keyword:     item.Keyword,
			SearchCount: item.SearchCount,
			Sort:        item.Sort,
		})
	}

	return &types.HotKeywordsResp{List: list}, nil
}
