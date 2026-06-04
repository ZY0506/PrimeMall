package userlogic

import (
	"context"
	"errors"
	"fmt"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/ZY0506/PrimeMall/pkg/pwd"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPasswordLogic {
	return &ResetPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ResetPassword 忘记密码重置
func (l *ResetPasswordLogic) ResetPassword(in *user.ResetPasswordReq) (*user.EmptyResp, error) {
	var ok bool
	var err error
	// 幂等判断
	idempotencyKey := fmt.Sprintf("%s%s", constants.IDEMPOTENCY_KEY+constants.USER_SERVICE, in.IdempotencyKey)
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

	// 检查用户是否存在
	userInfo, err := l.svcCtx.UserModel.FindOneByPhone(l.ctx, in.Phone)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Infof("用户不存在,phone=%s", in.Phone)
			return nil, errorx.NewBizError(response.ErrCodeUserNotFound, "用户不存在")
		}
		l.Logger.Errorf("数据库查询失败，error=%v", err)
		return nil, err
	}

	// 验证账号状态
	if userInfo.Status == constants.USER_STATUS_BANNED {
		l.Logger.Errorf("用户被封禁，userId=%v", userInfo.Id)
		return nil, errorx.NewBizError(response.ErrCodeUserDisabled, "用户被封禁")
	}
	if userInfo.DeletedAt.Valid {
		l.Logger.Errorf("用户已注销，userId=%v", userInfo.Id)
		return nil, errorx.NewBizError(response.ErrCodeUserDeleted, "用户已注销")
	}

	// 校验验证码
	ok, err = l.svcCtx.CaptchaSvc.Verify(l.ctx, in.Phone, constants.SCENE_RESET_PWD, in.Code)
	if err != nil {
		l.Logger.Errorf("验证码校验失败，error=%v", err)
		return nil, err
	}
	if !ok {
		l.Logger.Info("验证码错误", in.Phone)
		return nil, errorx.NewBizError(response.ErrCodeCaptchaWrong, "验证码错误")
	}

	// 检查密码是否一致
	if pwd.CompareHashAndPassword(userInfo.Password, in.NewPassword) {
		l.Logger.Info("新密码与旧密码一致", in.Phone)
		return nil, errorx.NewBizError(response.ErrCodeSameAsOldPassword, "新密码与旧密码一致")
	}

	// 更新密码
	newPassword, err := pwd.GenerateFromPassword(in.NewPassword)
	if err != nil {
		l.Logger.Errorf("密码加密失败,error=%v", err)
		return nil, err
	}
	userInfo.Password = newPassword
	err = l.svcCtx.UserModel.Update(l.ctx, userInfo)
	if err != nil {
		l.Logger.Errorf("数据库更新失败，error=%v", err)
		return nil, err
	}
	return &user.EmptyResp{}, nil
}
