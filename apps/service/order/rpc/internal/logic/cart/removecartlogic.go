package cartlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveCartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveCartLogic {
	return &RemoveCartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RemoveCart 批量删除
func (l *RemoveCartLogic) RemoveCart(in *order.RemoveCartRequest) (*order.Empty, error) {
	// 获取用户ID
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败,error=%v", err)
		return nil, err
	}

	// 批量删除购物车
	err = l.svcCtx.CartModel.BatchDelete(l.ctx, userId, in.SkuIds)
	if err != nil {
		l.Logger.Errorf("批量删除购物车失败,error=%v", err)
		return nil, err
	}
	l.Logger.Info("批量删除购物车成功")

	return &order.Empty{}, nil
}
