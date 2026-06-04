// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package cart

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ClearCartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewClearCartLogic 清空购物车
func NewClearCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearCartLogic {
	return &ClearCartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ClearCartLogic) ClearCart() error {
	// 注入userID
	var err error
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return err
	}
	// 调用rpc
	_, err = l.svcCtx.CartRpc.ClearCart(l.ctx, &order.Empty{})
	if err != nil {
		l.Logger.Errorf("调用rpc清空购物车失败，error=%v", err)
		return err
	}

	return nil
}
