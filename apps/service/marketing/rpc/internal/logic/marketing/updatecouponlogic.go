package marketinglogic

import (
	"context"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCouponLogic {
	return &UpdateCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCouponLogic) UpdateCoupon(in *marketing.UpdateCouponReq) (*marketing.Empty, error) {
	existing, err := l.svcCtx.CouponModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("更新优惠券：查询失败，错误：%v", err)
		return nil, err
	}

	existing.Name = in.Name
	existing.Type = int64(in.Type)
	existing.ThresholdAmount = in.ThresholdAmount
	existing.ReduceAmount = in.ReduceAmount
	existing.DiscountRate = int64(in.DiscountRate)
	existing.MaxDiscountAmount = in.MaxDiscountAmount
	existing.TotalQuantity = int64(in.TotalQuantity)
	existing.PerUserLimit = int64(in.PerUserLimit)
	existing.StartTime = time.Unix(in.StartTime, 0)
	existing.EndTime = time.Unix(in.EndTime, 0)
	existing.Description = in.Description
	if in.Status != 0 {
		existing.Status = int64(in.Status)
	}
	existing.UpdatedAt = time.Now()
	if err = l.svcCtx.CouponModel.Update(l.ctx, existing); err != nil {
		l.Logger.Errorf("更新优惠券：更新数据失败，错误：%v", err)
		return nil, err
	}
	return &marketing.Empty{}, nil
}
