package coupon

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CouponListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCouponListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CouponListLogic {
	return &CouponListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CouponListLogic) CouponList(req *types.CouponListReq) (resp *types.CouponListResp, err error) {
	// 将 HTTP 上下文中的 userId 注入到 RPC 调用上下文
	l.ctx, _ = ctxdata.ParseAndPutUserIdToCtx(l.ctx)

	rpcResp, err := l.svcCtx.MarketingRpc.ListCoupons(l.ctx, &marketing.ListCouponsReq{
		Page: &marketing.PageReq{
			Page: req.Page,
			Size: req.Size,
		},
	})
	if err != nil {
		l.Logger.Errorf("CouponList RPC error: %v", err)
		return nil, err
	}

	list := make([]types.Coupon, 0, len(rpcResp.List))
	for _, item := range rpcResp.List {
		list = append(list, types.Coupon{
			Id:                item.Id,
			Name:              item.Name,
			Type:              int64(item.Type),
			ThresholdAmount:   item.ThresholdAmount,
			ReduceAmount:      item.ReduceAmount,
			DiscountRate:      int64(item.DiscountRate),
			MaxDiscountAmount: item.MaxDiscountAmount,
			TotalQuantity:     int64(item.TotalQuantity),
			UsedQuantity:      int64(item.UsedQuantity),
			PerUserLimit:      int64(item.PerUserLimit),
			UserClaimCount:    int64(item.UserClaimCount),
			StartTime:         item.StartTime,
			EndTime:           item.EndTime,
			Description:       item.Description,
			IsClaimed:         item.UserClaimCount > 0,
		})
	}

	return &types.CouponListResp{
		Total: rpcResp.Page.Total,
		List:  list,
	}, nil
}
