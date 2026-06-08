package paymentinternallogic

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

type QueryRefundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryRefundLogic {
	return &QueryRefundLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryRefund 查询退款状态
// 核心职责：从Redis查询退款记录 → 返回退款状态信息
func (l *QueryRefundLogic) QueryRefund(in *payment.QueryRefundRequest) (*payment.QueryRefundResponse, error) {
	// 1. 入参校验
	if in.RefundSn == "" {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "退款单号不能为空")
	}

	// 2. 从Redis查询退款记录
	refundKey := fmt.Sprintf("payment:refund:%s", in.RefundSn)
	refundJson, err := l.svcCtx.Client.Get(l.ctx, refundKey).Result()
	if err != nil {
		l.Logger.Errorf("查询退款记录失败: refundSn=%s, err=%v", in.RefundSn, err)
		return nil, errorx.NewBizError(response.ErrCodeRefundNotFound, "退款单不存在")
	}

	// 3. 解析
	var record map[string]interface{}
	if err := json.Unmarshal([]byte(refundJson), &record); err != nil {
		l.Logger.Errorf("解析退款记录失败: %v", err)
		return nil, err
	}

	// 4. 组装响应
	resp := &payment.QueryRefundResponse{
		RefundSn:           getString(record, "refund_sn"),
		PaymentSn:          getString(record, "payment_sn"),
		OrderSn:            getString(record, "order_sn"),
		AfterSaleId:        getUint64(record, "after_sale_id"),
		RefundAmount:       getInt64(record, "refund_amount"),
		Status:             payment.RefundStatus(getInt64(record, "status")),
		ThirdPartyRefundNo: getString(record, "third_refund_no"),
		ErrorMsg:           getString(record, "error_msg"),
	}

	if t, err := time.Parse(time.RFC3339, getString(record, "refund_time")); err == nil {
		resp.RefundTime = timestamppb.New(t)
	}

	l.Logger.Infof("查询退款状态成功: refundSn=%s", in.RefundSn)
	return resp, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return int64(val)
		case int64:
			return val
		}
	}
	return 0
}

func getUint64(m map[string]interface{}, key string) uint64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return uint64(val)
		case int64:
			return uint64(val)
		case uint64:
			return val
		}
	}
	return 0
}
