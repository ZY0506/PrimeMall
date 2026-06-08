package adminlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPunishLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPunishLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPunishLogsLogic {
	return &ListPunishLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListPunishLogs 获取风控处罚日志列表
func (l *ListPunishLogsLogic) ListPunishLogs(in *user.ListPunishLogsReq) (*user.ListPunishLogsResp, error) {

	listResp, total, err := l.svcCtx.PunishLogModel.FindListByPage(l.ctx, &model.PunishLogListFilter{
		Page:       in.Page,
		PageSize:   in.Size,
		UserId:     in.UserId,
		ActionType: int64(in.ActionType),
		StartTime:  in.StartTime.AsTime(),
		EndTime:    in.EndTime.AsTime(),
		Operator:   in.Operator,
	})
	if err != nil {
		l.Logger.Errorf("获取风控处罚日志列表失败，error=%v", err)
		return nil, err
	}

	list := make([]*user.PunishLogItem, 0)
	for _, item := range listResp {
		list = append(list, &user.PunishLogItem{
			Id:         item.Id,
			UserId:     item.UserId,
			Phone:      item.Phone,
			ActionType: user.BanType(item.ActionType),
			Reason:     item.Reason,
			Operator:   item.BannedBy,
			StartTime:  timestamppb.New(item.StartTime.Time),
			EndTime:    timestamppb.New(item.EndTime.Time),
			CreatedAt:  timestamppb.New(item.CreatedAt),
		})
	}

	l.Logger.Info("获取风控处罚日志列表成功")

	return &user.ListPunishLogsResp{
		List:  list,
		Total: total,
	}, nil
}
