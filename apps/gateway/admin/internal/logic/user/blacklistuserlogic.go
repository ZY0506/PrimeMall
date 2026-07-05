// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type BlacklistUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBlacklistUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlacklistUserLogic {
	return &BlacklistUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BlacklistUserLogic) BlacklistUser(req *types.BlackUserReq) error {
	if req.Id == 0 {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "用户ID不能为空")
	}
	_, err := l.svcCtx.AdminUserRpc.BanUser(l.ctx, &admin.AdminBanUserReq{
		UserId:     req.Id,
		ActionType: 1, // 1=禁止下单
		Reason:     req.Reason,
	})
	if err != nil {
		l.Logger.Errorf("封禁用户失败 userId=%d: %v", req.Id, err)
		return err
	}
	l.Logger.Infof("封禁用户成功 userId=%d reason=%s", req.Id, req.Reason)
	return nil
}
