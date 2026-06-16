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

type GetUserCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserCouponLogic {
	return &GetUserCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserCouponLogic) GetUserCoupon(in *marketing.GetUserCouponReq) (*marketing.GetUserCouponResp, error) {
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户优惠券：获取用户ID失败，错误：%v", err)
		return nil, err
	}

	var uc *model.UserCoupon

	if in.UserCouponId > 0 {
		// 按 user_coupon_id 精确查询
		uc, err = l.svcCtx.UserCouponModel.FindOne(l.ctx, in.UserCouponId)
		if err != nil {
			l.Logger.Errorf("获取用户优惠券：查询用户优惠券失败，错误：%v", err)
			return nil, err
		}
		if uc == nil || uc.UserId != userId {
			return &marketing.GetUserCouponResp{Coupon: nil}, nil
		}
	} else if in.CouponId > 0 {
		// 按 coupon 定义ID 查询该用户已领取的可用优惠券
		uc, err = l.svcCtx.UserCouponModel.FindByUserAndCoupon(l.ctx, userId, in.CouponId)
		if err != nil {
			l.Logger.Errorf("获取用户优惠券：按用户和券定义查询失败，错误：%v", err)
			return nil, err
		}
		if uc == nil {
			l.Logger.Infof("用户未领取该优惠券：userId=%d, couponId=%d", userId, in.CouponId)
			return &marketing.GetUserCouponResp{Coupon: nil}, nil
		}
		l.Logger.Infof("找到用户优惠券实例：userCouponId=%d, status=%d, expireTime=%v",
			uc.Id, uc.Status, uc.ExpireTime)
	} else {
		return &marketing.GetUserCouponResp{Coupon: nil}, nil
	}

	coupon, err := l.svcCtx.CouponModel.FindOne(l.ctx, uc.CouponId)
	if err != nil || coupon == nil {
		l.Logger.Errorf("获取用户优惠券：查询优惠券定义失败，错误：%v", err)
		return &marketing.GetUserCouponResp{Coupon: nil}, nil
	}

	now := time.Now()
	status := uc.Status
	if status == 0 && now.After(uc.ExpireTime) {
		status = 2 // 已过期
	}

	unavailableReason := ""
	if status != 0 {
		switch status {
		case 1:
			unavailableReason = "已使用"
		case 2:
			unavailableReason = "已过期"
		}
	}

	couponType := marketing.CouponType(coupon.Type)
	resp := &marketing.GetUserCouponResp{
		Coupon: &marketing.UserCouponInfo{
			Id:                uc.Id,
			CouponId:          uc.CouponId,
			UserId:            uc.UserId,
			Name:              coupon.Name,
			Type:              couponType,
			Status:            int32(status),
			ExpireTime:        uc.ExpireTime.Unix(),
			OrderSn:           uc.OrderSn,
			UnavailableReason: unavailableReason,
			ThresholdAmount:   coupon.ThresholdAmount,
			ReduceAmount:      coupon.ReduceAmount,
			DiscountRate:      int32(coupon.DiscountRate),
			MaxDiscountAmount: coupon.MaxDiscountAmount,
		},
	}

	return resp, nil
}
