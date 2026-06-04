// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package cart

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SelectCartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSelectCartLogic 修改购物车商品的勾选状态
func NewSelectCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SelectCartLogic {
	return &SelectCartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SelectCartLogic) SelectCart(req *types.SelectCartReq) error {
	// 注入userID
	var err error
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return err
	}
	// 调用rpc
	_, err = l.svcCtx.CartRpc.SelectCart(l.ctx, &order.SelectCartRequest{
		SkuId:    req.SkuId,
		Selected: req.Selected,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc修改购物车商品勾选状态失败,error=%v", err)
		return err
	}
	return nil
}
