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
	discount := calcDiscount(coupon, in.OrderAmount)
	if discount <= 0 {
		return &marketing.UseCouponResp{Success: false, ErrorMsg: "未达到使用门槛"}, nil
	}

	// 乐观锁扣减：WHERE id = ? AND status = 0 保证并发安全
	affected, err := l.svcCtx.UserCouponModel.UpdateStatus(l.ctx, in.UserCouponId, 1, in.OrderSn, now, 0)
	if err != nil {
		l.Logger.Errorf("使用优惠券：更新状态失败，错误：%v", err)
		return &marketing.UseCouponResp{Success: false, ErrorMsg: "使用失败"}, nil
	}
	if affected == 0 {
		l.Logger.Errorf("使用优惠券：并发扣减失败，优惠券已被使用，user_coupon_id=%d", in.UserCouponId)
		return &marketing.UseCouponResp{Success: false, ErrorMsg: "优惠券已被使用"}, nil
	}

	return &marketing.UseCouponResp{
		Success:        true,
		DiscountAmount: discount,
	}, nil
}

// calcDiscount 根据优惠券类型和订单金额计算折扣
// coupon: 优惠券定义
// orderAmount: 订单总金额（分），用于门槛判断和折扣券计算
func calcDiscount(coupon *model.Coupon, orderAmount int64) int64 {
	switch coupon.Type {
	case 1: // 满减券: 满 thresholdAmount 减 reduceAmount
		if orderAmount >= coupon.ThresholdAmount {
			return coupon.ReduceAmount
		}
		return 0
	case 2: // 折扣券: 打 discountRate/100 折，封顶 maxDiscountAmount
		if orderAmount <= 0 {
			return 0
		}
		// discountRate 是万分比，如 8000 = 8折
		discount := orderAmount * int64(coupon.DiscountRate) / 10000
		discount = orderAmount - discount // 实际减免金额
		if coupon.MaxDiscountAmount > 0 && discount > coupon.MaxDiscountAmount {
			discount = coupon.MaxDiscountAmount
		}
		if discount < 0 {
			discount = 0
		}
		return discount
	case 3: // 无门槛券: 直接减 reduceAmount
		discount := coupon.ReduceAmount
		if discount > orderAmount {
			discount = orderAmount // 减免不超过订单金额
		}
		return discount
	default:
		return 0
	}
}
