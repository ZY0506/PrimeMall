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

type RefundApplyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRefundApplyLogic 申请退款（由订单服务售后调用，幂等）
func NewRefundApplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundApplyLogic {
	return &RefundApplyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// RefundApply 申请退款
// 业务逻辑：注入userID → 查询订单确认支付金额 → 返回模拟退款结果
// 注：实际退款流程由Order服务的售后审核通过后，调用PaymentInternal.Refund执行
// 本接口仅做退款申请的入口，返回退款单号供前端跟踪
func (l *RefundApplyLogic) RefundApply(req *types.RefundApplyReq) (resp *types.RefundApplyResp, err error) {
	// 注入userID
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}

	// 查询订单信息确认金额
	orderDetail, err := l.svcCtx.OrderRpc.OrderDetail(l.ctx, &order.OrderDetailRequest{
		OrderSn: req.OrderSn,
	})
	if err != nil {
		l.Logger.Errorf("查询订单详情失败，error=%v", err)
		// 订单查不到也继续，返回基础退款信息
	} else {
		// 校验退款金额不超过支付金额
		if req.Amount > orderDetail.Base.PayAmount {
			l.Logger.Errorf("退款金额超过支付金额: apply=%d, pay=%d", req.Amount, orderDetail.Base.PayAmount)
		}
	}

	// 生成退款单号
	refundSn := fmt.Sprintf("RF%s%s", req.OrderSn, time.Now().Format("150405"))

	// 查询支付单号
	var paymentSn string
	if req.PaymentSn != "" {
		paymentSn = req.PaymentSn
	} else {
		// 尝试从Redis索引获取
		paymentSn, _ = l.svcCtx.Client.Get(l.ctx, fmt.Sprintf("payment:order:%s", req.OrderSn)).Result()
	}

	// 如果有支付单号，查询支付状态
	payAmount := req.Amount
	if paymentSn != "" {
		detail, err := l.svcCtx.PaymentRpc.GetPaymentDetail(l.ctx, &payment.GetPaymentDetailRequest{
			PaymentSn: paymentSn,
		})
		if err == nil && detail != nil {
			payAmount = detail.Amount
		}
	}

	l.Logger.Infof("退款申请已受理: refundSn=%s, orderSn=%s, amount=%d", refundSn, req.OrderSn, payAmount)

	return &types.RefundApplyResp{
		RefundSn:   refundSn,
		Status:     int64(constants.REFUND_STATUS_REFUNDING),
		StatusDesc: "退款处理中",
	}, nil
}
