package adminlogic

import (
	"context"
	"errors"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/logic/utils"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/ZY0506/PrimeMall/pkg/pwd"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAdminLogic {
	return &CreateAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateAdminLogic) CreateAdmin(in *admin.CreateAdminReq) (*admin.Empty, error) {
	startTime := time.Now()

	// 从上下文中获取当前操作的管理员
	operatorId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取操作管理员ID失败,err=%v", err)
		return nil, err
	}

	_, err = l.svcCtx.AdminModel.FindOneByUsername(l.ctx, in.Username)
	if err == nil {
		return nil, errorx.NewBizError(response.ErrCodeAdminUsernameExists, "管理员用户名已存在")
	}
	if !errors.Is(err, model.ErrNotFound) {
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

	hashedPassword, err := pwd.GenerateFromPassword(in.Password)
	if err != nil {
		l.Logger.Errorf("生成密码失败,error=%v", err)
		return nil, err
	}

	id, err := l.svcCtx.IDGenerator.NextID()
	if err != nil {
		l.Logger.Errorf("生成ID失败,error=%v", err)
		return nil, err
	}

	_, err = l.svcCtx.AdminModel.Insert(l.ctx, &model.Admin{
		Id:       id,
		Username: in.Username,
		Password: hashedPassword,
		RealName: in.RealName,
		Avatar:   in.Avatar,
		RoleId:   in.RoleId,
		Status:   1,
	})
	if err != nil {
		l.Logger.Errorf("创建管理员失败,error=%v", err)
		return nil, err
	}

	// 记录操作日志
	durationMs := time.Since(startTime).Milliseconds()
	utils.RecordAdminLog(l.ctx, l.svcCtx, utils.LogModuleAdmin, "创建管理员", in, &admin.Empty{}, durationMs)

	l.Logger.Infof("创建管理员成功,operatorId=%d", operatorId)
	return &admin.Empty{}, nil
}
