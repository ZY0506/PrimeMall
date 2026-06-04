package orderlogic

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type AfterSaleDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAfterSaleDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AfterSaleDetailLogic {
	return &AfterSaleDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AfterSaleDetailLogic) AfterSaleDetail(in *order.AfterSaleDetailRequest) (*order.AfterSaleDetailResponse, error) {
	// 1. 获取当前登录用户ID
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败, error=%v", err)
		return nil, err
	}

	// 2. 入参校验
	if in.AfterSaleId == 0 {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "售后单ID不能为空")
	}

	// 3. 查询售后主表
	afterSaleInfo, err := l.svcCtx.AfterSaleModel.FindOne(l.ctx, in.AfterSaleId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("售后单不存在, AfterSaleId=%v", in.AfterSaleId)
			return nil, errorx.NewBizError(response.ErrCodeAfterSaleNotFound, "售后单不存在")
		}
		l.Logger.Errorf("查询售后单失败, error=%v", err)
		return nil, err
	}

	// 4. 权限校验：仅能查看自己的售后单
	if afterSaleInfo.UserId != userId {
		l.Logger.Errorf("无权限查看他人售后单, AfterSaleId=%v", in.AfterSaleId)
		return nil, errorx.NewBizError(response.ErrCodePermissionDenied, "无权限查看该售后单")
	}

	// 5. 查询售后商品快照
	afterSaleItem, err := l.svcCtx.AfterSaleItemModel.FindOneByAfterSaleSn(l.ctx, afterSaleInfo.AfterSaleSn)
	if err != nil {
		l.Logger.Errorf("查询售后商品项失败, error=%v", err)
		return nil, err
	}

	// 6. 解析售后图片
	var images []string
	if afterSaleInfo.Images.Valid {
		_ = json.Unmarshal([]byte(afterSaleInfo.Images.String), &images)
	}

	// 7. 组装商品信息
	var productItem *order.OrderItem
	if afterSaleItem != nil {
		productItem = &order.OrderItem{
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

	// 8. 组装基础响应数据（严格匹配你的 Protobuf 结构体）
	resp := &order.AfterSaleDetailResponse{
		AfterSaleId:        afterSaleInfo.Id,
		OrderSn:            afterSaleInfo.OrderSn,
		Type:               order.AfterSaleType(afterSaleInfo.Type),
		Status:             order.AfterSaleStatus(afterSaleInfo.Status),
		StatusDesc:         constants.OrderStatusMap[int(afterSaleInfo.Status)],
		ApplyAmount:        afterSaleInfo.ApplyAmount,
		Reason:             afterSaleInfo.Reason,
		Images:             images,
		AuditRemark:        afterSaleInfo.AuditRemark,
		CreatedAt:          timestamppb.New(afterSaleInfo.CreatedAt),
		ReturnTrackingSn:   afterSaleInfo.ReturnTrackingSn,
		ReturnTrackingCorp: afterSaleInfo.ReturnTrackingCorp,
		RefundAmount:       afterSaleInfo.RealRefundAmount,
		ProductItem:        productItem,
	}

	// 9. 空安全时间字段处理
	// 审核时间
	if afterSaleInfo.AuditTime.Valid {
		resp.AuditTime = timestamppb.New(afterSaleInfo.AuditTime.Time)
	}
	// 商家收货时间
	if afterSaleInfo.ReturnReceivedTime.Valid {
		resp.ReceiveTime = timestamppb.New(afterSaleInfo.ReturnReceivedTime.Time)
	}
	// 退款完成时间
	if afterSaleInfo.RefundTime.Valid {
		resp.RefundTime = timestamppb.New(afterSaleInfo.RefundTime.Time)
	}

	l.Logger.Infof("查询售后详情成功, AfterSaleId=%d", in.AfterSaleId)
	return resp, nil
}
