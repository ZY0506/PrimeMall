package search

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSearchHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSearchHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSearchHistoryLogic {
	return &GetSearchHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSearchHistoryLogic) GetSearchHistory() (resp *types.SearchHistoryResp, err error) {
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}

	rpcResp, err := l.svcCtx.SearchRpc.GetSearchHistory(l.ctx, &search.SearchHistoryReq{})
	if err != nil {
		l.Logger.Errorf("GetSearchHistory RPC error: %v", err)
		return nil, err
	}

	list := make([]types.SearchHistoryItem, 0, len(rpcResp.List))
	for _, item := range rpcResp.List {
		list = append(list, types.SearchHistoryItem{
			Keyword:   item.Keyword,
			CreatedAt: item.CreatedAt,
		})
	}

	return &types.SearchHistoryResp{List: list}, nil
}
