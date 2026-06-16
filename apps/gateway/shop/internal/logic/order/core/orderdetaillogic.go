package core

import (
	"context"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	payment "github.com/ZY0506/PrimeMall/apps/service/payment/rpc/types/payment"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewOrderDetailLogic 订单详情
func NewOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderDetailLogic {
	return &OrderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderDetailLogic) OrderDetail(req *types.OrderSnPathReq) (resp *types.OrderDetailResp, err error) {
	// 注入userID
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}
	// 调用rpc
	orderDetail, err := l.svcCtx.OrderRpc.OrderDetail(l.ctx, &order.OrderDetailRequest{
		OrderSn: req.OrderSn,
	})
	if err != nil {
		l.Logger.Errorf("调用RPC获取订单详情失败，error=%v", err)
		return nil, err
	}

	// 封装返回响应
	orderItems := make([]types.OrderItem, 0)
	for _, item := range orderDetail.Base.Items {
		orderItems = append(orderItems, types.OrderItem{
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

	// 查询支付类型：通过 Redis 索引获取支付流水号，再查询支付详情
	payType := int64(0)
	payTypeDesc := ""
	paymentSn, _ := l.svcCtx.Client.Get(l.ctx, fmt.Sprintf("payment:order:%s", req.OrderSn)).Result()
	if paymentSn != "" {
		detail, err := l.svcCtx.PaymentRpc.GetPaymentDetail(l.ctx, &payment.GetPaymentDetailRequest{
			PaymentSn: paymentSn,
		})
		if err == nil && detail != nil {
			payType = int64(detail.PayType)
			payTypeDesc = constants.PayTypeMap[int(payType)]
		}
	}

	resp = &types.OrderDetailResp{
		OrderSn:     orderDetail.Base.OrderSn,
		Status:      int64(orderDetail.Base.Status),
		StatusDesc:  constants.OrderStatusMap[int(orderDetail.Base.Status)],
		PayType:     payType,
		PayTypeDesc: payTypeDesc,
		PayAmount:   orderDetail.Base.PayAmount,
		CreateTime:  orderDetail.Base.CreateTime.AsTime().Format(time.RFC3339),
		Items:       orderItems,
		Address: types.AddressSnapshot{
			ReceiverName:  orderDetail.Address.ReceiverName,
			ReceiverPhone: orderDetail.Address.ReceiverPhone,
			Detail: types.AddressDetail{
				Province:      orderDetail.Address.Detail.Province,
				City:          orderDetail.Address.Detail.City,
				District:      orderDetail.Address.Detail.District,
				DetailAddress: orderDetail.Address.Detail.DetailAddress,
				PostalCode:    orderDetail.Address.Detail.PostalCode,
			},
		},
		FreightAmount: orderDetail.FreightAmount,
		CouponAmount:  orderDetail.CouponAmount,
		DeliverySn:    orderDetail.DeliverySn,
		DeliveryCorp:  orderDetail.DeliveryCorp,
		DeliveryTime:  orderDetail.DeliveryTime.AsTime().Format(time.RFC3339),
		ExpireTime:    orderDetail.ExpireTime.AsTime().Unix(),
		Remark:        orderDetail.Remark,
		PayTime:       orderDetail.PayTime.AsTime().Format(time.RFC3339),
		FinishTime:    orderDetail.FinishTime.AsTime().Format(time.RFC3339),
		CancelTime:    orderDetail.CancelTime.AsTime().Format(time.RFC3339),
	}
	return resp, nil
}
