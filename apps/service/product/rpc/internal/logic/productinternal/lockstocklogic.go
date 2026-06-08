package productinternallogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type LockStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLockStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LockStockLogic {
	return &LockStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// LockStock 锁定库存（创建订单预占）
func (l *LockStockLogic) LockStock(in *product.UpdateStockReq) (*product.StockChangeResp, error) {
	// 验证参数
	if in.OrderSn == "" {
		return nil, errorx.NewBizError(response.ErrCodeOrderNotFound, "订单号不能为空")
	}
	for _, item := range in.Items {
		if item.Quantity <= 0 {
			return nil, errorx.NewBizError(response.ErrCodeInvalidQuantity, "数量必须大于0")
		}
	}

	// 获取分布式锁（降低DB乐观锁竞争）
	skuIds := make([]uint64, len(in.Items))
	for i, item := range in.Items {
		skuIds[i] = item.SkuId
	}
	if !acquireStockLocks(l.ctx, l.svcCtx.Client, skuIds) {
		l.Logger.Errorf("订单：%s 获取库存锁超时", in.OrderSn)
		return nil, errorx.NewBizError(response.ErrCodeTooFrequent, "系统繁忙，请稍后重试")
	}
	defer releaseStockLocksByIds(l.ctx, l.svcCtx.Client, skuIds)

	// 锁定库存
	ret, err := l.svcCtx.ProductSkuModel.LockStock(l.ctx, in.Items, in.OrderSn)
	if err != nil {
		l.Logger.Errorf("订单：%s 锁定库存失败：%v", in.OrderSn, err)
		return nil, err
	}
	l.Logger.Infof("订单：%s 锁定库存成功", in.OrderSn)

	return &product.StockChangeResp{
		Success: true,
		Results: *ret,
	}, nil
}
