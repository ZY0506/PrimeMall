package coupon

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AvailableCouponsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAvailableCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AvailableCouponsLogic {
	return &AvailableCouponsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AvailableCouponsLogic) AvailableCoupons(req *types.AvailableCouponsReq) (resp *types.AvailableCouponsResp, err error) {
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}

	rpcResp, err := l.svcCtx.MarketingRpc.AvailableCoupons(l.ctx, &marketing.AvailableCouponsReq{
		OrderAmount: req.OrderAmount,
	})
	if err != nil {
		l.Logger.Errorf("AvailableCoupons RPC error: %v", err)
		return nil, err
	}

	return &types.AvailableCouponsResp{
		Available:   toUserCouponList(rpcResp.Available),
		Unavailable: toUserCouponList(rpcResp.Unavailable),
	}, nil
}

func toUserCouponList(items []*marketing.UserCouponInfo) []types.UserCoupon {
	list := make([]types.UserCoupon, 0, len(items))
	for _, item := range items {
		list = append(list, types.UserCoupon{
			Id:                item.Id,
			CouponId:          item.CouponId,
			Name:              item.Name,
			Type:              int64(item.Type),
			Status:            int64(item.Status),
			ExpireTime:        item.ExpireTime,
			OrderSn:           item.OrderSn,
			UnavailableReason: item.UnavailableReason,
			ThresholdAmount:   item.ThresholdAmount,
			ReduceAmount:      item.ReduceAmount,
			DiscountRate:      int64(item.DiscountRate),
			MaxDiscountAmount: item.MaxDiscountAmount,
		})
	}
	return list
}
