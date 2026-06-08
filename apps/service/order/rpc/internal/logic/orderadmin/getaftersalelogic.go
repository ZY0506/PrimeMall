package orderadminlogic

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAfterSaleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAfterSaleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAfterSaleLogic {
	return &GetAfterSaleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetAfterSale - 售后详情查询
func (l *GetAfterSaleLogic) GetAfterSale(in *order.AdminGetAfterSaleRequest) (*order.AdminAfterSaleDetailResponse, error) {
	if in.AfterSaleId == 0 {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "售后单ID不能为空")
	}

	afterSaleInfo, err := l.svcCtx.AfterSaleModel.FindOne(l.ctx, in.AfterSaleId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("售后单不存在,after_sale_id=%d", in.AfterSaleId)
			return nil, errorx.NewBizError(response.ErrCodeAfterSaleNotFound, "售后单不存在")
		}
		l.Logger.Errorf("查询售后单失败,error=%v", err)
		return nil, err
	}

	afterSaleItem, err := l.svcCtx.AfterSaleItemModel.FindOneByAfterSaleSn(l.ctx, afterSaleInfo.AfterSaleSn)
	if err != nil {
		l.Logger.Errorf("查询售后商品项失败,error=%v", err)
		return nil, err
	}

	var images []string
	if afterSaleInfo.Images.Valid {
		_ = json.Unmarshal([]byte(afterSaleInfo.Images.String), &images)
	}

	var productItem *order.AdminOrderProductItem
	if afterSaleItem != nil {
		productItem = &order.AdminOrderProductItem{
			SkuId:       afterSaleItem.SkuId,
			SpuId:       afterSaleItem.SpuId,
			ProductName: afterSaleItem.ProductName,
			SkuName:     afterSaleItem.SkuName,
			Pic:         afterSaleItem.SkuPic,
			Price:       afterSaleItem.Price,
			Quantity:    afterSaleItem.Quantity,
			TotalAmount: afterSaleItem.TotalAmount,
		}
	}

	resp := &order.AdminAfterSaleDetailResponse{
		AfterSaleId:        afterSaleInfo.Id,
		OrderSn:            afterSaleInfo.OrderSn,
		Type:               order.AfterSaleType(afterSaleInfo.Type),
		Status:             order.AfterSaleStatus(afterSaleInfo.Status),
		StatusDesc:         constants.AfterSaleStatusMap[int(afterSaleInfo.Status)],
		ApplyAmount:        afterSaleInfo.ApplyAmount,
		Reason:             afterSaleInfo.Reason,
		Images:             images,
		AuditRemark:        afterSaleInfo.AuditRemark,
		ReturnTrackingSn:   afterSaleInfo.ReturnTrackingSn,
		ReturnTrackingCorp: afterSaleInfo.ReturnTrackingCorp,
		RefundAmount:       afterSaleInfo.RealRefundAmount,
		ProductItem:        productItem,
		CreatedAt:          timestamppb.New(afterSaleInfo.CreatedAt),
	}

	if afterSaleInfo.AuditTime.Valid {
		resp.AuditTime = timestamppb.New(afterSaleInfo.AuditTime.Time)
	}
	if afterSaleInfo.RefundTime.Valid {
		resp.RefundTime = timestamppb.New(afterSaleInfo.RefundTime.Time)
	}

	l.Logger.Infof("查询售后详情成功,after_sale_id=%d", in.AfterSaleId)
	return resp, nil
}
