package admincouponlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
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

func (l *UpdateCouponLogic) UpdateCoupon(in *admin.AdminUpdateCouponReq) (*admin.Empty, error) {
	_, err := l.svcCtx.MarketingRpc.UpdateCoupon(l.ctx, &marketing.UpdateCouponReq{
		Id:                in.Id,
		Name:              in.Name,
		Type:              marketing.CouponType(in.Type),
		ThresholdAmount:   in.ThresholdAmount,
		ReduceAmount:      in.ReduceAmount,
		DiscountRate:      int32(in.DiscountRate),
		MaxDiscountAmount: in.MaxDiscountAmount,
		TotalQuantity:     in.TotalQuantity,
		PerUserLimit:      in.PerUserLimit,
		StartTime:         in.StartTime,
		EndTime:           in.EndTime,
		Description:       in.Description,
	})
	if err != nil {
		l.Logger.Errorf("更新优惠券：调用营销服务失败，id=%d, error=%v", in.Id, err)
		return nil, err
	}
	return &admin.Empty{}, nil
}
