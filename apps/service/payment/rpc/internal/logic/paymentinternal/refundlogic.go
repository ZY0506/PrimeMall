package paymentinternallogic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/client/orderinternal"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/types/payment"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
)

type RefundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundLogic {
	return &RefundLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Refund 执行退款（模拟实现）
// 核心职责：校验支付单 → 更新支付单状态为退款中 → 直接标记退款成功（模拟）
func (l *RefundLogic) Refund(in *payment.RefundRequest) (*payment.RefundResponse, error) {
	// 1. 入参校验
	if in.RefundSn == "" || in.PaymentSn == "" || in.OrderSn == "" {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "退款参数不完整")
	}
	if in.RefundAmount <= 0 {
		return nil, errorx.NewBizError(response.ErrCodeRefundAmountExceed, "退款金额必须大于0")
	}

	// 2. 查询原支付单（以MySQL为权威数据源）
	pay, err := l.svcCtx.PaymentModel.FindOneByPaymentSn(l.ctx, in.PaymentSn)
	if err != nil {
		l.Logger.Errorf("支付记录不存在: paymentSn=%s", in.PaymentSn)
		return nil, errorx.NewBizError(response.ErrCodePaymentNotFound, "支付单不存在")
	}

	// 3. 校验支付单状态（仅已支付的订单可退款）
	if pay.Status != int64(payment.PaymentStatus_PAYMENT_STATUS_SUCCESS) {
		l.Logger.Errorf("支付单状态不允许退款: paymentSn=%s, status=%d", in.PaymentSn, pay.Status)
		return nil, errorx.NewBizError(response.ErrCodeRefundStatusInvalid, "支付单状态不允许退款")
	}

	// 4. 更新MySQL支付单状态为退款中（权威数据）
	pay.Status = int64(payment.PaymentStatus_PAYMENT_STATUS_REFUNDING)
	if err := l.svcCtx.PaymentModel.Update(l.ctx, pay); err != nil {
		l.Logger.Errorf("更新支付单状态为退款中失败: %v", err)
		return nil, err
	}

	// 5. 同步更新Redis缓存（仅作缓存，过期自动失效）
	now := time.Now()
	paymentHashKey := fmt.Sprintf("payment:sn:%s", in.PaymentSn)
	_ = l.svcCtx.Client.HSet(l.ctx, paymentHashKey, map[string]interface{}{
		"status":    int(payment.PaymentStatus_PAYMENT_STATUS_REFUNDING),
		"error_msg": "",
		"refund_sn": in.RefundSn,
		"refund_at": now.Format(time.RFC3339),
	}).Err()

	l.Logger.Infof("退款处理中: refundSn=%s, paymentSn=%s, amount=%d", in.RefundSn, in.PaymentSn, in.RefundAmount)

	// 模拟：立即退款成功
	pay.Status = int64(payment.PaymentStatus_PAYMENT_STATUS_REFUNDED)
	if err := l.svcCtx.PaymentModel.Update(l.ctx, pay); err != nil {
		l.Logger.Errorf("更新支付单状态为已退款失败: %v", err)
		return nil, err
	}

	// 同步更新Redis缓存
	_ = l.svcCtx.Client.HSet(l.ctx, paymentHashKey, "status", int(payment.PaymentStatus_PAYMENT_STATUS_REFUNDED)).Err()

	// 6. 存储退款记录到Redis（缓存）
	refundKey := fmt.Sprintf("payment:refund:%s", in.RefundSn)
	refundRecord := map[string]interface{}{
		"refund_sn":       in.RefundSn,
		"payment_sn":      in.PaymentSn,
		"order_sn":        in.OrderSn,
		"after_sale_id":   in.AfterSaleId,
		"refund_amount":   in.RefundAmount,
		"pay_type":        int64(in.PayType),
		"reason":          in.Reason,
		"operator":        in.Operator,
		"status":          int(payment.RefundStatus_REFUND_STATUS_SUCCESS),
		"third_refund_no": fmt.Sprintf("MOCK_REFUND_%s", in.RefundSn),
		"refund_time":     now.Format(time.RFC3339),
		"created_at":      now.Format(time.RFC3339),
	}
	refundJson, _ := json.Marshal(refundRecord)
	_ = l.svcCtx.Client.Set(l.ctx, refundKey, refundJson, 24*time.Hour).Err()

	l.Logger.Infof("退款成功: refundSn=%s, paymentSn=%s, amount=%d", in.RefundSn, in.PaymentSn, in.RefundAmount)

	// 7. 通知订单服务：认领退款 + 回补库存（退货退款）+ 恢复订单状态
	l.notifyOrderRefundSuccess(in)

	return &payment.RefundResponse{
		RefundSn:           in.RefundSn,
		ThirdPartyRefundNo: fmt.Sprintf("MOCK_REFUND_%s", in.RefundSn),
		Status:             payment.RefundStatus_REFUND_STATUS_SUCCESS,
	}, nil
}

// notifyOrderRefundSuccess 通知订单服务退款已完成，带有限重试。
//
// 失败不回滚退款状态——钱确实已经退出去了，把 REFUNDED 改回去只会制造更严重的不一致。
// 通知本身是幂等的（订单侧以退款单号为键回补库存、以 CAS 认领退款状态），
// 所以重试以及未来任何形式的重投都是安全的。
func (l *RefundLogic) notifyOrderRefundSuccess(in *payment.RefundRequest) {
	const maxAttempts = 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := l.svcCtx.OrderInternalRpc.NotifyRefundSuccess(ctx, &orderinternal.RefundSuccessNotifyRequest{
			RefundSn:    in.RefundSn,
			PaymentSn:   in.PaymentSn,
			OrderSn:     in.OrderSn,
			AfterSaleId: in.AfterSaleId,
		})
		cancel()
		if err == nil {
			l.Logger.Infof("通知订单服务退款成功完成: refundSn=%s, attempt=%d", in.RefundSn, attempt)
			return
		}
		l.Logger.Errorf("通知订单服务退款成功失败: refundSn=%s, attempt=%d/%d, err=%v",
			in.RefundSn, attempt, maxAttempts, err)
		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
		}
	}
	l.Logger.Errorf("通知订单服务退款成功重试耗尽（需人工处理）: refundSn=%s, orderSn=%s", in.RefundSn, in.OrderSn)
}
