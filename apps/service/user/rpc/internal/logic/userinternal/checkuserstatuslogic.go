package userinternallogic

import (
	"context"
	"errors"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckUserStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckUserStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckUserStatusLogic {
	return &CheckUserStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CheckUserStatus 校验用户状态（是否被封禁/限制下单）
func (l *CheckUserStatusLogic) CheckUserStatus(in *user.CheckUserStatusReq) (*user.CheckUserStatusResp, error) {

	userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("用户不存在,userId=%d", in.UserId)
			return nil, errorx.NewBizError(response.ErrCodeUserNotFound, "用户不存在")
		}
		l.Logger.Errorf("获取用户信息失败,userId=%d,err=%v", in.UserId, err)
		return nil, err
	}
	if userInfo.Status == constants.USER_STATUS_NORMAL {
		return &user.CheckUserStatusResp{
			Allowed: true,
		}, nil
	}
	punishLog, err := l.svcCtx.PunishLogModel.FindLastLogByUserId(l.ctx, userInfo.Id)
	if err != nil {
		l.Logger.Errorf("获取用户处罚记录失败,userId=%d,err=%v", in.UserId, err)
		return nil, err
	}

	// 如果被封禁 ,任何操作都不允许
	if userInfo.Status == constants.USER_STATUS_BANNED {
		return &user.CheckUserStatusResp{Allowed: false, Reason: punishLog.Reason}, nil
	}

	// 如果被限制下单
	if userInfo.Status == constants.USER_STATUS_RESTRICTED {
		if in.CheckType == user.CheckType_CHECK_TYPE_LOGIN {
			return &user.CheckUserStatusResp{Allowed: true}, nil
		}
		return &user.CheckUserStatusResp{Allowed: false, Reason: punishLog.Reason}, nil
	}

	return &user.CheckUserStatusResp{
		Allowed: true,
	}, nil
}
