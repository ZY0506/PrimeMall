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

type UpdateProductStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductStatusLogic {
	return &UpdateProductStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateProductStatusLogic) UpdateProductStatus(in *product.UpdateProductStatusReq) (*product.Empty, error) {
	if in.Id == 0 {
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "商品ID不能为空")
	}
	if in.Status == product.ProductStatus_PRODUCT_STATUS_UNKNOWN {
		return nil, errorx.NewBizError(response.ErrCodeParamType, "状态参数无效")
	}

	spu, err := l.svcCtx.ProductSpuModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewBizError(response.ErrCodeProductNotFound, "商品不存在")
		}
		l.Logger.Errorf("查询商品失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	spu.Status = int64(in.Status)
	err = l.svcCtx.ProductSpuModel.Update(l.ctx, spu)
	if err != nil {
		l.Logger.Errorf("更新商品状态失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	l.Logger.Infof("更新商品状态成功, id=%d, status=%d", in.Id, in.Status)
	return &product.Empty{}, nil
}
