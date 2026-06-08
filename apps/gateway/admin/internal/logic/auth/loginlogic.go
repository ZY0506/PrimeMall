// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 管理员登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.AdminLoginReq) (resp *types.AdminLoginResp, err error) {
	rpcResp, err := l.svcCtx.AdminRpc.AdminLogin(l.ctx, &admin.AdminLoginReq{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		l.Logger.Errorf("管理员登录失败,username=%s,err=%v", req.Username, err)
		return nil, err
	}

	return &types.AdminLoginResp{
		AccessToken:  rpcResp.Token,
		AccessExpire: rpcResp.ExpireTime,
	}, nil
}
