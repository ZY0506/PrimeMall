package marketinglogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/internal/svc"
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

func (l *ListCouponsLogic) ListCoupons(in *marketing.ListCouponsReq) (*marketing.ListCouponsResp, error) {
	page := in.Page.GetPage()
	size := in.Page.GetSize()
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	items, total, err := l.svcCtx.CouponModel.FindPageList(l.ctx, page, size)
	if err != nil {
		l.Logger.Errorf("ListCoupons FindPageList error: %v", err)
		return nil, err
	}

	var list []*marketing.CouponInfo
	for _, item := range items {
		list = append(list, couponToProto(item))
	}
	if list == nil {
		list = []*marketing.CouponInfo{}
	}

	return &marketing.ListCouponsResp{
		Page: &marketing.PageResp{
			Total: total,
			Page:  page,
			Size:  size,
		},
		List: list,
	}, nil
}
