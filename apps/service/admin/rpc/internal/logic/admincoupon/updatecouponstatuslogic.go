package admincouponlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCouponStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCouponStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCouponStatusLogic {
	return &UpdateCouponStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCouponStatusLogic) UpdateCouponStatus(in *admin.AdminUpdateCouponStatusReq) (*admin.Empty, error) {
	// 先获取当前优惠券详情，更新 status 后调用营销服务的 UpdateCoupon
	coupon, err := l.svcCtx.MarketingRpc.GetCoupon(l.ctx, &marketing.IdReq{Id: in.Id})
	if err != nil {
		l.Logger.Errorf("更新优惠券状态：查询优惠券失败，id=%d, error=%v", in.Id, err)
		return nil, err
	}

	info := coupon.Coupon
	_, err = l.svcCtx.MarketingRpc.UpdateCoupon(l.ctx, &marketing.UpdateCouponReq{
		Id:                info.Id,
		Name:              info.Name,
		Type:              info.Type,
		ThresholdAmount:   info.ThresholdAmount,
		ReduceAmount:      info.ReduceAmount,
		DiscountRate:      info.DiscountRate,
		MaxDiscountAmount: info.MaxDiscountAmount,
		TotalQuantity:     info.TotalQuantity,
		PerUserLimit:      info.PerUserLimit,
		StartTime:         info.StartTime,
		EndTime:           info.EndTime,
		Description:       info.Description,
		Status:            in.Status,
	})
	if err != nil {
		l.Logger.Errorf("更新优惠券状态：调用营销服务失败，id=%d, error=%v", in.Id, err)
		return nil, err
	}
	return &admin.Empty{}, nil
}
