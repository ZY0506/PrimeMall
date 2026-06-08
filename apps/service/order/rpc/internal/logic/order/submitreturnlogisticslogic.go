package orderlogic

import (
	"context"
	"errors"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitReturnLogisticsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitReturnLogisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitReturnLogisticsLogic {
	return &SubmitReturnLogisticsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SubmitReturnLogisticsLogic) SubmitReturnLogistics(in *order.SubmitReturnLogisticsRequest) (*order.Empty, error) {
	// 1. 修复：简化用户ID获取（删除冗余声明）
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败,error=%v", err)
		return nil, err
	}

	// 2. 校验必填参数：物流单号/公司不能为空
	if in.TrackingNo == "" || in.LogisticsCompany == "" {
		l.Logger.Errorf("物流信息不能为空, AfterSaleId=%v", in.AfterSaleId)
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "物流单号、物流公司不能为空")
	}

	// 3. 查询售后单
	afterSaleInfo, err := l.svcCtx.AfterSaleModel.FindOne(l.ctx, in.AfterSaleId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("售后单号不存在,AfterSaleId=%v", in.AfterSaleId)
			return nil, errorx.NewBizError(response.ErrCodeAfterSaleNotFound, "售后单号不存在")
		}
		l.Logger.Errorf("查询售后信息失败,error=%v", err)
		return nil, err
	}

	// 4. 校验用户权限
	if afterSaleInfo.UserId != userId {
		l.Logger.Errorf("用户无权限提交售后物流信息,AfterSaleId=%v", in.AfterSaleId)
		return nil, errorx.NewBizError(response.ErrCodePermissionDenied, "用户无权限提交售后物流信息")
	}

	// 5. 校验售后类型：仅【退货退款(2)】可提交物流
	if afterSaleInfo.Type != constants.AFTER_SALE_TYPE_RETURN_REFUND {
		l.Logger.Errorf("仅退货退款售后单可提交物流,AfterSaleId=%v,type=%v", in.AfterSaleId, afterSaleInfo.Type)
		return nil, errorx.NewBizError(response.ErrCodeAfterSaleTypeInvalid, "仅退货退款订单可提交物流信息")
	}

	// 6. 状态校验
	if afterSaleInfo.Status != constants.AFTER_SALE_STATUS_RETURN {
		l.Logger.Errorf("售后单状态错误,仅待退货可提交物流,AfterSaleId=%v,status=%v", in.AfterSaleId, afterSaleInfo.Status)
		return nil, errorx.NewBizError(response.ErrCodeAfterSaleStatusInvalid, "仅待退货的售后单可提交物流")
	}

	// 7. 查询关联订单
	orderInfo, err := l.svcCtx.OrderInfoModel.FindOne(l.ctx, afterSaleInfo.OrderId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("订单不存在, OrderId=%v", afterSaleInfo.OrderId)
			return nil, errorx.NewBizError(response.ErrCodeOrderNotFound, "订单不存在")
		}
		l.Logger.Errorf("查询订单失败, error=%v", err)
		return nil, err
	}

	// 8. 校验订单状态
	if orderInfo.Status != constants.ORDER_STATUS_AFTER_SALE {
		l.Logger.Errorf("订单非售后中状态,无法提交售后物流信息,OrderId=%v,Status=%v", orderInfo.Id, orderInfo.Status)
		return nil, errorx.NewBizError(response.ErrCodeOrderStatusInvalid, "订单状态异常，无法提交售后物流信息")
	}

	afterSaleInfo.ReturnTrackingSn = in.TrackingNo
	afterSaleInfo.ReturnTrackingCorp = in.LogisticsCompany
	err = l.svcCtx.AfterSaleModel.Update(l.ctx, afterSaleInfo)
	if err != nil {
		l.Logger.Errorf("提交售后物流信息失败,error=%v", err)
		return nil, err
	}

	l.Logger.Infof("提交售后物流信息成功, AfterSaleId=%v", in.AfterSaleId)
	return &order.Empty{}, nil
}
