// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/jwt"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRefreshTokenLogic token刷新
func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenReq) (resp *types.LoginResp, err error) {
	// 1、token校验
	c, err := jwt.ParseToken(req.RefreshToken, []byte(l.svcCtx.Config.JwtAuth.RefreshSecret))
	if err != nil {
		l.Logger.Errorf("解析token失败,error=%v", err)
		return nil, response.NewBizError(response.ErrCodeTokenInvalid, "Token 无效")
	}

	// 2、将user_id写入ctx中
	l.ctx, err = ctxdata.PutUserIdToCtx(l.ctx, c.UserId)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}

	// 3、调用rpc
	res, err := l.svcCtx.UserRpc.RefreshToken(l.ctx, &user.RefreshTokenReq{
		RefreshJti: c.JTI,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc出错，error=%v", err)
		return nil, err
	}
	return &types.LoginResp{
		AccessExpire:  res.AccessExpire.AsTime().Unix(),
		AccessToken:   res.AccessToken,
		RefreshExpire: res.RefreshExpire.AsTime().Unix(),
		RefreshToken:  res.RefreshToken,
	}, nil
}
