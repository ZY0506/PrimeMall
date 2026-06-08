package adminlogic

import (
	"context"
	"errors"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCurrentAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCurrentAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrentAdminLogic {
	return &GetCurrentAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCurrentAdminLogic) GetCurrentAdmin(in *admin.Empty) (*admin.GetCurrentAdminResp, error) {
	adminId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取操作管理员ID失败,err=%v", err)
		return nil, err
	}

	adminUser, err := l.svcCtx.AdminModel.FindOne(l.ctx, adminId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errorx.NewBizError(response.ErrCodeAdminNotFound, "管理员不存在")
		}
		l.Logger.Errorf("查询管理员失败,error=%v", err)
		return nil, err
	}

	roleName := ""
	if adminUser.RoleId > 0 {
		role, err := l.svcCtx.RoleModel.FindOne(l.ctx, adminUser.RoleId)
		if err == nil {
			roleName = role.Name
		}
	}

	var permissionCodes []string
	if adminUser.RoleId > 0 {
		rps, err := l.svcCtx.RolePermissionModel.FindByRoleId(l.ctx, adminUser.RoleId)
		if err == nil {
			for _, rp := range rps {
				perm, err := l.svcCtx.PermissionModel.FindOne(l.ctx, rp.PermissionId)
				if err == nil {
					permissionCodes = append(permissionCodes, perm.Code)
				}
			}
		}
	}

	return &admin.GetCurrentAdminResp{
		Admin: &admin.AdminInfo{
			Id:        adminUser.Id,
			Username:  adminUser.Username,
			RealName:  adminUser.RealName,
			Avatar:    adminUser.Avatar,
			RoleId:    adminUser.RoleId,
			RoleName:  roleName,
			Status:    admin.AdminStatus(adminUser.Status),
			CreatedAt: timestamppb.New(adminUser.CreatedAt),
		},
		Permissions: permissionCodes,
	}, nil
}
