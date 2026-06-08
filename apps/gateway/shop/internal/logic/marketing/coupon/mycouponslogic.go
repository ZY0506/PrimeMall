package coupon

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MyCouponsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMyCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MyCouponsLogic {
	return &MyCouponsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MyCouponsLogic) MyCoupons(req *types.MyCouponListReq) (resp *types.MyCouponListResp, err error) {
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}

	rpcResp, err := l.svcCtx.MarketingRpc.MyCoupons(l.ctx, &marketing.MyCouponsReq{
		Status: int32(req.Status),
		Page: &marketing.PageReq{
			Page: req.Page,
			Size: req.Size,
		},
	})
	if err != nil {
		l.Logger.Errorf("MyCoupons RPC error: %v", err)
		return nil, err
	}

	list := make([]types.UserCoupon, 0, len(rpcResp.List))
	for _, item := range rpcResp.List {
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

	return &types.MyCouponListResp{
		Total: rpcResp.Page.Total,
		List:  list,
	}, nil
}
