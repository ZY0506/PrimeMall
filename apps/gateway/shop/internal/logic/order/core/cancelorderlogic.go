// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package core

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCancelOrderLogic 取消订单（仅待支付状态可取消）
func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelOrderLogic) CancelOrder(req *types.CancelOrderReq) error {
	// 注入userID
	var err error
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return err
	}
	// 调用rpc
	_, err = l.svcCtx.OrderRpc.CancelOrder(l.ctx, &order.CancelOrderRequest{
		OrderSn:    req.OrderSn,
		Reason:     req.CancelReason,
		ReasonType: order.CancelReasonType_CANCEL_REASON_TYPE_USER,
	})
	if err != nil {
		l.Logger.Errorf("调用RPC取消订单失败，error=%v", err)
		return err
	}

	return nil
}
