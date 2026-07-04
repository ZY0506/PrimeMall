// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package admin

import (
	"context"
	"time"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserDetailLogic {
	return &UserDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserDetailLogic) UserDetail(req *types.IdReq) (resp *types.AdminUserDetailResp, err error) {
	rpcResp, err := l.svcCtx.AdminUserRpc.GetUserDetail(l.ctx, &admin.AdminGetUserDetailReq{
		UserId: req.Id,
	})
	if err != nil {
		l.Logger.Errorf("用户详情 RPC 调用失败, id=%d, error=%v", req.Id, err)
		return nil, err
	}
	lastLoginTime := ""
	if rpcResp.LastLoginTime != nil {
		lastLoginTime = rpcResp.LastLoginTime.AsTime().Format(time.RFC3339)
	}
	createdAt := ""
	if rpcResp.CreatedAt != nil {
		createdAt = rpcResp.CreatedAt.AsTime().Format(time.RFC3339)
	}
	return &types.AdminUserDetailResp{
		Id:            rpcResp.Id,
		Phone:         rpcResp.Phone,
		Nickname:      rpcResp.Nickname,
		Avatar:        rpcResp.Avatar,
		Gender:        rpcResp.Gender,
		Birthday:      rpcResp.Birthday,
		Status:        rpcResp.Status,
		StatusDesc:    rpcResp.StatusDesc,
		OrderCount:    rpcResp.OrderCount,
		TotalAmount:   rpcResp.TotalAmount,
		LastLoginTime: lastLoginTime,
		CreatedAt:     createdAt,
	}, nil
}
