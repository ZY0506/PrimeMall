package marketinglogic

import (
	"context"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/internal/svc"
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

func (l *CreateCouponLogic) CreateCoupon(in *marketing.CreateCouponReq) (*marketing.Empty, error) {
	now := time.Now()
	data := &model.Coupon{
		Name:              in.Name,
		Type:              int64(in.Type),
		ThresholdAmount:   in.ThresholdAmount,
		ReduceAmount:      in.ReduceAmount,
		DiscountRate:      int64(in.DiscountRate),
		MaxDiscountAmount: in.MaxDiscountAmount,
		TotalQuantity:     int64(in.TotalQuantity),
		UsedQuantity:      0,
		PerUserLimit:      int64(in.PerUserLimit),
		StartTime:         time.Unix(in.StartTime, 0),
		EndTime:           time.Unix(in.EndTime, 0),
		Status:            1,
		Description:       in.Description,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	_, err := l.svcCtx.CouponModel.Insert(l.ctx, data)
	if err != nil {
		l.Logger.Errorf("创建优惠券：插入数据失败，错误：%v", err)
		return nil, err
	}
	return &marketing.Empty{}, nil
}
