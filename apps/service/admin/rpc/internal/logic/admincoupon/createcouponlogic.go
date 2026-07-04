package admincouponlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCouponLogic {
	return &CreateCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCouponLogic) CreateCoupon(in *admin.AdminCreateCouponReq) (*admin.Empty, error) {
	_, err := l.svcCtx.MarketingRpc.CreateCoupon(l.ctx, &marketing.CreateCouponReq{
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
		l.Logger.Errorf("创建优惠券：调用营销服务失败，error=%v", err)
		return nil, err
	}
	return &admin.Empty{}, nil
}
