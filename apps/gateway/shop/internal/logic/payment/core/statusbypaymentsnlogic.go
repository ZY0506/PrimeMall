package core

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/types/payment"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"time"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type StatusByPaymentSnLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewStatusByPaymentSnLogic 根据支付流水号查询支付状态
func NewStatusByPaymentSnLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StatusByPaymentSnLogic {
	return &StatusByPaymentSnLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// StatusByPaymentSn 根据支付流水号查询支付状态
// 业务逻辑：注入userID → 调用PaymentRpc查询支付详情 → 组装响应
func (l *StatusByPaymentSnLogic) StatusByPaymentSn(req *types.PaymentSnPathReq) (resp *types.PaymentStatusResp, err error) {
	// 注入userID
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}

	// 调用PaymentRpc查询支付详情
	detail, err := l.svcCtx.PaymentRpc.GetPaymentDetail(l.ctx, &payment.GetPaymentDetailRequest{
		PaymentSn: req.PaymentSn,
	})
	if err != nil {
		l.Logger.Errorf("调用RPC查询支付详情失败，error=%v", err)
		return nil, err
	}

	var payTime string
	if detail.PayTime != nil {
		payTime = detail.PayTime.AsTime().Format(time.RFC3339)
	}

	return &types.PaymentStatusResp{
		PaymentSn:      detail.PaymentSn,
		OrderSn:        detail.OrderSn,
		Channel:        l.payTypeToChannel(int64(detail.PayType)),
		ChannelOrderSn: detail.ChannelOrderSn,
		TransactionId:  detail.TransactionId,
		Status:         int64(detail.Status),
		StatusDesc:     l.paymentStatusDesc(int64(detail.Status)),
		Amount:         detail.Amount,
		PayTime:        payTime,
		ErrorMsg:       detail.ErrorMsg,
	}, nil
}

func (l *StatusByPaymentSnLogic) payTypeToChannel(payType int64) string {
	switch payType {
	case 1:
		return "wechat"
	case 2:
		return "alipay"
	default:
		return ""
	}
}

func (l *StatusByPaymentSnLogic) paymentStatusDesc(status int64) string {
	switch status {
	case 0:
		return "待支付"
	case 1:
		return "支付成功"
	case 2:
		return "支付失败"
	case 3:
		return "退款中"
	case 4:
		return "已退款"
	default:
		return "未知"
	}
}
