package adminlogic

import (
	"context"
	"errors"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserDetailLogic {
	return &GetUserDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUserDetail 获取用户详情（含扩展信息、最近登录、处罚记录）
func (l *GetUserDetailLogic) GetUserDetail(in *user.GetUserDetailReq) (*user.GetUserDetailResp, error) {

	userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("用户不存在,userId=%d", in.UserId)
			return nil, errorx.NewBizError(response.ErrCodeUserNotFound, "用户不存在")
		}
		l.Logger.Errorf("获取用户信息失败,userId=%d,err=%v", in.UserId, err)
		return nil, err
	}
	if userInfo.DeletedAt.Valid {
		l.Logger.Errorf("账号已注销,userId=%d", in.UserId)
		return nil, errorx.NewBizError(response.ErrCodeUserDeleted, "账号已注销")
	}

	// 查询登录日志
	loginLogsResp, err := l.svcCtx.LoginLogModel.FindLimitListByUserId(l.ctx, in.UserId, 5)
	if err != nil {
		l.Logger.Errorf("获取登录日志失败,userId=%d,err=%v", in.UserId, err)
		return nil, err
	}
	loginLogs := make([]*user.LoginLogItem, 0)
	for _, loginLog := range loginLogsResp {
		loginLogs = append(loginLogs, &user.LoginLogItem{
			Id:         loginLog.Id,
			UserId:     loginLog.UserId,
			LoginType:  loginLog.LoginType,
			LoginIp:    loginLog.LoginIp,
			UserAgent:  loginLog.UserAgent,
			Status:     loginLog.Status,
			FailReason: loginLog.FailReason,
			CreatedAt:  timestamppb.New(loginLog.CreatedAt),
		})
	}

	// 查询处罚记录
	punishLogsResp, err := l.svcCtx.PunishLogModel.FindListByUserId(l.ctx, in.UserId)
	if err != nil {
		l.Logger.Errorf("获取处罚记录失败,userId=%d,err=%v", in.UserId, err)
		return nil, err
	}
	punishLogs := make([]*user.PunishLogItem, 0)
	for _, punishLog := range punishLogsResp {
		punishLogs = append(punishLogs, &user.PunishLogItem{
			Id:         punishLog.Id,
			UserId:     punishLog.UserId,
			Phone:      punishLog.Phone,
			ActionType: user.BanType(punishLog.ActionType),
			Reason:     punishLog.Reason,
			Operator:   punishLog.BannedBy,
			StartTime:  timestamppb.New(punishLog.StartTime.Time),
			EndTime:    timestamppb.New(punishLog.EndTime.Time),
		})
	}

	l.Logger.Info("获取用户详情成功")

	return &user.GetUserDetailResp{
		UserInfo: &user.UserListItem{
			Id:            userInfo.Id,
			Phone:         userInfo.Phone,
			Nickname:      userInfo.Nickname,
			Avatar:        userInfo.Avatar,
			Gender:        userInfo.Gender,
			Birthday:      timestamppb.New(userInfo.Birthday.Time),
			Status:        user.UserStatus(userInfo.Status),
			CreatedAt:     timestamppb.New(userInfo.CreatedAt),
			LastLoginTime: timestamppb.New(userInfo.LastLoginTime.Time),
			LastLoginIp:   userInfo.LastLoginIp,
		},
		RecentLoginLogs: loginLogs,
		PunishLogs:      punishLogs,
	}, nil
}
