package paymentlogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/client/order"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/types/payment"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CreatePaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaymentLogic {
	return &CreatePaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreatePayment 创建支付单
func (l *CreatePaymentLogic) CreatePayment(in *payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	if in.OrderSn == "" {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "订单号不能为空")
	}
	if in.IdempotencyKey == "" {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "幂等键不能为空")
	}

	// 幂等校验
	idempotencyKey := fmt.Sprintf("%s%s", constants.IDEMPOTENCY_KEY+constants.PAYMENT_SERVICE, in.IdempotencyKey)
	exists, err := l.svcCtx.Client.SetNX(l.ctx, idempotencyKey, "1", constants.IDEMPOTENCY_EXIRE).Result()
	if err != nil {
		l.Logger.Errorf("幂等性校验失败: %v", err)
		return nil, errorx.NewBizError(response.InternalError, "系统繁忙，请稍后重试")
	}
	if !exists {
		paymentKey := fmt.Sprintf("%s%s", constants.PAYMENT_IDEMPOTENCY_PREFIX, in.IdempotencyKey)
		paymentSn, err := l.svcCtx.Client.Get(l.ctx, paymentKey).Result()
		if err == nil && paymentSn != "" {
			return l.queryExistingPayment(paymentSn)
		}
		return nil, errorx.NewBizError(response.ErrCodeTooFrequent, "请勿重复提交")
	}
	defer func() {
		if err != nil {
			_ = l.svcCtx.Client.Del(l.ctx, idempotencyKey).Err()
		}
	}()

	// 查询订单获取真实支付金额
	// 注意：从当前上下文提取 user_id 并注入到 order.rpc 调用的 metadata 中
	orderCtx := l.ctx
	if md, ok := metadata.FromIncomingContext(l.ctx); ok {
		orderCtx = metadata.NewOutgoingContext(l.ctx, md)
	}
	orderDetail, err := l.svcCtx.OrderRpc.OrderDetail(orderCtx, &order.OrderDetailRequest{
		OrderSn: in.OrderSn,
	})
	if err != nil {
		l.Logger.Errorf("查询订单详情失败, orderSn=%s, err=%v", in.OrderSn, err)
		// 映射常见错误码到更有意义的提示
		if code, msg, ok := errorx.ParseBizError(err); ok {
			l.Logger.Errorf("订单详情业务错误, code=%d, msg=%s", code, msg)
			return nil, err
		}
		return nil, errorx.NewBizError(response.ErrCodeOrderNotFound, "订单不存在")
	}
	if orderDetail.Base == nil {
		l.Logger.Errorf("订单详情为空, orderSn=%s", in.OrderSn)
		return nil, errorx.NewBizError(response.ErrCodeOrderNotFound, "订单不存在")
	}

	// 校验订单状态：仅待支付订单可以发起支付
	if int32(orderDetail.Base.Status) != constants.ORDER_STATUS_PENDING_PAY {
		l.Logger.Errorf("订单状态不允许支付, orderSn=%s, status=%v", in.OrderSn, orderDetail.Base.Status)
		return nil, errorx.NewBizError(response.ErrCodeOrderExpired, "订单已超时或状态异常，无法支付")
	}

	// 生成支付流水号
	paymentSn, err := l.svcCtx.IDGenerator.GenWithPrefix("PAY")
	if err != nil {
		l.Logger.Errorf("生成支付流水号失败: %v", err)
		return nil, errorx.NewBizError(response.InternalError, "系统繁忙，请稍后重试")
	}
	now := time.Now()
	expireTime := orderDetail.ExpireTime.AsTime()

	// 构建支付记录（模拟实现，直接设为成功）
	paymentRecord := map[string]interface{}{
		"payment_sn":       paymentSn,
		"order_sn":         in.OrderSn,
		"pay_type":         int64(in.PayType),
		"status":           int(payment.PaymentStatus_PAYMENT_STATUS_SUCCESS),
		"amount":           orderDetail.Base.PayAmount,
		"channel_order_sn": fmt.Sprintf("MOCK_CHANNEL_%s", paymentSn),
		"transaction_id":   fmt.Sprintf("MOCK_TXN_%s", paymentSn),
		"client_ip":        in.ClientIp,
		"idempotency_key":  in.IdempotencyKey,
		"created_at":       now.Format(time.RFC3339),
		"pay_time":         now.Format(time.RFC3339),
		"expire_time":      expireTime.Format(time.RFC3339),
		"error_msg":        "",
	}
	dataJson, _ := json.Marshal(paymentRecord)

	// 存储到Redis
	paymentHashKey := fmt.Sprintf("%s%s", constants.PAYMENT_SN_PREFIX, paymentSn)
	err = l.svcCtx.Client.HSet(l.ctx, paymentHashKey, paymentRecord).Err()
	if err != nil {
		l.Logger.Errorf("存储支付记录到Redis失败: %v", err)
		return nil, errorx.NewBizError(response.InternalError, "系统繁忙，请稍后重试")
	}
	_ = l.svcCtx.Client.Expire(l.ctx, paymentHashKey, constants.PAYMENT_CACHE_EXPIRE).Err()

	orderIndexKey := fmt.Sprintf("%s%s", constants.PAYMENT_ORDER_PREFIX, in.OrderSn)
	_ = l.svcCtx.Client.Set(l.ctx, orderIndexKey, paymentSn, constants.PAYMENT_CACHE_EXPIRE).Err()
	_ = l.svcCtx.Client.Set(l.ctx, fmt.Sprintf("%s%s", constants.PAYMENT_IDEMPOTENCY_PREFIX, in.IdempotencyKey), paymentSn, constants.PAYMENT_CACHE_EXPIRE).Err()
	_ = l.svcCtx.Client.Set(l.ctx, fmt.Sprintf("%s%s", constants.PAYMENT_DETAIL_PREFIX, paymentSn), dataJson, constants.PAYMENT_CACHE_EXPIRE).Err()

	// 持久化到MySQL
	payTime := sql.NullTime{Time: now, Valid: true}
	_, err = l.svcCtx.PaymentModel.Insert(l.ctx, &model.Payment{
		PaymentSn:      paymentSn,
		OrderSn:        in.OrderSn,
		Amount:         orderDetail.Base.PayAmount,
		Channel:        "mock",
		ChannelOrderSn: fmt.Sprintf("MOCK_CHANNEL_%s", paymentSn),
		TransactionId:  fmt.Sprintf("MOCK_TXN_%s", paymentSn),
		Status:         int64(payment.PaymentStatus_PAYMENT_STATUS_SUCCESS),
		PayTime:        payTime,
	})
	if err != nil {
		l.Logger.Errorf("持久化支付记录到MySQL失败: %v", err)
		// Redis回滚
		_ = l.svcCtx.Client.Del(l.ctx, paymentHashKey, orderIndexKey).Err()
		return nil, errorx.NewBizError(response.InternalError, "系统繁忙，请稍后重试")
	}

	mockPayParams := fmt.Sprintf(`{"payment_sn":"%s","mock_url":"https://mock.pay.example.com/pay?sn=%s"}`, paymentSn, paymentSn)
	l.Logger.Infof("创建支付单成功: paymentSn=%s, orderSn=%s, amount=%d", paymentSn, in.OrderSn, orderDetail.Base.PayAmount)

	return &payment.CreatePaymentResponse{
		PaymentSn:  paymentSn,
		PayParams:  mockPayParams,
		ExpireTime: orderDetail.ExpireTime,
	}, nil
}

// queryExistingPayment 查询已存在的支付单
func (l *CreatePaymentLogic) queryExistingPayment(paymentSn string) (*payment.CreatePaymentResponse, error) {
	paymentHashKey := fmt.Sprintf("%s%s", constants.PAYMENT_SN_PREFIX, paymentSn)
	exists, err := l.svcCtx.Client.Exists(l.ctx, paymentHashKey).Result()
	if err != nil || exists == 0 {
		return nil, errorx.NewBizError(response.ErrCodePaymentNotFound, "支付单不存在")
	}
	mockPayParams := fmt.Sprintf(`{"payment_sn":"%s","mock_url":"https://mock.pay.example.com/pay?sn=%s"}`, paymentSn, paymentSn)

	expireTimeStr, _ := l.svcCtx.Client.HGet(l.ctx, paymentHashKey, "expire_time").Result()
	var expireTime time.Time
	if expireTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, expireTimeStr); err == nil {
			expireTime = t
		} else {
			return nil, errorx.NewBizError(response.InternalError, "支付单数据异常")
		}
	}
	l.Logger.Infof("查询已存在支付单: paymentSn=%s", paymentSn)
	return &payment.CreatePaymentResponse{
		PaymentSn:  paymentSn,
		PayParams:  mockPayParams,
		ExpireTime: timestamppb.New(expireTime),
	}, nil
}
