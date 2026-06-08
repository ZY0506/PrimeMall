package marketinglogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCouponLogic {
	return &GetCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCouponLogic) GetCoupon(in *marketing.IdReq) (*marketing.CouponResp, error) {
	coupon, err := l.svcCtx.CouponModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("GetCoupon FindOne error: %v", err)
		return nil, err
	}
	return &marketing.CouponResp{
		Coupon: couponToProto(coupon),
	}, nil
}
