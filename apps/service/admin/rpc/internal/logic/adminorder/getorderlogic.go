package adminorderlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	ordertypes "github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOrderLogic) GetOrder(in *admin.AdminGetOrderReq) (*admin.AdminOrderDetailResp, error) {
	resp, err := l.svcCtx.OrderRpc.GetOrder(l.ctx, &ordertypes.AdminGetOrderRequest{
		OrderSn: in.OrderSn,
	})
	if err != nil {
		l.Logger.Errorf("查询订单详情失败,order_sn=%s,error=%v", in.OrderSn, err)
		return nil, err
	}

	var items []*admin.AdminOrderProductItem
	for _, item := range resp.Items {
		items = append(items, &admin.AdminOrderProductItem{
			SkuId:       item.SkuId,
			SpuId:       item.SpuId,
			ProductName: item.ProductName,
			SkuName:     item.SkuName,
			Pic:         item.Pic,
			Price:       item.Price,
			Quantity:    item.Quantity,
			TotalAmount: item.TotalAmount,
		})
	}

	var address *admin.AdminOrderAddress
	if resp.Address != nil {
		address = &admin.AdminOrderAddress{
			ReceiverName:  resp.Address.ReceiverName,
			ReceiverPhone: resp.Address.ReceiverPhone,
			Province:      resp.Address.Province,
			City:          resp.Address.City,
			District:      resp.Address.District,
			DetailAddress: resp.Address.DetailAddress,
		}
	}

	return &admin.AdminOrderDetailResp{
		OrderSn:         resp.OrderSn,
		UserId:          resp.UserId,
		UserPhone:       resp.UserPhone,
		Status:          admin.AdminOrderStatus(resp.Status),
		StatusDesc:      resp.StatusDesc,
		PayType:         admin.AdminPayType(resp.PayType),
		PayTypeDesc:     resp.PayTypeDesc,
		PayAmount:       resp.PayAmount,
		TotalAmount:     resp.TotalAmount,
		FreightAmount:   resp.FreightAmount,
		CouponAmount:    resp.CouponAmount,
		Remark:          resp.Remark,
		Address:         address,
		Items:           items,
		DeliverySn:      resp.DeliverySn,
		DeliveryCompany: resp.DeliveryCompany,
		CreateTime:      resp.CreateTime,
		PayTime:         resp.PayTime,
		DeliveryTime:    resp.DeliveryTime,
		FinishTime:      resp.FinishTime,
		CancelTime:      resp.CancelTime,
	}, nil
}
