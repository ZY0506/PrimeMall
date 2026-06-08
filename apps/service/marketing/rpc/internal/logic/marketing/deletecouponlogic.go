package marketinglogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCouponLogic {
	return &DeleteCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteCouponLogic) DeleteCoupon(in *marketing.IdReq) (*marketing.Empty, error) {
	err := l.svcCtx.CouponModel.Delete(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("删除优惠券：删除失败，错误：%v", err)
		return nil, err
	}
	return &marketing.Empty{}, nil
}
