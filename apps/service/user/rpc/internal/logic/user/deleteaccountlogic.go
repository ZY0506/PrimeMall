package userlogic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/ZY0506/PrimeMall/pkg/pwd"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAccountLogic {
	return &DeleteAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteAccount 注销账号
func (l *DeleteAccountLogic) DeleteAccount(in *user.DeleteAccountReq) (*user.EmptyResp, error) {
	// 获取当前账号信息
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

	// 验证密码
	if !pwd.CompareHashAndPassword(u.Password, in.Password) {
		l.Logger.Errorf("密码错误,userId=%d", userId)
		return nil, errorx.NewBizError(response.ErrCodePasswordWrong, "密码错误")
	}

	// 验证账号是否被删除
	if u.DeletedAt.Valid {
		l.Logger.Errorf("账号已注销,userId=%d", userId)
		return nil, errorx.NewBizError(response.ErrCodeUserDeleted, "账号已注销")
	}

	// 拉黑token
	// 计算剩余时间
	accessExpire := l.svcCtx.Config.JWT.AccessTokenExpire
	refreshExpire := l.svcCtx.Config.JWT.RefreshTokenExpire

	accessKey := constants.ATOKEN_BLACKLIST_KEY + in.AccessJti
	refreshKey := constants.RTOKEN_BLACKLIST_KEY + in.RefreshJti

	// access token 黑名单
	if err := l.svcCtx.Client.SetNX(l.ctx, accessKey, "1", time.Duration(accessExpire)*time.Second).Err(); err != nil {
		l.Logger.Error("拉黑 access_token 失败")
		return nil, err
	}

	// refresh token 黑名单
	if err := l.svcCtx.Client.SetNX(l.ctx, refreshKey, "1", time.Duration(refreshExpire)*time.Second).Err(); err != nil {
		l.Logger.Error("拉黑 refresh_token 失败")
		return nil, err
	}

	// 删除账号
	u.DeletedAt = sql.NullTime{Time: time.Now(), Valid: true}
	err = l.svcCtx.UserModel.Update(l.ctx, u)
	if err != nil {
		l.Logger.Errorf("账号注销失败,userId=%d,err=%v", userId, err)
		return nil, err
	}
	l.Logger.Infof("账号注销成功,userId=%d", userId)

	return &user.EmptyResp{}, nil
}
