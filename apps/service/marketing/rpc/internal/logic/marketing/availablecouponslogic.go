package marketinglogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type AvailableCouponsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAvailableCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AvailableCouponsLogic {
	return &AvailableCouponsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AvailableCouponsLogic) AvailableCoupons(in *marketing.AvailableCouponsReq) (*marketing.AvailableCouponsResp, error) {
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("AvailableCoupons GetUserIdFromCtx error: %v", err)
		return nil, err
	}

	userCoupons, err := l.svcCtx.UserCouponModel.FindAvailableByUser(l.ctx, userId)
	if err != nil {
		l.Logger.Errorf("AvailableCoupons FindAvailableByUser error: %v", err)
		return nil, err
	}

	var available []*marketing.UserCouponInfo
	var unavailable []*marketing.UserCouponInfo

	for _, uc := range userCoupons {
		coupon, err := l.svcCtx.CouponModel.FindOne(l.ctx, uc.CouponId)
		if err != nil || coupon == nil {
			continue
		}
		reason := ""
		if in.OrderAmount < coupon.ThresholdAmount {
			reason = "未达到使用门槛"
		}
		info := userCouponToProto(coupon, uc, reason)
		if reason == "" {
			available = append(available, info)
		} else {
			unavailable = append(unavailable, info)
		}
	}

	if available == nil {
		available = []*marketing.UserCouponInfo{}
	}
	if unavailable == nil {
		unavailable = []*marketing.UserCouponInfo{}
	}

	return &marketing.AvailableCouponsResp{
		Available:   available,
		Unavailable: unavailable,
	}, nil
}
