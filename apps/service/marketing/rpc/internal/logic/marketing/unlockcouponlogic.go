package marketinglogic

import (
	"context"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnlockCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnlockCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlockCouponLogic {
	return &UnlockCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnlockCouponLogic) UnlockCoupon(in *marketing.UnlockCouponReq) (*marketing.UnlockCouponResp, error) {
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("解锁优惠券：获取用户ID失败，错误：%v", err)
		return &marketing.UnlockCouponResp{Success: false, ErrorMsg: "获取用户信息失败"}, nil
	}

	uc, err := l.svcCtx.UserCouponModel.FindOne(l.ctx, in.UserCouponId)
	if err != nil {
		l.Logger.Errorf("解锁优惠券：查询用户优惠券失败，错误：%v", err)
		return &marketing.UnlockCouponResp{Success: false, ErrorMsg: "优惠券不存在"}, nil
	}
	if uc == nil || uc.UserId != userId {
		return &marketing.UnlockCouponResp{Success: false, ErrorMsg: "优惠券不存在"}, nil
	}
	if uc.Status != 1 {
		return &marketing.UnlockCouponResp{Success: false, ErrorMsg: "优惠券状态异常"}, nil
	}

	// 恢复为未使用状态
	zeroTime := time.Time{}
	affected, err := l.svcCtx.UserCouponModel.UpdateStatus(l.ctx, in.UserCouponId, 0, "", zeroTime, 1)
	if err != nil {
		l.Logger.Errorf("解锁优惠券：更新状态失败，错误：%v", err)
		return &marketing.UnlockCouponResp{Success: false, ErrorMsg: "解锁失败"}, nil
	}
	if affected == 0 {
		l.Logger.Errorf("解锁优惠券：并发竞争，优惠券状态异常，user_coupon_id=%d", in.UserCouponId)
		return &marketing.UnlockCouponResp{Success: false, ErrorMsg: "优惠券状态异常"}, nil
	}

	return &marketing.UnlockCouponResp{Success: true}, nil
}
