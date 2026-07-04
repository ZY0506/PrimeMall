package admincouponlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCouponsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCouponsLogic {
	return &ListCouponsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListCouponsLogic) ListCoupons(in *admin.AdminListCouponsReq) (*admin.AdminListCouponsResp, error) {
	resp, err := l.svcCtx.MarketingRpc.ListCoupons(l.ctx, &marketing.ListCouponsReq{
		Page: &marketing.PageReq{Page: in.Page, Size: in.PageSize},
	})
	if err != nil {
		l.Logger.Errorf("优惠券列表：调用营销服务失败，error=%v", err)
		return nil, err
	}
	list := make([]*admin.AdminCouponItem, 0, len(resp.List))
	for _, c := range resp.List {
		list = append(list, &admin.AdminCouponItem{
			Id:                c.Id,
			Name:              c.Name,
			Type:              int64(c.Type),
			ThresholdAmount:   c.ThresholdAmount,
			ReduceAmount:      c.ReduceAmount,
			DiscountRate:      int32(c.DiscountRate),
			MaxDiscountAmount: c.MaxDiscountAmount,
			TotalQuantity:     c.TotalQuantity,
			UsedQuantity:      c.UsedQuantity,
			PerUserLimit:      c.PerUserLimit,
			StartTime:         c.StartTime,
			EndTime:           c.EndTime,
			Status:            c.Status,
			Description:       c.Description,
			CreatedAt:         c.CreatedAt,
		})
	}
	return &admin.AdminListCouponsResp{Total: resp.Page.Total, List: list}, nil
}
