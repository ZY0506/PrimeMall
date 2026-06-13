package admincouponlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCouponDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCouponDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCouponDetailLogic {
	return &GetCouponDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCouponDetailLogic) GetCouponDetail(in *admin.IdReq) (*admin.CouponDetailResp, error) {
	// 调用营销服务的 GetCoupon 获取完整优惠券信息
	coupon, err := l.svcCtx.MarketingRpc.GetCoupon(l.ctx, &marketing.IdReq{Id: in.Id})
	if err != nil {
		l.Logger.Errorf("获取优惠券详情：调用营销服务失败，id=%d, error=%v", in.Id, err)
		return nil, err
	}

	info := coupon.Coupon
	return &admin.CouponDetailResp{
		Id:                info.Id,
		Name:              info.Name,
		Type:              int64(info.Type),
		ThresholdAmount:   info.ThresholdAmount,
		ReduceAmount:      info.ReduceAmount,
		DiscountRate:      int64(info.DiscountRate),
		MaxDiscountAmount: info.MaxDiscountAmount,
		TotalQuantity:     info.TotalQuantity,
		UsedQuantity:      info.UsedQuantity,
		PerUserLimit:      info.PerUserLimit,
		StartTime:         info.StartTime,
		EndTime:           info.EndTime,
		Status:            int32(info.Status),
		Description:       info.Description,
		CreatedAt:         info.CreatedAt,
	}, nil
}
