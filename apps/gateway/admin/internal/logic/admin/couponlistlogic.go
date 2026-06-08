package admin

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type CouponListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCouponListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CouponListLogic {
	return &CouponListLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *CouponListLogic) CouponList(req *types.AdminCouponListReq) (resp *types.AdminCouponListResp, err error) {
	return &types.AdminCouponListResp{}, nil
}
