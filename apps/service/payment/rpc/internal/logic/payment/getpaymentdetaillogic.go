package paymentlogic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/types/payment"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetPaymentDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPaymentDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentDetailLogic {
	return &GetPaymentDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetPaymentDetail 获取支付详情
// 核心职责：从Redis获取完整支付信息 → 组装响应
func (l *GetPaymentDetailLogic) GetPaymentDetail(in *payment.GetPaymentDetailRequest) (*payment.GetPaymentDetailResponse, error) {
	// 1. 入参校验
	if in.PaymentSn == "" {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "支付流水号不能为空")
	}

	// 2. 先尝试从JSON快照获取完整记录
	detailKey := fmt.Sprintf("payment:detail:%s", in.PaymentSn)
	detailJson, err := l.svcCtx.Client.Get(l.ctx, detailKey).Result()
	if err != nil {
		l.Logger.Errorf("查询支付详情快照失败: %v", err)
		// 降级：从Hash中逐字段读取
		return l.getPaymentDetailFromHash(in.PaymentSn)
	}

	// 3. 解析JSON到map
	var record map[string]interface{}
	if err := json.Unmarshal([]byte(detailJson), &record); err != nil {
		l.Logger.Errorf("解析支付详情快照失败: %v", err)
		return l.getPaymentDetailFromHash(in.PaymentSn)
	}

	// 4. 组装响应
	resp := &payment.GetPaymentDetailResponse{
		PaymentSn:      getStringField(record, "payment_sn"),
		OrderSn:        getStringField(record, "order_sn"),
		Amount:         getInt64Field(record, "amount"),
		PayType:        payment.PayType(getInt64Field(record, "pay_type")),
		Status:         payment.PaymentStatus(getInt64Field(record, "status")),
		ChannelOrderSn: getStringField(record, "channel_order_sn"),
		TransactionId:  getStringField(record, "transaction_id"),
		ErrorMsg:       getStringField(record, "error_msg"),
	}

	// 时间字段
	if t, err := time.Parse(time.RFC3339, getStringField(record, "created_at")); err == nil {
		resp.CreatedAt = timestamppb.New(t)
	}
	if t, err := time.Parse(time.RFC3339, getStringField(record, "pay_time")); err == nil {
		resp.PayTime = timestamppb.New(t)
	}

	l.Logger.Infof("查询支付详情成功: paymentSn=%s", in.PaymentSn)
	return resp, nil
}

// getPaymentDetailFromHash 从Hash中逐字段读取支付详情（降级方案）
func (l *GetPaymentDetailLogic) getPaymentDetailFromHash(paymentSn string) (*payment.GetPaymentDetailResponse, error) {
	paymentHashKey := fmt.Sprintf("payment:sn:%s", paymentSn)
	exists, err := l.svcCtx.Client.Exists(l.ctx, paymentHashKey).Result()
	if err != nil || exists == 0 {
		l.Logger.Errorf("支付单不存在: paymentSn=%s", paymentSn)
		return nil, errorx.NewBizError(response.ErrCodePaymentNotFound, "支付单不存在")
	}

	orderSn, _ := l.svcCtx.Client.HGet(l.ctx, paymentHashKey, "order_sn").Result()
	status, _ := l.svcCtx.Client.HGet(l.ctx, paymentHashKey, "status").Int()
	amount, _ := l.svcCtx.Client.HGet(l.ctx, paymentHashKey, "amount").Int64()
	payType, _ := l.svcCtx.Client.HGet(l.ctx, paymentHashKey, "pay_type").Int()
	channelOrderSn, _ := l.svcCtx.Client.HGet(l.ctx, paymentHashKey, "channel_order_sn").Result()
	transactionId, _ := l.svcCtx.Client.HGet(l.ctx, paymentHashKey, "transaction_id").Result()
	createdAtStr, _ := l.svcCtx.Client.HGet(l.ctx, paymentHashKey, "created_at").Result()
	payTimeStr, _ := l.svcCtx.Client.HGet(l.ctx, paymentHashKey, "pay_time").Result()
	errorMsg, _ := l.svcCtx.Client.HGet(l.ctx, paymentHashKey, "error_msg").Result()

	resp := &payment.GetPaymentDetailResponse{
		PaymentSn:      paymentSn,
		OrderSn:        orderSn,
		Amount:         amount,
		PayType:        payment.PayType(payType),
		Status:         payment.PaymentStatus(status),
		ChannelOrderSn: channelOrderSn,
		TransactionId:  transactionId,
		ErrorMsg:       errorMsg,
	}

	if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
		resp.CreatedAt = timestamppb.New(t)
	}
	if t, err := time.Parse(time.RFC3339, payTimeStr); err == nil {
		resp.PayTime = timestamppb.New(t)
	}

	return resp, nil
}

// 辅助函数 - 字段类型安全转换
func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt64Field(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return int64(val)
		case int64:
			return val
		case int:
			return int64(val)
		}
	}
	return 0
}
