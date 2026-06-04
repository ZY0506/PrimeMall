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

type RemoveCartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRemoveCartLogic 批量删除购物车商品
func NewRemoveCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveCartLogic {
	return &RemoveCartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveCartLogic) RemoveCart(req *types.SkuIdsReq) error {
	// 注入userID
	var err error
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return err
	}
	// 调用rpc
	_, err = l.svcCtx.CartRpc.RemoveCart(l.ctx, &order.RemoveCartRequest{
		SkuIds: req.SkuIds,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc批量删除购物车商品失败，error=%v", err)
		return err
	}

	return nil
}
