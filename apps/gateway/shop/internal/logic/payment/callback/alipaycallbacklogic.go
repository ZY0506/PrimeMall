package callback

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

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
// 业务逻辑：模拟支付宝回调处理 → 返回固定成功
// 注：真实项目中，此处应验证签名、解析回调参数、调用PaymentRpc幂等更新支付状态
// 由于支付服务为模拟实现，直接返回成功
func (l *AlipayCallbackLogic) AlipayCallback() (resp *types.CallbackResp, err error) {
	l.Logger.Info("模拟支付宝回调处理成功")
	return &types.CallbackResp{
		Code:    "success",
		Message: "success",
	}, nil
}
