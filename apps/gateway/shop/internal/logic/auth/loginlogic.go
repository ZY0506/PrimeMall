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

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewLoginLogic 手机号密码登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	// 1、校验手机号是否合法
	if !utils.ValidatePhone(req.Phone) {
		return nil, response.NewBizError(response.ErrCodeInvalidParam, "手机号格式错误")
	}
	// 2、校验密码是否合法
	if !utils.ValidatePassword(req.Password) {
		return nil, response.NewBizError(response.ErrCodeInvalidParam, "密码格式错误")
	}
	// 3、调用rpc
	res, err := l.svcCtx.UserRpc.Login(l.ctx, &user.LoginReq{
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc登录失败,err=%v", err)
		return nil, err
	}
	// 4、返回结果
	return &types.LoginResp{
		AccessExpire:  res.AccessExpire.AsTime().Unix(),
		AccessToken:   res.AccessToken,
		RefreshExpire: res.RefreshExpire.AsTime().Unix(),
		RefreshToken:  res.RefreshToken,
	}, nil
}
