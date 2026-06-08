package searchlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchSyncProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchSyncProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchSyncProductsLogic {
	return &BatchSyncProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量同步商品到ES
func (l *BatchSyncProductsLogic) BatchSyncProducts(in *search.BatchSyncProductsReq) (*search.Empty, error) {
	if l.svcCtx.ES == nil {
		return &search.Empty{}, nil
	}

	_ = l.svcCtx.ES.EnsureIndex(l.ctx)

	for _, pid := range in.ProductIds {
		doc := &svc.ProductDoc{ID: pid}
		if err := l.svcCtx.ES.IndexProduct(l.ctx, doc); err != nil {
			l.Logger.Errorf("批量同步商品：索引商品%d失败，错误：%v", pid, err)
			continue
		}
	}
	return &search.Empty{}, nil
}
