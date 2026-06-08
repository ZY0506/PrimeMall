package userlogic

import (
	"context"
	"time"

	"github.com/ZY0506/PrimeMall/common/constants"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Logout 退出登录
func (l *LogoutLogic) Logout(in *user.LogoutReq) (*user.EmptyResp, error) {
	// 计算剩余时间
	accessExpire := l.svcCtx.Config.JWT.AccessTokenExpire
	refreshExpire := l.svcCtx.Config.JWT.RefreshTokenExpire

	accessKey := constants.ATOKEN_BLACKLIST_KEY + in.AccessJti
	refreshKey := constants.RTOKEN_BLACKLIST_KEY + in.RefreshJti

	// access token 黑名单
	if err := l.svcCtx.Client.SetNX(
		l.ctx,
		accessKey,
		"1",
		time.Duration(accessExpire)*time.Second,
	).Err(); err != nil {
		l.Logger.Error("拉黑 access_token 失败")
		return nil, err
	}

	// refresh token 黑名单
	if err := l.svcCtx.Client.SetNX(
		l.ctx,
		refreshKey,
		"1",
		time.Duration(refreshExpire)*time.Second,
	).Err(); err != nil {
		l.Logger.Error("拉黑 refresh_token 失败")
		return nil, err
	}
	l.Logger.Info("退出登录成功")

	return &user.EmptyResp{}, nil
}
