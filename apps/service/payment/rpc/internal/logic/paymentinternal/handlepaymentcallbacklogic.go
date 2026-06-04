package paymentinternallogic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/client/orderinternal"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/types/payment"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type HandlePaymentCallbackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandlePaymentCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandlePaymentCallbackLogic {
	return &HandlePaymentCallbackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// HandlePaymentCallback 处理支付回调
func (l *HandlePaymentCallbackLogic) HandlePaymentCallback(in *payment.PaymentCallbackRequest) (*payment.PaymentCallbackResponse, error) {
	if in.PaymentSn == "" || in.OrderSn == "" {
		l.Logger.Errorf("回调参数不完整: paymentSn=%s, orderSn=%s", in.PaymentSn, in.OrderSn)
		return &payment.PaymentCallbackResponse{Success: false, Message: "参数不完整"}, nil
	}

	// 查询支付记录（MySQL）
	pay, err := l.svcCtx.PaymentModel.FindOneByPaymentSn(l.ctx, in.PaymentSn)
	if err != nil {
		l.Logger.Infof("支付记录不存在, 可能已过期: paymentSn=%s", in.PaymentSn)
		return &payment.PaymentCallbackResponse{Success: true, Message: "支付单不存在，忽略"}, nil
	}

	// 校验：订单号匹配
	if pay.OrderSn != in.OrderSn {
		l.Logger.Errorf("订单号不匹配: paymentSn=%s, 预期=%s, 实际=%s", in.PaymentSn, pay.OrderSn, in.OrderSn)
		return &payment.PaymentCallbackResponse{Success: false, Message: "订单号不匹配"}, nil
	}

	// 校验：幂等处理（已成功的不再处理）
	if pay.Status == int64(payment.PaymentStatus_PAYMENT_STATUS_SUCCESS) {
		l.Logger.Infof("支付已处理，幂等跳过: paymentSn=%s", in.PaymentSn)
		return &payment.PaymentCallbackResponse{Success: true, Message: "已处理"}, nil
	}

	now := time.Now()

	// 记录回调日志
	callbackData := ""
	if in.RawData != "" {
		callbackData = in.RawData

	}
	_, _ = l.svcCtx.PaymentCallbackLogModel.Insert(l.ctx, &model.PaymentCallbackLog{
		PaymentSn:   in.PaymentSn,
		OrderSn:     in.OrderSn,
		Channel:     "mock",
		RequestBody: sql.NullString{String: callbackData, Valid: callbackData != ""},
		Status:      1,
	})

	// 更新MySQL支付记录
	pay.Status = int64(payment.PaymentStatus_PAYMENT_STATUS_SUCCESS)
	pay.PayTime = sql.NullTime{Time: now, Valid: true}
	pay.CallbackData = sql.NullString{String: callbackData, Valid: callbackData != ""}
	if err := l.svcCtx.PaymentModel.Update(l.ctx, pay); err != nil {
		l.Logger.Errorf("更新支付记录失败: %v", err)
		return &payment.PaymentCallbackResponse{Success: false, Message: "更新失败"}, nil
	}

	// 更新Redis缓存
	paymentHashKey := fmt.Sprintf("payment:sn:%s", in.PaymentSn)
	_ = l.svcCtx.Client.HSet(l.ctx, paymentHashKey, map[string]interface{}{
		"status":   int(payment.PaymentStatus_PAYMENT_STATUS_SUCCESS),
		"pay_time": now.Format(time.RFC3339),
	}).Err()

	// 异步通知订单服务
	go func() {
		defer func() {
			if r := recover(); r != nil {
				l.Logger.Errorf("支付回调通知订单协程panic: %v", r)
			}
		}()
		notifyCtx := context.Background()
		_, err := l.svcCtx.OrderInternalRpc.NotifyPaymentSuccess(notifyCtx, &orderinternal.PaymentSuccessNotifyRequest{
			PaymentSn: in.PaymentSn,
			OrderSn:   in.OrderSn,
			PayTime:   timestamppb.New(now),
		})
		if err != nil {
			l.Logger.Errorf("通知订单服务支付成功失败: %v", err)
		}
	}()

	l.Logger.Infof("支付回调处理成功: paymentSn=%s, orderSn=%s", in.PaymentSn, in.OrderSn)
	return &payment.PaymentCallbackResponse{Success: true, Message: "success"}, nil
}
