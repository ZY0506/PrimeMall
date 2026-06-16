// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

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
	// 调用rpc
	res, err := l.svcCtx.UserRpc.Login(l.ctx, &user.LoginReq{
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc登录失败,err=%v", err)
		return nil, err
	}
	// 返回结果
	return &types.LoginResp{
		AccessExpire:  res.AccessExpire.AsTime().Unix(),
		AccessToken:   res.AccessToken,
		RefreshExpire: res.RefreshExpire.AsTime().Unix(),
		RefreshToken:  res.RefreshToken,
	}, nil
}
