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

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmReceiptLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfirmReceiptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmReceiptLogic {
	return &ConfirmReceiptLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConfirmReceiptLogic) ConfirmReceipt(in *order.OrderDetailRequest) (*order.Empty, error) {
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

	// 校验订单状态
	if orderInfo.Status != constants.ORDER_STATUS_SHIPPED {
		l.Logger.Errorf("订单状态错误,error=%v", err)
		return nil, errorx.NewBizError(response.ErrCodeOrderStatusInvalid, "订单状态不允许当前操作")
	}

	// 更新订单状态和信息
	orderInfo.Status = constants.ORDER_STATUS_COMPLETED
	orderInfo.ReceiveTime = sql.NullTime{Time: time.Now(), Valid: true}
	err = l.svcCtx.OrderInfoModel.Update(l.ctx, orderInfo)
	if err != nil {
		l.Logger.Errorf("更新订单信息失败,error=%v", err)
		return nil, err
	}
	l.Logger.Info("确认收货成功")
	return &order.Empty{}, nil
}
