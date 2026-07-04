package admin

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
)

type RecoverUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecoverUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecoverUserLogic {
	return &RecoverUserLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *RecoverUserLogic) RecoverUser(req *types.RecoverUserReq) error {
	if req.Id == 0 {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "用户ID不能为空")
	}
	_, err := l.svcCtx.AdminUserRpc.UnbanUser(l.ctx, &admin.AdminUnbanUserReq{
		UserId: req.Id,
	})
	if err != nil {
		l.Logger.Errorf("解封用户失败 userId=%d: %v", req.Id, err)
		return err
	}
	l.Logger.Infof("解封用户成功 userId=%d", req.Id)
	return nil
}
