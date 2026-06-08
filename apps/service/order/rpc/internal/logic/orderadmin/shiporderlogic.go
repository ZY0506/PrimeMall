package orderadminlogic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShipOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewShipOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShipOrderLogic {
	return &ShipOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ShipOrder - 订单发货
func (l *ShipOrderLogic) ShipOrder(in *order.AdminShipOrderRequest) (*order.Empty, error) {
	if in.OrderSn == "" {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "订单号不能为空")
	}
	if in.DeliveryCompany == "" {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "物流公司不能为空")
	}
	if in.DeliverySn == "" {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "物流单号不能为空")
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

	if orderInfo.Status != constants.ORDER_STATUS_PAID {
		l.Logger.Errorf("订单状态不正确,无法发货,order_sn=%s,status=%d", in.OrderSn, orderInfo.Status)
		return nil, errorx.NewBizError(response.ErrCodeOrderStatusInvalid, "订单状态不正确，只有已支付的订单才能发货")
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		orderInfo.DeliverySn = in.DeliverySn
		orderInfo.DeliveryCorp = in.DeliveryCompany
		orderInfo.DeliveryTime = sql.NullTime{Time: time.Now(), Valid: true}
		orderInfo.Status = constants.ORDER_STATUS_SHIPPED

		err := l.svcCtx.OrderInfoModel.UpdateTx(ctx, session, orderInfo)
		if err != nil {
			l.Logger.Errorf("更新订单发货信息失败,error=%v", err)
			return err
		}

		return nil
	})

	if err != nil {
		l.Logger.Errorf("订单发货失败,error=%v", err)
		return nil, err
	}

	l.Logger.Infof("订单发货成功,order_sn=%s,delivery_company=%s,delivery_sn=%s", in.OrderSn, in.DeliveryCompany, in.DeliverySn)
	return &order.Empty{}, nil
}
