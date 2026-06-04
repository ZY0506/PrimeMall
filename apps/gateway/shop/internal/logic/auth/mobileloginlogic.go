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

type MobileLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewMobileLoginLogic 手机号验证码登录
func NewMobileLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MobileLoginLogic {
	return &MobileLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MobileLoginLogic) MobileLogin(req *types.MobileLoginReq) (resp *types.LoginResp, err error) {
	// 参数验证（手机号）
	if !utils.ValidatePhone(req.Phone) {
		return nil, response.NewBizError(response.ErrCodeInvalidParam, "手机号格式错误")
	}
	// 调用rpc
	res, err := l.svcCtx.UserRpc.MobileLogin(l.ctx, &user.MobileLoginReq{
		Phone: req.Phone,
		Code:  req.Code,
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
