package adminlogic

import (
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRoleLogic {
	return &ListRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListRoleLogic) ListRole(in *admin.Empty) (*admin.ListRoleResp, error) {
	roles, err := l.svcCtx.RoleModel.FindAll(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询角色列表失败,error=%v", err)
		return nil, err
	}

	var list []*admin.RoleInfo
	for _, r := range roles {
		list = append(list, &admin.RoleInfo{
			Id:        r.Id,
			Name:      r.Name,
			Code:      r.Code,
			Remark:    r.Remark,
			CreatedAt: timestamppb.New(r.CreatedAt),
		})
	}

	return &admin.ListRoleResp{List: list}, nil
}
