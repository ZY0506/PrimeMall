package adminuserlogic

import (
	"context"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/logic/utils"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	userrpc "github.com/ZY0506/PrimeMall/apps/service/user/rpc/client/admin"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnbanUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnbanUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnbanUserLogic {
	return &UnbanUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnbanUserLogic) UnbanUser(in *admin.AdminUnbanUserReq) (*admin.Empty, error) {
	startTime := time.Now()

	// 从上下文中获取管理员信息
	adminId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取管理员ID失败,err=%v", err)
		return nil, err
	}

	// 获取管理员信息
	adminInfo, err := l.svcCtx.AdminModel.FindOne(l.ctx, adminId)
	if err != nil {
		l.Logger.Errorf("获取管理员信息失败,adminId=%d,err=%v", adminId, err)
		return nil, err
	}

	operator := fmt.Sprintf("%s(%d)", adminInfo.RealName, adminId)

	// 调用用户服务解封用户
	_, err = l.svcCtx.UserRpc.UnbanUser(l.ctx, &userrpc.UnbanUserReq{
		UserId:   in.UserId,
		Reason:   in.Reason,
		Operator: operator,
	})
	if err != nil {
		l.Logger.Errorf("解封用户失败,userId=%d,err=%v", in.UserId, err)
		return nil, err
	}

	// 记录操作日志
	durationMs := time.Since(startTime).Milliseconds()
	utils.RecordAdminLog(l.ctx, l.svcCtx, utils.LogModuleUser, "解封用户", in, nil, durationMs)

	l.Logger.Infof("解封用户成功,userId=%d,operator=%s", in.UserId, operator)
	return &admin.Empty{}, nil
}
