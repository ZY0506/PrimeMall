// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package profile

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/jwt"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/ZY0506/PrimeMall/common/utils"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteAccountLogic 注销账号
func NewDeleteAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAccountLogic {
	return &DeleteAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAccountLogic) DeleteAccount(req *types.DeleteAccountReq) error {
	var err error
	// 参数校验
	if !utils.ValidatePassword(req.Password) {
		return response.NewBizError(response.ErrCodeInvalidParam, "密码格式错误")
	}
	// 注入userId
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return err
	}
	// 获取access_token的jti和expire
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
	// 调用rpc
	_, err = l.svcCtx.UserRpc.DeleteAccount(l.ctx, &user.DeleteAccountReq{
		Password:   req.Password,
		AccessJti:  accessJti,
		RefreshJti: c.JTI,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc出错，error=%v", err)
		return err
	}
	return nil
}
