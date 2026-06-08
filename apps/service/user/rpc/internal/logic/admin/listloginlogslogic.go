package adminlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/constants"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLoginLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLoginLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLoginLogsLogic {
	return &ListLoginLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListLoginLogs 获取登录日志列表
func (l *ListLoginLogsLogic) ListLoginLogs(in *user.ListLoginLogsReq) (*user.ListLoginLogsResp, error) {

	loginType := ""
	if in.LoginType == "password" {
		loginType = constants.LOGIN_BY_PASSWORD
	} else {
		loginType = constants.LOGIN_BY_CAPTCHA
	}

	listResp, total, err := l.svcCtx.LoginLogModel.FindListByPage(l.ctx, &model.LoginLogListFilter{
		Page:      in.Page,
		PageSize:  in.Size,
		UserId:    in.UserId,
		LoginType: loginType,
		Status:    in.Status,
		StartTime: in.StartTime.AsTime(),
		EndTime:   in.EndTime.AsTime(),
	})
	if err != nil {
		l.Logger.Errorf("获取登录日志列表失败，error=%v", err)
		return nil, err
	}

	list := make([]*user.LoginLogItem, 0)
	for _, loginLog := range listResp {
		list = append(list, &user.LoginLogItem{
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

	l.Logger.Info("获取登录日志列表成功")

	return &user.ListLoginLogsResp{
		List:  list,
		Total: total,
	}, nil
}
