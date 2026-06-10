package userlogic

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserInfoLogic {
	return &UpdateUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateUserInfo 更新个人资料
func (l *UpdateUserInfoLogic) UpdateUserInfo(in *user.UpdateUserInfoReq) (*user.EmptyResp, error) {
	// 获取id
	uid, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 验证id
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, uid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Infof("用户不存在,userId=%d", uid)
			return nil, errorx.NewBizError(response.ErrCodeUserNotFound, "用户不存在")
		}
		l.Logger.Errorf("获取用户信息失败,userId=%d,err=%v", uid, err)
		return nil, err
	}
	if u.DeletedAt.Valid {
		l.Logger.Infof("用户已注销,userId=%d", uid)
		return nil, errorx.NewBizError(response.ErrCodeUserDeleted, "用户已注销")
	}

	if in.Nickname != "" {
		if len([]rune(in.Nickname)) > 30 {
			return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "昵称最大30个字符")
		}
		u.Nickname = in.Nickname
	}
	if in.Avatar != "" {
		u.Avatar = in.Avatar
	}
	if in.Gender != 0 {
		if in.Gender != 1 && in.Gender != 2 {
			return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "性别值非法，合法值: 0未知/1男/2女")
		}
		u.Gender = in.Gender
	}
	if in.Birthday != nil {
		u.Birthday = sql.NullTime{Time: in.Birthday.AsTime(), Valid: true}
	}

	err = l.svcCtx.UserModel.Update(l.ctx, u)
	if err != nil {
		l.Logger.Errorf("更新用户信息失败,userId=%d,err=%v", uid, err)
		return nil, err
	}
	l.Logger.Infof("更新用户信息成功,userId=%d", uid)

	return &user.EmptyResp{}, nil
}
