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

type ResetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewResetPasswordLogic 忘记密码重置（幂等）
func NewResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPasswordLogic {
	return &ResetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetPasswordLogic) ResetPassword(req *types.ResetPasswordReq) error {
	// 手机号检验
	if !utils.ValidatePhone(req.Phone) {
		return response.NewBizError(response.ErrCodeInvalidParam, "手机号格式错误")
	}
	// 密码校验
	if !utils.ValidatePassword(req.NewPassword) {
		return response.NewBizError(response.ErrCodeInvalidParam, "密码格式错误")
	}
	// 调用rpc
	_, err := l.svcCtx.UserRpc.ResetPassword(l.ctx, &user.ResetPasswordReq{
		Code:           req.Code,
		IdempotencyKey: req.IdempotencyKey,
		NewPassword:    req.NewPassword,
		Phone:          req.Phone,
	})
	if err != nil {
		l.Logger.Errorf("重置密码失败,err=%v", err)
		return err
	}
	return nil
}
