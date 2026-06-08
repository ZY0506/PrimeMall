package admin

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/zeromicro/go-zero/core/logx"
)

type OperateLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOperateLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperateLogsLogic {
	return &OperateLogsLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *OperateLogsLogic) OperateLogs(req *types.Pagination) (resp *types.AdminOperateLogResp, err error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	rpcResp, err := l.svcCtx.AdminRpc.ListAdminLog(l.ctx, &admin.ListAdminLogReq{
		Page: req.Page, PageSize: req.Size,
	})
	if err != nil {
		l.Logger.Errorf("查询操作日志失败: %v", err)
		return nil, err
	}
	list := make([]types.AdminOperateLogItem, 0, len(rpcResp.List))
	for _, lg := range rpcResp.List {
		createdAt := ""
		if lg.CreatedAt != nil {
			createdAt = lg.CreatedAt.AsTime().Format("2006-01-02 15:04:05")
		}
		list = append(list, types.AdminOperateLogItem{
			Id: lg.Id, AdminId: lg.AdminId, Module: lg.Module.String(),
			Action: lg.Action, Content: lg.RequestParams, Ip: lg.Ip, CreatedAt: createdAt,
		})
	}
	return &types.AdminOperateLogResp{Total: rpcResp.Total, List: list}, nil
}
