package userlogic

import (
	"context"
	"errors"
	"fmt"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePhoneLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePhoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePhoneLogic {
	return &UpdatePhoneLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdatePhone 修改手机号
func (l *UpdatePhoneLogic) UpdatePhone(in *user.UpdatePhoneReq) (*user.EmptyResp, error) {
	var ok bool
	var err error
	// 幂等判断
	idempotencyKey := fmt.Sprintf("%s%s%s", constants.IDEMPOTENCY_KEY, constants.USER_SERVICE, in.IdempotencyKey)
	ok, err = l.svcCtx.Client.SetNX(l.ctx, idempotencyKey, "1", constants.IDEMPOTENCY_EXIRE).Result()
	if err != nil {
		l.Logger.Errorf("幂等校验失败,error=%v", err)
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

	// 查询当前用户
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, userId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Infof("用户不存在,userId=%d", userId)
			return nil, errorx.NewBizError(response.ErrCodeUserNotFound, "用户不存在")
		}
		l.Logger.Errorf("获取用户信息失败,userId=%d,err=%v", userId, err)
		return nil, err
	}

	// 验证旧手机号验证码
	ok, err = l.svcCtx.CaptchaSvc.Verify(l.ctx, u.Phone, constants.SCENE_UPDATE_PHONE, in.OldPhoneCode)
	if err != nil {
		l.Logger.Errorf("验证码校验失败，error=%v", err)
		return nil, errorx.NewBizError(response.ErrCodeCaptchaWrong, "验证码校验失败")
	}
	if !ok {
		l.Logger.Infof("验证码错误,phone=%v", u.Phone)
		return nil, errorx.NewBizError(response.ErrCodeCaptchaWrong, "验证码错误")
	}

	// 验证新手机号验证码
	ok, err = l.svcCtx.CaptchaSvc.Verify(l.ctx, in.NewPhone, constants.SCENE_UPDATE_PHONE, in.NewPhoneCode)
	if err != nil {
		l.Logger.Errorf("验证码校验失败，error=%v", err)
		return nil, errorx.NewBizError(response.ErrCodeCaptchaWrong, "验证码校验失败")
	}
	if !ok {
		l.Logger.Infof("验证码错误,phone=%v", in.NewPhone)
		return nil, errorx.NewBizError(response.ErrCodeCaptchaWrong, "验证码错误")
	}

	// 检查新手机号是否已被其他用户注册
	existingUser, err := l.svcCtx.UserModel.FindOneByPhone(l.ctx, in.NewPhone)
	if err == nil && existingUser.Id != userId {
		// 如果查到的是已注销的账号，允许重新绑定
		if !existingUser.DeletedAt.Valid {
			l.Logger.Errorf("新手机号已被注册,phone=%s,ownerId=%d,requesterId=%d", in.NewPhone, existingUser.Id, userId)
			return nil, errorx.NewBizError(response.ErrCodePhoneRegistered, "手机号已注册")
		}
	} else if err != nil && !errors.Is(err, model.ErrNotFound) {
		l.Logger.Errorf("查询新手机号失败,error=%v", err)
		return nil, err
	}

	// 更新手机号
	u.Phone = in.NewPhone
	err = l.svcCtx.UserModel.Update(l.ctx, u)
	if err != nil {
		l.Logger.Errorf("更新用户信息失败,userId=%d,err=%v", userId, err)
		return nil, err
	}
	l.Logger.Infof("更新用户信息成功,userId=%d", userId)

	return &user.EmptyResp{}, nil
}
