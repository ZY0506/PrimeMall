package searchlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncProductToESLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncProductToESLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncProductToESLogic {
	return &SyncProductToESLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步商品到ES
func (l *SyncProductToESLogic) SyncProductToES(in *search.SyncProductReq) (*search.Empty, error) {
	if l.svcCtx.ES == nil {
		return &search.Empty{}, nil
	}

	_ = l.svcCtx.ES.EnsureIndex(l.ctx)

	// In production, fetch from product service via RPC.
	// For now we just ensure the index exists.
	// Actual sync should be triggered via MQ or admin API with full product data.
	doc := &svc.ProductDoc{
		ID: in.ProductId,
	}
	if err := l.svcCtx.ES.IndexProduct(l.ctx, doc); err != nil {
		l.Logger.Errorf("同步商品到ES：索引失败，错误：%v", err)
		return nil, err
	}
	return &search.Empty{}, nil
}
