package core

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/types/payment"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePaymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreatePaymentLogic 创建支付单（获取第三方支付拉起参数，幂等）
func NewCreatePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaymentLogic {
	return &CreatePaymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreatePayment 创建支付单
// 业务逻辑：注入userID → 调用PaymentRpc创建支付 → 返回支付参数
func (l *CreatePaymentLogic) CreatePayment(req *types.CreatePaymentReq) (resp *types.CreatePaymentResp, err error) {
	// 注入userID
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}

	// 将渠道字符串转为proto支付类型
	var payType payment.PayType
	switch req.Channel {
	case "wechat":
		payType = payment.PayType_PAY_TYPE_WECHAT
	case "alipay":
		payType = payment.PayType_PAY_TYPE_ALIPAY
	default:
		payType = payment.PayType_PAY_TYPE_UNKNOWN
	}

	// 调用PaymentRpc
	createResp, err := l.svcCtx.PaymentRpc.CreatePayment(l.ctx, &payment.CreatePaymentRequest{
		OrderSn:        req.OrderSn,
		PayType:        payType,
		IdempotencyKey: req.IdempotencyKey,
		ClientIp:       "", // 可从请求上下文中获取
	})
	if err != nil {
		l.Logger.Errorf("调用RPC创建支付单失败，error=%v", err)
		return nil, err
	}

	return &types.CreatePaymentResp{
		PaymentSn: createResp.PaymentSn,
		PayData:   createResp.PayParams,
	}, nil
}
