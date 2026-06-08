package admin

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
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
	return nil
}
