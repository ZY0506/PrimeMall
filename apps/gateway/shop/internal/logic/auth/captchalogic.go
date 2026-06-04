// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/ZY0506/PrimeMall/common/utils"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCaptchaLogic 获取短信验证码（幂等）
func NewCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaptchaLogic {
	return &CaptchaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Captcha 获取短信验证码（幂等）
func (l *CaptchaLogic) Captcha(req *types.CaptchaReq) error {
	// 参数验证（手机号）
	if !utils.ValidatePhone(req.Phone) {
		return response.NewBizError(response.ErrCodeInvalidParam, "手机号格式错误")
	}
	// 调用rpc
	_, err := l.svcCtx.UserRpc.Captcha(l.ctx, &user.CaptchaReq{
		Phone:          req.Phone,
		Scene:          req.Scene,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		l.Logger.Errorf("调用RPC失败，error=%v", err)
		return err
	}
	return nil
}
