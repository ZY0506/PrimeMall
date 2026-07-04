package admin

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveCouponLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveCouponLogic {
	return &SaveCouponLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *SaveCouponLogic) SaveCoupon(req *types.SaveCouponReq) error {
	if req.Id > 0 {
		// 更新优惠券
		startTime := req.StartTime
		endTime := req.EndTime
		_, err := l.svcCtx.AdminCouponRpc.UpdateCoupon(l.ctx, &admin.AdminUpdateCouponReq{
			Id:                req.Id,
			Name:              req.Name,
			Type:              req.Type,
			ThresholdAmount:   req.ThresholdAmount,
			ReduceAmount:      req.ReduceAmount,
			DiscountRate:      req.DiscountRate,
			MaxDiscountAmount: req.MaxDiscountAmount,
			TotalQuantity:     int32(req.TotalQuantity),
			PerUserLimit:      int32(req.PerUserLimit),
			StartTime:         startTime,
			EndTime:           endTime,
			Description:       req.Description,
		})
		if err != nil {
			l.Logger.Errorf("更新优惠券失败: %v", err)
			return err
		}
	} else {
		// 创建优惠券
		startTime := req.StartTime
		endTime := req.EndTime
		_, err := l.svcCtx.AdminCouponRpc.CreateCoupon(l.ctx, &admin.AdminCreateCouponReq{
			Name:              req.Name,
			Type:              req.Type,
			ThresholdAmount:   req.ThresholdAmount,
			ReduceAmount:      req.ReduceAmount,
			DiscountRate:      req.DiscountRate,
			MaxDiscountAmount: req.MaxDiscountAmount,
			TotalQuantity:     int32(req.TotalQuantity),
			PerUserLimit:      int32(req.PerUserLimit),
			StartTime:         startTime,
			EndTime:           endTime,
			Description:       req.Description,
		})
		if err != nil {
			l.Logger.Errorf("创建优惠券失败: %v", err)
			return err
		}
	}
	return nil
}
