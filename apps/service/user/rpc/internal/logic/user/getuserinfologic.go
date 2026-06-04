package userlogic

import (
	"context"
	"errors"
	"fmt"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUserInfo 获取个人资料
func (l *GetUserInfoLogic) GetUserInfo(in *user.EmptyReq) (*user.UserInfoResp, error) {
	// 获取用户ID
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	res, err := l.svcCtx.UserModel.FindOne(l.ctx, userId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Infof("用户不存在,userId=%d", userId)
			return nil, errorx.NewBizError(response.ErrCodeUserNotFound, "用户不存在")
		}
		l.Logger.Errorf("获取用户信息失败,userId=%d,err=%v", userId, err)
		return nil, err
	}
	statusDesc := ""
	if res.Status != constants.USER_STATUS_NORMAL {
		userPunish, err := l.svcCtx.PunishLogModel.FindLastLogByUserId(l.ctx, res.Id)
		if err != nil {
			l.Logger.Errorf("获取用户封禁信息失败,userId=%d,err=%v", userId, err)
			return nil, err
		}
		if res.Status == constants.USER_STATUS_BANNED {
			statusDesc = fmt.Sprintf("因违反社区规定，封禁时间：%s-%s",
				userPunish.StartTime.Time.Format("2006-01-02 15:04:05"),
				userPunish.EndTime.Time.Format("2006-01-02 15:04:05"),
			)
		} else {
			statusDesc = fmt.Sprintf("因违反社区规定，限制下单：%s-%s",
				userPunish.StartTime.Time.Format("2006-01-02 15:04:05"),
				userPunish.EndTime.Time.Format("2006-01-02 15:04:05"),
			)
		}
	}

	l.Logger.Infof("获取用户信息成功,userId=%d", userId)

	return &user.UserInfoResp{
		Id:         res.Id,
		Phone:      res.Phone,
		Nickname:   res.Nickname,
		Avatar:     res.Avatar,
		Gender:     res.Gender,
		Birthday:   timestamppb.New(res.Birthday.Time),
		Status:     user.UserStatus(res.Status),
		StatusDesc: statusDesc,
	}, nil
}
