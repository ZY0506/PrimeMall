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

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRegisterLogic 用户注册（幂等）
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.LoginResp, err error) {
	// 1、检验手机号合法性
	if !utils.ValidatePhone(req.Phone) {
		return nil, response.NewBizError(response.ErrCodeInvalidParam, "手机号格式错误")
	}
	// 2、检验密码合法性
	if !utils.ValidatePassword(req.Password) {
		return nil, response.NewBizError(response.ErrCodeInvalidParam, "密码格式错误")
	}
	// 3、调用rpc
	res, err := l.svcCtx.UserRpc.Register(l.ctx, &user.RegisterReq{
		Phone:           req.Phone,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
		Code:            req.Code,
		Nickname:        req.Nickname,
		IdempotencyKey:  req.IdempotencyKey,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc注册用户失败，err=%v", err)
		return nil, err
	}
	return &types.LoginResp{
		AccessExpire:  res.AccessExpire.AsTime().Unix(),
		AccessToken:   res.AccessToken,
		RefreshExpire: res.RefreshExpire.AsTime().Unix(),
		RefreshToken:  res.RefreshToken,
	}, nil
}
