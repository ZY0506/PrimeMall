package paymentinternallogic

import (
	"context"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/types/payment"
	"github.com/zeromicro/go-zero/core/logx"
)

type HandleRefundCallbackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleRefundCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleRefundCallbackLogic {
	return &HandleRefundCallbackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// HandleRefundCallback 处理退款回调（模拟实现）
// 核心职责：接收模拟退款回调 → 更新退款状态 → 返回成功
func (l *HandleRefundCallbackLogic) HandleRefundCallback(in *payment.RefundCallbackRequest) (*payment.RefundCallbackResponse, error) {
	// 1. 入参校验
	if in.RefundSn == "" || in.PaymentSn == "" {
		l.Logger.Infof("退款回调参数不完整: refundSn=%s, paymentSn=%s", in.RefundSn, in.PaymentSn)
		return &payment.RefundCallbackResponse{
			Success: false,
			Message: "退款回调参数不完整",
		}, nil
	}

	// 2. 更新支付单状态为已退款
	paymentHashKey := fmt.Sprintf("payment:sn:%s", in.PaymentSn)
	exists, _ := l.svcCtx.Client.Exists(l.ctx, paymentHashKey).Result()
	if exists > 0 {
		now := time.Now()
		_ = l.svcCtx.Client.HSet(l.ctx, paymentHashKey, "status", int(payment.PaymentStatus_PAYMENT_STATUS_REFUNDED)).Err()
		_ = l.svcCtx.Client.HSet(l.ctx, paymentHashKey, "refund_time", now.Format(time.RFC3339)).Err()
	}

	l.Logger.Infof("退款回调处理成功: refundSn=%s, paymentSn=%s, status=%d",
		in.RefundSn, in.PaymentSn, in.RefundStatus)

	return &payment.RefundCallbackResponse{
		Success: true,
		Message: "success",
	}, nil
}
