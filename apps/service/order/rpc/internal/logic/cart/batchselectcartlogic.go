package cartlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchSelectCartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchSelectCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchSelectCartLogic {
	return &BatchSelectCartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// BatchSelectCart 批量修改选中状态（支持全选/全不选）
func (l *BatchSelectCartLogic) BatchSelectCart(in *order.BatchSelectCartRequest) (*order.Empty, error) {
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败,error=%v", err)
		return nil, err
	}

	// 批量修改购物车选中状态
	err = l.svcCtx.CartModel.BatchSelect(l.ctx, userId, in.SkuIds, in.Selected, in.All)
	if err != nil {
		l.Logger.Errorf("批量修改购物车选中状态失败,error=%v", err)
		return nil, err
	}
	l.Logger.Info("批量修改购物车选中状态成功")
	return &order.Empty{}, nil
}
