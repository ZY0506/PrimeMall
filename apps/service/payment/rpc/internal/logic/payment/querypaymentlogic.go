package paymentlogic

import (
	"context"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/types/payment"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type QueryPaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryPaymentLogic {
	return &QueryPaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryPayment 查询支付状态
// 核心职责：从Redis中查询支付记录 → 返回支付状态
func (l *QueryPaymentLogic) QueryPayment(in *payment.QueryPaymentRequest) (*payment.QueryPaymentResponse, error) {
	// 1. 入参校验
	if in.PaymentSn == "" {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "支付流水号不能为空")
	}

	// 2. 从Redis查询支付记录
	paymentHashKey := fmt.Sprintf("payment:sn:%s", in.PaymentSn)
	exists, err := l.svcCtx.Client.Exists(l.ctx, paymentHashKey).Result()
	if err != nil {
		l.Logger.Errorf("查询支付记录失败: %v", err)
		return nil, err
	}
	if exists == 0 {
		l.Logger.Errorf("支付单不存在: paymentSn=%s", in.PaymentSn)
		return nil, errorx.NewBizError(response.ErrCodePaymentNotFound, "支付单不存在")
	}

	// 3. 获取各字段
	orderSn, _ := l.svcCtx.Client.HGet(l.ctx, paymentHashKey, "order_sn").Result()
	status, _ := l.svcCtx.Client.HGet(l.ctx, paymentHashKey, "status").Int()
	amount, _ := l.svcCtx.Client.HGet(l.ctx, paymentHashKey, "amount").Int64()
	payTimeStr, _ := l.svcCtx.Client.HGet(l.ctx, paymentHashKey, "pay_time").Result()

	var payTime *timestamppb.Timestamp
	if payTimeStr != "" {
		t, err := time.Parse(time.RFC3339, payTimeStr)
		if err == nil {
			payTime = timestamppb.New(t)
		}
	}

	l.Logger.Infof("查询支付状态成功: paymentSn=%s, status=%d", in.PaymentSn, status)

	return &payment.QueryPaymentResponse{
		PaymentSn: in.PaymentSn,
		OrderSn:   orderSn,
		Status:    payment.PaymentStatus(status),
		Amount:    amount,
		PayTime:   payTime,
	}, nil
}
