package marketinglogic

import (
	"context"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type UseCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUseCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UseCouponLogic {
	return &UseCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UseCouponLogic) UseCoupon(in *marketing.UseCouponReq) (*marketing.UseCouponResp, error) {
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("使用优惠券：获取用户ID失败，错误：%v", err)
		return &marketing.UseCouponResp{Success: false, ErrorMsg: "获取用户信息失败"}, nil
	}

	uc, err := l.svcCtx.UserCouponModel.FindOne(l.ctx, in.UserCouponId)
	if err != nil {
		l.Logger.Errorf("使用优惠券：查询用户优惠券失败，错误：%v", err)
		return &marketing.UseCouponResp{Success: false, ErrorMsg: "优惠券不存在"}, nil
	}
	if uc == nil || uc.UserId != userId {
		return &marketing.UseCouponResp{Success: false, ErrorMsg: "优惠券不存在"}, nil
	}
	if uc.Status != 0 {
		return &marketing.UseCouponResp{Success: false, ErrorMsg: "优惠券已使用或已过期"}, nil
	}

	coupon, err := l.svcCtx.CouponModel.FindOne(l.ctx, uc.CouponId)
	if err != nil || coupon == nil {
		return &marketing.UseCouponResp{Success: false, ErrorMsg: "优惠券不存在"}, nil
	}

	now := time.Now()
	if now.After(uc.ExpireTime) {
		return &marketing.UseCouponResp{Success: false, ErrorMsg: "优惠券已过期"}, nil
	}

	// 计算折扣金额
	discount := calcDiscount(coupon, in.OrderSn)

	err = l.svcCtx.UserCouponModel.UpdateStatus(l.ctx, in.UserCouponId, 1, in.OrderSn, now)
	if err != nil {
		l.Logger.Errorf("使用优惠券：更新状态失败，错误：%v", err)
		return &marketing.UseCouponResp{Success: false, ErrorMsg: "使用失败"}, nil
	}

	return &marketing.UseCouponResp{
		Success:        true,
		DiscountAmount: discount,
	}, nil
}

func calcDiscount(coupon *model.Coupon, _ string) int64 {
	switch coupon.Type {
	case 1: // 满减券
		return coupon.ReduceAmount
	case 2: // 折扣券
		// The discount needs order amount; default to max discount
		if coupon.MaxDiscountAmount > 0 {
			return coupon.MaxDiscountAmount
		}
		return 0
	case 3: // 无门槛券
		return coupon.ReduceAmount
	default:
		return 0
	}
}
