package orderadminlogic

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/client/userinternal"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

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

// GetOrder - 订单详情查询
func (l *GetOrderLogic) GetOrder(in *order.AdminGetOrderRequest) (*order.AdminOrderDetailResponse, error) {
	if in.OrderSn == "" {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "订单号不能为空")
	}

	orderInfo, err := l.svcCtx.OrderInfoModel.FindOneByOrderSn(l.ctx, in.OrderSn)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("订单不存在,order_sn=%s", in.OrderSn)
			return nil, errorx.NewBizError(response.ErrCodeOrderNotFound, "订单不存在")
		}
		l.Logger.Errorf("查询订单失败,error=%v", err)
		return nil, err
	}

	orderItems, err := l.svcCtx.OrderItemModel.FindListByOrderId(l.ctx, orderInfo.Id)
	if err != nil {
		l.Logger.Errorf("查询订单项失败,error=%v", err)
		return nil, err
	}

	var address userinternal.AddressItem
	err = json.Unmarshal([]byte(orderInfo.AddressSnap), &address)
	if err != nil {
		l.Logger.Errorf("解析地址快照失败,error=%v", err)
		return nil, err
	}

	userResp, err := l.svcCtx.UserRpc.GetUserById(l.ctx, &userinternal.GetUserByIdReq{UserId: orderInfo.UserId})
	if err != nil {
		l.Logger.Errorf("查询用户信息失败,userId=%d,error=%v", orderInfo.UserId, err)
		return nil, err
	}

	payTypeDesc := constants.PayTypeMap[int(orderInfo.PayType)]

	productItems := make([]*order.AdminOrderProductItem, 0, len(orderItems))
	for _, item := range orderItems {
		productItems = append(productItems, &order.AdminOrderProductItem{
			SkuId:       item.SkuId,
			SpuId:       item.SpuId,
			ProductName: item.SpuName,
			SkuName:     item.SkuName,
			Pic:         item.SkuPic,
			Price:       item.Price,
			Quantity:    item.Count,
			TotalAmount: item.TotalAmount,
		})
	}

	resp := &order.AdminOrderDetailResponse{
		OrderSn:       orderInfo.OrderSn,
		UserId:        orderInfo.UserId,
		UserPhone:     userResp.Phone,
		Status:        order.OrderStatus(orderInfo.Status),
		StatusDesc:    constants.OrderStatusMap[int(orderInfo.Status)],
		PayType:       order.PayType(orderInfo.PayType),
		PayTypeDesc:   payTypeDesc,
		PayAmount:     orderInfo.PayAmount,
		TotalAmount:   orderInfo.TotalAmount,
		FreightAmount: orderInfo.FreightAmount,
		CouponAmount:  orderInfo.CouponDiscount,
		Remark:        orderInfo.Remark,
		Address: &order.AdminOrderAddress{
			ReceiverName:  address.ReceiverName,
			ReceiverPhone: address.ReceiverPhone,
			Province:      address.Address.Province,
			City:          address.Address.City,
			District:      address.Address.District,
			DetailAddress: address.Address.Detail,
		},
		Items:           productItems,
		DeliverySn:      orderInfo.DeliverySn,
		DeliveryCompany: orderInfo.DeliveryCorp,
		CreateTime:      timestamppb.New(orderInfo.CreatedAt),
	}

	if orderInfo.PayTime.Valid {
		resp.PayTime = timestamppb.New(orderInfo.PayTime.Time)
	}
	if orderInfo.DeliveryTime.Valid {
		resp.DeliveryTime = timestamppb.New(orderInfo.DeliveryTime.Time)
	}
	if orderInfo.ReceiveTime.Valid {
		resp.FinishTime = timestamppb.New(orderInfo.ReceiveTime.Time)
	}
	if orderInfo.CancelTime.Valid {
		resp.CancelTime = timestamppb.New(orderInfo.CancelTime.Time)
	}

	l.Logger.Infof("查询订单详情成功,order_sn=%s", in.OrderSn)
	return resp, nil
}
