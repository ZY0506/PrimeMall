package userlogic

import (
	"context"
	"fmt"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
)

type CaptchaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaptchaLogic {
	return &CaptchaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Captcha 获取短信验证码
func (l *CaptchaLogic) Captcha(in *user.CaptchaReq) (*user.EmptyResp, error) {
	var ok bool
	var err error
	// 幂等校验
	idempotencyKey := fmt.Sprintf("%s%s", constants.IDEMPOTENCY_KEY+constants.USER_SERVICE, in.IdempotencyKey)
	ok, err = l.svcCtx.Client.SetNX(l.ctx, idempotencyKey, "1", constants.IDEMPOTENCY_EXIRE).Result()
	if err != nil {
		l.Logger.Errorf("[Captcha] 幂等性校验失败,error=%v", err)
		return nil, err
	}
	if !ok {
		return nil, errorx.NewBizError(response.ErrCodeTooFrequent, "请勿重复操作")
	}
	// TODO: 若后续失败，删除键（需确保删除操作的原子性，可配合 Lua 脚本）
	defer func() {
		if err != nil {
			l.svcCtx.Client.Del(l.ctx, idempotencyKey)
		}
	}()

	err = l.svcCtx.CaptchaSvc.Send(l.ctx, in.Phone, in.Scene, func(phone, code string) error {
		return nil
	})
	l.Logger.Infof("[Captcha] 发送验证码成功", logx.Field("phone", in.Phone))

	// TODO: 添加发送回调、风控
	return &user.EmptyResp{}, err
}
