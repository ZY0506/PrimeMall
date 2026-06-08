package marketinglogic

import (
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"
)

func couponToProto(c *model.Coupon) *marketing.CouponInfo {
	if c == nil {
		return nil
	}
	return &marketing.CouponInfo{
		Id:                c.Id,
		Name:              c.Name,
		Type:              marketing.CouponType(c.Type),
		ThresholdAmount:   c.ThresholdAmount,
		ReduceAmount:      c.ReduceAmount,
		DiscountRate:      int32(c.DiscountRate),
		MaxDiscountAmount: c.MaxDiscountAmount,
		TotalQuantity:     int32(c.TotalQuantity),
		UsedQuantity:      int32(c.UsedQuantity),
		PerUserLimit:      int32(c.PerUserLimit),
		StartTime:         c.StartTime.Unix(),
		EndTime:           c.EndTime.Unix(),
		Status:            int32(c.Status),
		Description:       c.Description,
		CreatedAt:         c.CreatedAt.Unix(),
	}
}

func userCouponToProto(c *model.Coupon, uc *model.UserCoupon, reason string) *marketing.UserCouponInfo {
	if uc == nil {
		return nil
	}
	info := &marketing.UserCouponInfo{
		Id:                uc.Id,
		CouponId:          uc.CouponId,
		UserId:            uc.UserId,
		Type:              marketing.CouponType(c.Type),
		Status:            int32(uc.Status),
		ExpireTime:        uc.ExpireTime.Unix(),
		OrderSn:           uc.OrderSn,
		UnavailableReason: reason,
		ThresholdAmount:   c.ThresholdAmount,
		ReduceAmount:      c.ReduceAmount,
		DiscountRate:      int32(c.DiscountRate),
		MaxDiscountAmount: c.MaxDiscountAmount,
	}
	if c != nil {
		info.Name = c.Name
	}
	return info
}
