package searchlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteProductFromESLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteProductFromESLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductFromESLogic {
	return &DeleteProductFromESLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 从ES删除商品
func (l *DeleteProductFromESLogic) DeleteProductFromES(in *search.DeleteProductReq) (*search.Empty, error) {
	if l.svcCtx.ES == nil {
		return &search.Empty{}, nil
	}

	if err := l.svcCtx.ES.DeleteProduct(l.ctx, in.ProductId); err != nil {
		l.Logger.Errorf("从ES删除商品：删除失败，错误：%v", err)
		return nil, err
	}
	return &search.Empty{}, nil
}
