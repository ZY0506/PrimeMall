package marketinglogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type MyCouponsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMyCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MyCouponsLogic {
	return &MyCouponsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MyCouponsLogic) MyCoupons(in *marketing.MyCouponsReq) (*marketing.MyCouponsResp, error) {
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取我的优惠券：获取用户ID失败，错误：%v", err)
		return nil, err
	}

	page := in.Page.GetPage()
	size := in.Page.GetSize()
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	items, total, err := l.svcCtx.UserCouponModel.FindByUserID(l.ctx, userId, int64(in.Status), page, size)
	if err != nil {
		l.Logger.Errorf("获取我的优惠券：查询用户优惠券失败，错误：%v", err)
		return nil, err
	}

	var list []*marketing.UserCouponInfo
	for _, uc := range items {
		coupon, _ := l.svcCtx.CouponModel.FindOne(l.ctx, uc.CouponId)
		list = append(list, userCouponToProto(coupon, uc, ""))
	}
	if list == nil {
		list = []*marketing.UserCouponInfo{}
	}

	return &marketing.MyCouponsResp{
		Page: &marketing.PageResp{
			Total: total,
			Page:  page,
			Size:  size,
		},
		List: list,
	}, nil
}
