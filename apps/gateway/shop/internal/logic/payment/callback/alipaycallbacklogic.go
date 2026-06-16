package callback

import (
	"context"
	"encoding/json"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/types/payment"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type AlipayCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAlipayCallbackLogic 支付宝异步回调（服务端自行幂等）
func NewAlipayCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AlipayCallbackLogic {
	return &AlipayCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AlipayCallback 处理支付宝异步回调
// 业务逻辑：解析回调JSON → 调用 PaymentInternalRpc 幂等更新支付状态 → 返回处理结果
func (l *AlipayCallbackLogic) AlipayCallback(rawBody []byte) (resp *types.CallbackResp, err error) {
	// 1. 解析回调数据
	var req struct {
		PaymentSn string `json:"payment_sn"`
		OrderSn   string `json:"order_sn"`
		Signature string `json:"signature"`
	}
	if err := json.Unmarshal(rawBody, &req); err != nil {
		l.Logger.Errorf("解析支付宝回调数据失败: %v, raw=%s", err, string(rawBody))
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "回调数据格式错误")
	}

	if req.PaymentSn == "" || req.OrderSn == "" {
		l.Logger.Errorf("支付宝回调数据缺少必要参数: payment_sn=%s, order_sn=%s", req.PaymentSn, req.OrderSn)
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "回调数据缺少必要参数")
	}

	// 2. 调用 PaymentInternalRpc 处理支付回调
	result, err := l.svcCtx.PaymentInternalRpc.HandlePaymentCallback(l.ctx, &payment.PaymentCallbackRequest{
		PaymentSn: req.PaymentSn,
		OrderSn:   req.OrderSn,
		Channel:   "alipay",
		RawData:   string(rawBody),
		Signature: req.Signature,
	})
	if err != nil {
		l.Logger.Errorf("调用支付回调处理RPC失败: payment_sn=%s, err=%v", req.PaymentSn, err)
		return nil, err
	}

	l.Logger.Infof("支付宝回调处理完成: payment_sn=%s, order_sn=%s, success=%v, msg=%s",
		req.PaymentSn, req.OrderSn, result.Success, result.Message)

	if !result.Success {
		return &types.CallbackResp{
			Code:    "success",
			Message: result.Message,
		}, nil
	}

	return &types.CallbackResp{
		Code:    "success",
		Message: result.Message,
	}, nil
}
