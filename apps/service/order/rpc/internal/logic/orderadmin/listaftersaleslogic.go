package orderadminlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/common/constants"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAfterSalesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAfterSalesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAfterSalesLogic {
	return &ListAfterSalesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListAfterSales - 售后列表查询
func (l *ListAfterSalesLogic) ListAfterSales(in *order.AdminListAfterSalesRequest) (*order.AdminListAfterSalesResponse, error) {
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 20
	}

	afterSaleList, total, err := l.svcCtx.AfterSaleModel.AdminFindPage(l.ctx, int64(in.Status), in.Page, in.PageSize)
	if err != nil {
		l.Logger.Errorf("查询售后列表失败,error=%v", err)
		return nil, err
	}

	if total == 0 {
		return &order.AdminListAfterSalesResponse{
			Total: 0,
			List:  []*order.AdminAfterSaleListItem{},
		}, nil
	}

	afterSaleSns := make([]string, 0, len(afterSaleList))
	for _, afterSale := range afterSaleList {
		afterSaleSns = append(afterSaleSns, afterSale.AfterSaleSn)
	}

	itemMap, err := l.svcCtx.AfterSaleItemModel.FindMapByAfterSaleSns(l.ctx, afterSaleSns)
	if err != nil {
		l.Logger.Errorf("批量查询售后商品失败,error=%v", err)
		return nil, err
	}

	resultList := make([]*order.AdminAfterSaleListItem, 0, len(afterSaleList))
	for _, afterSale := range afterSaleList {
		item := itemMap[afterSale.AfterSaleSn]
		if item == nil {
			continue
		}

		resultList = append(resultList, &order.AdminAfterSaleListItem{
			AfterSaleId:  afterSale.Id,
			OrderSn:      afterSale.OrderSn,
			Type:         order.AfterSaleType(afterSale.Type),
			Status:       order.AfterSaleStatus(afterSale.Status),
			StatusDesc:   constants.AfterSaleStatusMap[int(afterSale.Status)],
			RefundAmount: afterSale.RealRefundAmount,
			SpuName:      item.ProductName,
			SkuName:      item.SkuName,
			Pic:          item.SkuPic,
			CreatedAt:    timestamppb.New(afterSale.CreatedAt),
		})
	}

	l.Logger.Infof("查询售后列表成功,total=%d", total)
	return &order.AdminListAfterSalesResponse{
		Total: total,
		List:  resultList,
	}, nil
}
