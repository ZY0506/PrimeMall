package admin

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCouponStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCouponStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCouponStatusLogic {
	return &UpdateCouponStatusLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *UpdateCouponStatusLogic) UpdateCouponStatus(req *types.UpdateStatusReq) error {
	_, err := l.svcCtx.AdminCouponRpc.UpdateCouponStatus(l.ctx, &admin.AdminUpdateCouponStatusReq{
		Id:     req.Id,
		Status: int32(req.Status),
	})
	if err != nil {
		l.Logger.Errorf("更新优惠券状态失败, id=%d, error=%v", req.Id, err)
		return err
	}
	return nil
}
