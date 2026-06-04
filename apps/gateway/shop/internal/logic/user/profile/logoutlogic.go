// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package profile

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/jwt"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 退出登录
func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout(req *types.RefreshTokenReq) error {
	// 获取access_token的jti
	accessJti, ok := l.ctx.Value("token_id").(string)
	if !ok {
		l.Logger.Error("获取access_token Jti失败")
		return response.NewBizError(response.ErrCodeMissingParam, "缺少必要参数")
	}

	// 解析refresh_token
	c, err := jwt.ParseToken(req.RefreshToken, []byte(l.svcCtx.Config.JwtAuth.RefreshSecret))
	if err != nil {
		l.Logger.Errorf("解析token失败,error=%v", err)
		return response.NewBizError(response.ErrCodeTokenInvalid, "Token 无效")
	}

	_, err = l.svcCtx.UserRpc.Logout(l.ctx, &user.LogoutReq{
		AccessJti:  accessJti,
		RefreshJti: c.JTI,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc出错，error=%v", err)
		return err
	}
	l.Logger.Info("退出登录成功")

	return nil
}
