package productadminlogic

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ZY0506/PrimeMall/common/constants"
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
	// 清除商品详情缓存
	cacheKey := constants.ProductDetailKey + strconv.FormatUint(in.Id, 10)
	l.svcCtx.Client.Del(l.ctx, cacheKey)

	// 清除该分类下的商品列表缓存
	pattern := constants.ProductListKey + fmt.Sprintf("%d:*", spu.CategoryId)
	if keys, err := l.svcCtx.Client.Keys(l.ctx, pattern).Result(); err == nil && len(keys) > 0 {
		l.svcCtx.Client.Del(l.ctx, keys...)
	}
	return &product.Empty{}, nil
}
