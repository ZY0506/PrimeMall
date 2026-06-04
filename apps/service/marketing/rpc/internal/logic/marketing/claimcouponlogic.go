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

type ClaimCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClaimCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClaimCouponLogic {
	return &ClaimCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ClaimCouponLogic) ClaimCoupon(in *marketing.ClaimCouponReq) (*marketing.ClaimCouponResp, error) {
	// 从gRPC context获取userId
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("ClaimCoupon GetUserIdFromCtx error: %v", err)
		return nil, err
	}

	// 1. 查找优惠券
	coupon, err := l.svcCtx.CouponModel.FindOne(l.ctx, in.CouponId)
	if err != nil {
		l.Logger.Errorf("ClaimCoupon FindOne error: %v", err)
		return nil, err
	}
	if coupon == nil {
		return &marketing.ClaimCouponResp{}, nil
	}

	// 2. 检查状态和有效期
	if coupon.Status != 1 {
		return &marketing.ClaimCouponResp{}, nil
	}
	now := time.Now()
	if now.Before(coupon.StartTime) || now.After(coupon.EndTime) {
		return &marketing.ClaimCouponResp{}, nil
	}

	// 3. 检查库存
	if coupon.UsedQuantity >= coupon.TotalQuantity {
		return &marketing.ClaimCouponResp{}, nil
	}

	// 4. 检查用户领取次数
	count, err := l.svcCtx.UserCouponModel.CountByUserAndCoupon(l.ctx, userId, in.CouponId)
	if err != nil {
		l.Logger.Errorf("领取优惠券：统计用户领取次数失败，错误：%v", err)
		return nil, err
	}
	if int64(count) >= coupon.PerUserLimit {
		return &marketing.ClaimCouponResp{}, nil
	}

	// 5. 扣减库存 + 创建用户优惠券记录
	err = l.svcCtx.CouponModel.IncrUsedQuantity(l.ctx, in.CouponId, 1)
	if err != nil {
		l.Logger.Errorf("领取优惠券：扣减库存失败，错误：%v", err)
		return nil, err
	}

	expireTime := coupon.EndTime
	userCoupon := &model.UserCoupon{
		CouponId:   in.CouponId,
		UserId:     userId,
		OrderSn:    "",
		Status:     0,
		Source:     "ACTIVITY",
		CreatedAt:  now,
		ExpireTime: expireTime,
	}
	result, err := l.svcCtx.UserCouponModel.Insert(l.ctx, userCoupon)
	if err != nil {
		_ = l.svcCtx.CouponModel.IncrUsedQuantity(l.ctx, in.CouponId, -1)
		l.Logger.Errorf("领取优惠券：创建用户优惠券记录失败，错误：%v", err)
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &marketing.ClaimCouponResp{
		UserCouponId: uint64(id),
		ExpireTime:   expireTime.Unix(),
	}, nil
}
