package callback

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WechatCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewWechatCallbackLogic 微信支付异步回调（服务端自行幂等）
func NewWechatCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WechatCallbackLogic {
	return &WechatCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// WechatCallback 处理微信支付异步回调
// 业务逻辑：模拟微信回调处理 → 返回固定成功
// 注：真实项目中，此处应验证签名、解析回调XML、调用PaymentRpc幂等更新支付状态
// 由于支付服务为模拟实现，直接返回成功
func (l *WechatCallbackLogic) WechatCallback() (resp *types.CallbackResp, err error) {
	l.Logger.Info("模拟微信支付回调处理成功")
	return &types.CallbackResp{
		Code:    "SUCCESS",
		Message: "OK",
	}, nil
}
