package core

import (
	"context"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/types/payment"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type StatusByOrderSnLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewStatusByOrderSnLogic 根据订单号查询支付状态
func NewStatusByOrderSnLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StatusByOrderSnLogic {
	return &StatusByOrderSnLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// StatusByOrderSn 根据订单号查询支付状态
// 业务逻辑：注入userID → 从Redis获取订单对应的支付流水号 → 查询支付详情 → 返回支付状态
// 若Redis中无支付流水号，则从订单状态推断支付状态
func (l *StatusByOrderSnLogic) StatusByOrderSn(req *types.OrderSnPathReq) (resp *types.PaymentStatusResp, err error) {
	// 注入userID
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}

	// 尝试从Redis查询订单号对应的支付流水号
	orderIndexKey := fmt.Sprintf("payment:order:%s", req.OrderSn)
	paymentSn, redisErr := l.svcCtx.Client.Get(l.ctx, orderIndexKey).Result()

	// 如果Redis中有支付流水号，优先通过PaymentRpc查询
	if redisErr == nil && paymentSn != "" {
		detail, err := l.svcCtx.PaymentRpc.GetPaymentDetail(l.ctx, &payment.GetPaymentDetailRequest{
			PaymentSn: paymentSn,
		})
		if err == nil {
			var payTime string
			if detail.PayTime != nil {
				payTime = detail.PayTime.AsTime().Format(time.RFC3339)
			}
			return &types.PaymentStatusResp{
				PaymentSn:      detail.PaymentSn,
				OrderSn:        req.OrderSn,
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
	}

	// 降级：通过订单服务查询订单状态来推断支付状态
	orderDetail, err := l.svcCtx.OrderRpc.OrderDetail(l.ctx, &order.OrderDetailRequest{
		OrderSn: req.OrderSn,
	})
	if err != nil {
		l.Logger.Errorf("查询订单详情失败，error=%v", err)
		return nil, err
	}

	payStatus := l.inferPaymentStatus(int64(orderDetail.Base.Status))
	return &types.PaymentStatusResp{
		PaymentSn:  paymentSn,
		OrderSn:    req.OrderSn,
		Channel:    "",
		Status:     payStatus,
		StatusDesc: l.paymentStatusDesc(payStatus),
		Amount:     orderDetail.Base.PayAmount,
	}, nil
}

// inferPaymentStatus 从订单状态推断支付状态
func (l *StatusByOrderSnLogic) inferPaymentStatus(orderStatus int64) int64 {
	switch orderStatus {
	case constants.ORDER_STATUS_PENDING_PAY:
		return 0 // 待支付
	case constants.ORDER_STATUS_PAID, constants.ORDER_STATUS_SHIPPED, constants.ORDER_STATUS_COMPLETED:
		return 1 // 支付成功
	case constants.ORDER_STATUS_CANCELED:
		return 2 // 支付失败/已取消
	default:
		return 0
	}
}

func (l *StatusByOrderSnLogic) payTypeToChannel(payType int64) string {
	switch payType {
	case 1:
		return "wechat"
	case 2:
		return "alipay"
	default:
		return ""
	}
}

func (l *StatusByOrderSnLogic) paymentStatusDesc(status int64) string {
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
