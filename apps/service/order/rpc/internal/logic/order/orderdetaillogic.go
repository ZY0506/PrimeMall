package orderlogic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderDetailLogic {
	return &OrderDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OrderDetailLogic) OrderDetail(in *order.OrderDetailRequest) (*order.OrderDetailResponse, error) {
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败,error=%v", err)
		return nil, err
	}
	// 查找订单信息
	orderInfo, err := l.svcCtx.OrderInfoModel.FindOneByOrderSn(l.ctx, in.OrderSn)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("订单号：%s 不存在", in.OrderSn)
			return nil, errorx.NewBizError(response.ErrCodeOrderNotFound, "订单不存在")
		}
		l.Logger.Errorf("获取订单信息失败,error=%v", err)
		return nil, err
	}
	if orderInfo.UserId != userId {
		l.Logger.Errorf("用户ID不匹配,error=%v", err)
		return nil, errorx.NewBizError(response.ErrCodePermissionDenied, "用户无权限操作")
	}

	// 懒检查：如果订单是待支付且已过期，自动取消
	if orderInfo.Status == constants.ORDER_STATUS_PENDING_PAY && orderInfo.ExpireTime.Valid && time.Now().After(orderInfo.ExpireTime.Time) {
		l.Infof("懒检查：订单已过期，自动取消, order_sn=%s", in.OrderSn)
		if l.svcCtx.AutoCancelExpiredOrder(l.ctx, in.OrderSn) {
			// 刷新订单状态
			orderInfo.Status = constants.ORDER_STATUS_CANCELED
		}
	}

	// 查找订单项
	orderItems, err := l.svcCtx.OrderItemModel.FindListByOrderId(l.ctx, orderInfo.Id)
	if err != nil {
		l.Logger.Errorf("获取订单项失败,error=%v", err)
		return nil, err
	}
	// 反序列化地址快照
	addrSnap := order.AddressSnapshot{}
	err = json.Unmarshal([]byte(orderInfo.AddressSnap), &addrSnap)
	if err != nil {
		l.Logger.Errorf("反序列化地址快照失败,error=%v", err)
		return nil, err
	}
	// 待支付时间
	var expireTime *timestamppb.Timestamp
	if orderInfo.ExpireTime.Valid {
		expireTime = timestamppb.New(orderInfo.ExpireTime.Time)
	}
	// 支付时间
	var payTime *timestamppb.Timestamp
	if orderInfo.PayTime.Valid {
		payTime = timestamppb.New(orderInfo.PayTime.Time)
	}
	// 发货时间
	var deliveryTime *timestamppb.Timestamp
	if orderInfo.DeliveryTime.Valid {
		deliveryTime = timestamppb.New(orderInfo.DeliveryTime.Time)
	}
	// 取消时间
	var cancelTime *timestamppb.Timestamp
	if orderInfo.CancelTime.Valid {
		cancelTime = timestamppb.New(orderInfo.CancelTime.Time)
	}
	// 完成时间
	var finishTime *timestamppb.Timestamp
	if orderInfo.ReceiveTime.Valid {
		finishTime = timestamppb.New(orderInfo.ReceiveTime.Time)
	}
	var items []*order.OrderItem
	for _, v := range orderItems {
		items = append(items, &order.OrderItem{
			SkuId:       v.SkuId,
			SpuId:       v.SpuId,
			ProductName: v.SpuName,
			SkuName:     v.SkuName,
			Pic:         v.SkuPic,
			Price:       v.Price,
			Quantity:    v.Count,
			TotalAmount: v.TotalAmount,
		})
	}
	l.Logger.Info("获取订单详情成功")
	return &order.OrderDetailResponse{
		Base: &order.OrderListItem{
			OrderSn:    orderInfo.OrderSn,
			Status:     order.OrderStatus(orderInfo.Status),
			PayAmount:  orderInfo.PayAmount,
			CreateTime: timestamppb.New(orderInfo.CreatedAt),
			Items:      items,
		},
		Address: &order.AddressSnapshot{
			ReceiverName:  addrSnap.ReceiverName,
			ReceiverPhone: addrSnap.ReceiverPhone,
			Detail: &order.AddressDetail{
				Province:      addrSnap.Detail.Province,
				City:          addrSnap.Detail.City,
				District:      addrSnap.Detail.District,
				DetailAddress: addrSnap.Detail.DetailAddress,
				PostalCode:    addrSnap.Detail.PostalCode,
			},
		},
		FreightAmount: orderInfo.FreightAmount,
		CouponAmount:  orderInfo.CouponDiscount,
		DeliverySn:    orderInfo.DeliverySn,
		DeliveryCorp:  orderInfo.DeliveryCorp,
		ExpireTime:    expireTime,
		Remark:        orderInfo.Remark,
		PayTime:       payTime,
		DeliveryTime:  deliveryTime,
		CancelTime:    cancelTime,
		FinishTime:    finishTime,
	}, nil
}
