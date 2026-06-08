package adminlogic

import (
	"context"
	"errors"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAdminLogic {
	return &GetAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAdminLogic) GetAdmin(in *admin.GetAdminReq) (*admin.GetAdminResp, error) {
	adminUser, err := l.svcCtx.AdminModel.FindOne(l.ctx, in.Id)
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

	return &admin.GetAdminResp{
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
	}, nil
}
