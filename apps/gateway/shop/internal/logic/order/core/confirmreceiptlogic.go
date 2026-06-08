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

type ConfirmReceiptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewConfirmReceiptLogic 确认收货
func NewConfirmReceiptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmReceiptLogic {
	return &ConfirmReceiptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfirmReceiptLogic) ConfirmReceipt(req *types.OrderSnPathReq) error {
	// 注入userID
	var err error
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return err
	}
	// 调用rpc
	_, err = l.svcCtx.OrderRpc.ConfirmReceipt(l.ctx, &order.OrderDetailRequest{
		OrderSn: req.OrderSn,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc确认收货失败，error=%v", err)
		return err
	}

	return nil
}
