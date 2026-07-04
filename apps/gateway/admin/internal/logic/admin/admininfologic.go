package admin

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
)

type AdminInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminInfoLogic {
	return &AdminInfoLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *AdminInfoLogic) AdminInfo() (resp *types.AdminInfoResp, err error) {
	rpcResp, err := l.svcCtx.AdminRpc.GetCurrentAdmin(l.ctx, &admin.Empty{})
	if err != nil {
		return nil, err
	}
	if rpcResp.Admin != nil {
		return &types.AdminInfoResp{
			Id: rpcResp.Admin.Id, Username: rpcResp.Admin.Username,
			Nickname: rpcResp.Admin.RealName, Avatar: rpcResp.Admin.Avatar, Role: rpcResp.Admin.RoleName,
		}, nil
	}
	return nil, errorx.NewBizError(response.ErrCodeAdminNotFound, "管理员不存在")
}
