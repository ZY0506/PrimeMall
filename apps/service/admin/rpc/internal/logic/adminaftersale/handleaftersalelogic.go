package adminaftersalelogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	ordertypes "github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleAfterSaleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleAfterSaleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleAfterSaleLogic {
	return &HandleAfterSaleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *HandleAfterSaleLogic) HandleAfterSale(in *admin.AdminHandleAfterSaleReq) (*admin.Empty, error) {
	_, err := l.svcCtx.OrderRpc.HandleAfterSale(l.ctx, &ordertypes.AdminHandleAfterSaleRequest{
		AfterSaleId: in.AfterSaleId,
		Action:      ordertypes.AfterSaleAuditAction(in.Action),
		Remark:      in.Remark,
	})
	if err != nil {
		l.Logger.Errorf("处理售后失败,after_sale_id=%d,error=%v", in.AfterSaleId, err)
		return nil, err
	}

	return &admin.Empty{}, nil
}
