package adminlogic

import (
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAdminLogic {
	return &ListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListAdminLogic) ListAdmin(in *admin.ListAdminReq) (*admin.ListAdminResp, error) {
	admins, total, err := l.svcCtx.AdminModel.FindListByPage(l.ctx, &model.AdminListFilter{
		Page:     in.Page,
		PageSize: in.PageSize,
		Username: in.Username,
		Status:   int64(in.Status),
	})
	if err != nil {
		l.Logger.Errorf("查询管理员列表失败,error=%v", err)
		return nil, err
	}

	var list []*admin.AdminInfo
	for _, a := range admins {
		roleName := ""
		if a.RoleId > 0 {
			role, err := l.svcCtx.RoleModel.FindOne(l.ctx, a.RoleId)
			if err == nil {
				roleName = role.Name
			}
		}
		list = append(list, &admin.AdminInfo{
			Id:        a.Id,
			Username:  a.Username,
			RealName:  a.RealName,
			Avatar:    a.Avatar,
			RoleId:    a.RoleId,
			RoleName:  roleName,
			Status:    admin.AdminStatus(a.Status),
			CreatedAt: timestamppb.New(a.CreatedAt),
		})
	}

	return &admin.ListAdminResp{
		Total: total,
		List:  list,
	}, nil
}
