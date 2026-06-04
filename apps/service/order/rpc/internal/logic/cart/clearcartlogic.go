package cartlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearCartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearCartLogic {
	return &ClearCartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ClearCart 清空购物车
func (l *ClearCartLogic) ClearCart(in *order.Empty) (*order.Empty, error) {
	// 获取当前用户ID
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败,error=%v", err)
		return nil, err
	}

	// 清空购物车
	err = l.svcCtx.CartModel.ClearCart(l.ctx, userId)
	if err != nil {
		l.Logger.Errorf("清空购物车失败,error=%v", err)
		return nil, err
	}
	l.Logger.Info("清空购物车成功")

	return &order.Empty{}, nil
}
