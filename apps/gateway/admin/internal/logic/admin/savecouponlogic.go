package admin

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type SaveCouponLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveCouponLogic {
	return &SaveCouponLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *SaveCouponLogic) SaveCoupon(req *types.SaveCouponReq) error {
	return nil
}
