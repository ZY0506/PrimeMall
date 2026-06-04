package adminuserlogic

import (
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	userrpc "github.com/ZY0506/PrimeMall/apps/service/user/rpc/client/admin"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUsersLogic) ListUsers(in *admin.AdminListUsersReq) (*admin.AdminListUsersResp, error) {
	resp, err := l.svcCtx.UserRpc.ListUsers(l.ctx, &userrpc.ListUsersReq{
		Page:     in.Page,
		Size:     in.PageSize,
		Phone:    in.Keyword,
		Nickname: in.Keyword,
	})
	if err != nil {
		l.Logger.Errorf("查询用户列表失败,error=%v", err)
		return nil, err
	}

	var list []*admin.AdminUserItem
	for _, item := range resp.List {
		statusDesc := getUserStatusDesc(int64(item.Status))
		list = append(list, &admin.AdminUserItem{
			Id:         item.Id,
			Phone:      item.Phone,
			Nickname:   item.Nickname,
			Avatar:     item.Avatar,
			Status:     int64(item.Status),
			StatusDesc: statusDesc,
			CreatedAt:  timestamppb.New(item.CreatedAt.AsTime()),
		})
	}

	return &admin.AdminListUsersResp{
		Total: resp.Total,
		List:  list,
	}, nil
}

func getUserStatusDesc(status int64) string {
	switch status {
	case 1:
		return "正常"
	case 2:
		return "已禁用"
	case 3:
		return "已注销"
	default:
		return "未知"
	}
}
