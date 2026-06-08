package orderlogic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelAfterSaleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelAfterSaleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelAfterSaleLogic {
	return &CancelAfterSaleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelAfterSaleLogic) CancelAfterSale(in *order.CancelAfterSaleRequest) (*order.Empty, error) {
	// 查询用户信息
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败,error=%v", err)
		return nil, err
	}
	afterSaleInfo, err := l.svcCtx.AfterSaleModel.FindOne(l.ctx, in.AfterSaleId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("售后单号不存在,AfterSaleId=%v", in.AfterSaleId)
			return nil, errorx.NewBizError(response.ErrCodeAfterSaleNotFound, "售后单号不存在")
		}
		l.Logger.Errorf("查询售后信息失败,error=%v", err)
		return nil, err
	}

	// 校验用户权限
	if afterSaleInfo.UserId != userId {
		l.Logger.Errorf("用户无权限取消售后单,AfterSaleId=%v", in.AfterSaleId)
		return nil, errorx.NewBizError(response.ErrCodePermissionDenied, "用户无权限取消售后单")
	}

	// 校验售后状态
	if afterSaleInfo.Status != constants.AFTER_SALE_STATUS_PENDING {
		l.Logger.Infof("售后单状态错误,AfterSaleId=%v,status=%v", in.AfterSaleId, afterSaleInfo.Status)
		return nil, errorx.NewBizError(response.ErrCodeAfterSaleStatusInvalid, "售后单状态不允许操作")
	}

	// 查找关联订单
	orderInfo, err := l.svcCtx.OrderInfoModel.FindOne(l.ctx, afterSaleInfo.OrderId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("订单不存在, OrderId=%v", afterSaleInfo.OrderId)
			return nil, errorx.NewBizError(response.ErrCodeOrderNotFound, "订单不存在")
		}
		l.Logger.Errorf("查询订单失败, error=%v", err)
		return nil, err
	}

	// 检查订单状态
	if orderInfo.Status != constants.ORDER_STATUS_AFTER_SALE {
		l.Logger.Errorf("订单非售后中状态,无法取消,OrderId=%v,Status=%v", orderInfo.Id, orderInfo.Status)
		return nil, errorx.NewBizError(response.ErrCodeOrderStatusInvalid, "订单状态异常，无法取消售后")
	}

	// 更新售后单状态
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		afterSaleInfo.Status = constants.AFTER_SALE_STATUS_CANCELED
		afterSaleInfo.CancelTime = sql.NullTime{Time: time.Now(), Valid: true}
		err = l.svcCtx.AfterSaleModel.UpdateTx(ctx, session, afterSaleInfo)
		if err != nil {
			l.Logger.Errorf("更新售后单失败,error=%v", err)
			return err
		}
		if orderInfo.ReceiveTime.Valid {
			orderInfo.Status = constants.ORDER_STATUS_COMPLETED
		} else if orderInfo.DeliveryTime.Valid {
			orderInfo.Status = constants.ORDER_STATUS_SHIPPED
		} else {
			orderInfo.Status = constants.ORDER_STATUS_PAID
		}
		err = l.svcCtx.OrderInfoModel.UpdateTx(ctx, session, orderInfo)
		if err != nil {
			l.Logger.Errorf("更新订单失败,error=%v", err)
			return err
		}
		return nil
	})
	if err != nil {
		l.Logger.Errorf("取消售后事务执行失败,error=%v", err)
		return nil, err
	}
	l.Logger.Info("更新售后单成功")
	return &order.Empty{}, nil
}
