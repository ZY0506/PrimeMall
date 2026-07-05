// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"time"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserListLogic {
	return &UserListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserListLogic) UserList(req *types.AdminUserListReq) (resp *types.AdminUserListResp, err error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 10
	}
	rpcResp, err := l.svcCtx.AdminUserRpc.ListUsers(l.ctx, &admin.AdminListUsersReq{
		Page:     req.Page,
		PageSize: req.Size,
		Keyword:  req.Keyword,
		Status:   req.Status,
	})
	if err != nil {
		l.Logger.Errorf("用户列表 RPC 调用失败: %v", err)
		return nil, err
	}
	list := make([]types.AdminUserItem, 0, len(rpcResp.List))
	for _, u := range rpcResp.List {
		createdAt := ""
		if u.CreatedAt != nil {
			createdAt = u.CreatedAt.AsTime().Format(time.RFC3339)
		}
		list = append(list, types.AdminUserItem{
			Id:         u.Id,
			Phone:      u.Phone,
			Nickname:   u.Nickname,
			Avatar:     u.Avatar,
			Status:     u.Status,
			StatusDesc: u.StatusDesc,
			CreatedAt:  createdAt,
		})
	}
	return &types.AdminUserListResp{Total: rpcResp.Total, List: list}, nil
}
