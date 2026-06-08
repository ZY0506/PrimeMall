package adminlogic

import (
	"context"
	"errors"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/logic/utils"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/model"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAdminLogic {
	return &UpdateAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateAdminLogic) UpdateAdmin(in *admin.UpdateAdminReq) (*admin.Empty, error) {
	startTime := time.Now()

	adminUser, err := l.svcCtx.AdminModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errorx.NewBizError(response.ErrCodeAdminNotFound, "管理员不存在")
		}
		l.Logger.Errorf("查询管理员失败,error=%v", err)
		return nil, err
	}

	if in.RoleId > 0 {
		_, err := l.svcCtx.RoleModel.FindOne(l.ctx, in.RoleId)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, errorx.NewBizError(response.ErrCodeRoleNotFound, "角色不存在")
			}
			l.Logger.Errorf("查询角色失败,error=%v", err)
			return nil, err
		}
	}

	adminUser.RealName = in.RealName
	adminUser.Avatar = in.Avatar
	adminUser.RoleId = in.RoleId
	if in.Status != admin.AdminStatus_ADMIN_STATUS_UNKNOWN {
		adminUser.Status = int64(in.Status)
	}

	if err := l.svcCtx.AdminModel.Update(l.ctx, adminUser); err != nil {
		l.Logger.Errorf("更新管理员失败,error=%v", err)
		return nil, err
	}

	// 记录操作日志
	durationMs := time.Since(startTime).Milliseconds()
	utils.RecordAdminLog(l.ctx, l.svcCtx, utils.LogModuleAdmin, "更新管理员", in, nil, durationMs)

	return &admin.Empty{}, nil
}
