package marketinglogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

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
		l.Logger.Errorf("获取优惠券列表：分页查询失败，错误：%v", err)
		return nil, err
	}

	// 获取当前用户 ID（未登录为 0）
	userId, _ := ctxdata.GetUserIdFromCtx(l.ctx)

	// 批量查询用户领取计数
	var claimCounts map[uint64]int64
	if userId > 0 && len(items) > 0 {
		couponIDs := make([]uint64, 0, len(items))
		for _, item := range items {
			couponIDs = append(couponIDs, item.Id)
		}
		claimCounts, _ = l.svcCtx.UserCouponModel.CountByUserAndCouponIDs(l.ctx, userId, couponIDs)
	}

	var list []*marketing.CouponInfo
	for _, item := range items {
		pb := couponToProto(item)
		if claimCounts != nil {
			if cnt, ok := claimCounts[item.Id]; ok {
				pb.UserClaimCount = int32(cnt)
			}
		}
		list = append(list, pb)
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
