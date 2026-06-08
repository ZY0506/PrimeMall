package core

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/types/payment"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClosePaymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewClosePaymentLogic 关闭支付单（未支付时）
func NewClosePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClosePaymentLogic {
	return &ClosePaymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ClosePayment 关闭支付单
// 业务逻辑：注入userID → 查询支付单状态 → 若为待支付则调用支付服务关闭（模拟实现直接返回成功）
func (l *ClosePaymentLogic) ClosePayment(req *types.PaymentSnPathReq) error {
	// 注入userID
	var err error
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return err
	}

	// 通过PaymentRpc查询当前支付状态
	_, err = l.svcCtx.PaymentRpc.GetPaymentDetail(l.ctx, &payment.GetPaymentDetailRequest{
		PaymentSn: req.PaymentSn,
	})
	if err != nil {
		l.Logger.Errorf("查询支付详情失败，error=%v", err)
		return err
	}

	// 模拟关闭：支付服务是mock，直接返回成功
	// 实际项目中需要调用支付渠道的关闭订单接口
	l.Logger.Infof("模拟关闭支付单成功: paymentSn=%s", req.PaymentSn)

	return nil
}
