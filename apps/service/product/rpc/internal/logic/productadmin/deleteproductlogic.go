package productadminlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductLogic {
	return &DeleteProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteProductLogic) DeleteProduct(in *product.IdReq) (*product.Empty, error) {
	if in.Id == 0 {
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "商品ID不能为空")
	}

	spu, err := l.svcCtx.ProductSpuModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewBizError(response.ErrCodeProductNotFound, "商品不存在")
		}
		l.Logger.Errorf("查询商品失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	err = l.svcCtx.ProductSpuModel.Delete(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("删除商品失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	l.Logger.Infof("删除商品成功, id=%d, name=%s", in.Id, spu.Name)
	return &product.Empty{}, nil
}
