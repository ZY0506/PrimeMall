package search

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearSearchHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClearSearchHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearSearchHistoryLogic {
	return &ClearSearchHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ClearSearchHistoryLogic) ClearSearchHistory() (resp *types.EmptyResp, err error) {
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}

	_, err = l.svcCtx.SearchRpc.ClearSearchHistory(l.ctx, &search.ClearSearchHistoryReq{})
	if err != nil {
		l.Logger.Errorf("ClearSearchHistory RPC error: %v", err)
		return nil, err
	}

	return &types.EmptyResp{}, nil
}
