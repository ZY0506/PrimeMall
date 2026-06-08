package adminlogic

import (
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAdminLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAdminLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAdminLogLogic {
	return &ListAdminLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListAdminLogLogic) ListAdminLog(in *admin.ListAdminLogReq) (*admin.ListAdminLogResp, error) {
	adminId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取操作管理员ID失败,err=%v", err)
		return nil, err
	}

	logs, total, err := l.svcCtx.AdminLogModel.FindListByPage(l.ctx, &model.AdminLogListFilter{
		Page:     in.Page,
		PageSize: in.PageSize,
		AdminId:  adminId,
		Module:   in.Module.String(),
	})
	if err != nil {
		l.Logger.Errorf("查询日志列表失败,error=%v", err)
		return nil, err
	}

	var list []*admin.AdminLogInfo
	for _, log := range logs {
		list = append(list, &admin.AdminLogInfo{
			Id:             log.Id,
			AdminId:        log.AdminId,
			Username:       log.Username,
			Module:         admin.LogModule(admin.LogModule_value[log.Module]),
			Action:         log.Action,
			RequestMethod:  log.RequestMethod,
			RequestUrl:     log.RequestUrl,
			RequestParams:  log.RequestParams.String,
			ResponseResult: log.ResponseResult.String,
			Ip:             log.Ip,
			DurationMs:     log.DurationMs,
			CreatedAt:      timestamppb.New(log.CreatedAt),
		})
	}

	return &admin.ListAdminLogResp{
		Total: total,
		List:  list,
	}, nil
}
