package orderlogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
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

type ApplyAfterSaleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApplyAfterSaleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyAfterSaleLogic {
	return &ApplyAfterSaleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ApplyAfterSale 申请售后
func (l *ApplyAfterSaleLogic) ApplyAfterSale(in *order.AfterSaleRequest) (resp *order.Empty, err error) {
	var ok bool
	// 幂等校验
	idempotencyKey := fmt.Sprintf("%s%s", constants.IDEMPOTENCY_KEY+constants.ORDER_SERVICE, in.IdempotencyKey)
	ok, err = l.svcCtx.Client.SetNX(l.ctx, idempotencyKey, "1", constants.IDEMPOTENCY_EXIRE).Result()
	if err != nil {
		l.Logger.Errorf("幂等性校验失败,error=%v", err)
		return nil, err
	}
	if !ok {
		return nil, errorx.NewBizError(response.ErrCodeTooFrequent, "请勿重复操作")
	}
	// TODO: 若后续失败，删除键（需确保删除操作的原子性，可配合 Lua 脚本）
	defer func() {
		if err != nil {
			l.svcCtx.Client.Del(l.ctx, idempotencyKey)
		}
	}()

	// 查询用户信息
	var userId uint64
	userId, err = ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败,error=%v", err)
		return nil, err
	}
	// 检查是否存在售后订单
	var afterSaleInfo *model.AfterSale
	afterSaleInfo, err = l.svcCtx.AfterSaleModel.FindOneByUserIdIdempotencyKey(l.ctx, userId, in.IdempotencyKey)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		l.Logger.Errorf("查询售后单信息失败，error=%v", err)
		return nil, err
	}
	// 如果找到了未完成的售后单，可在此处返回错误提示
	if afterSaleInfo != nil {
		return nil, errorx.NewBizError(response.ErrCodeAfterSaleAlreadyExist, "该商品已申请过售后")
	}
	unfinishedStatus := []int64{
		constants.AFTER_SALE_STATUS_PENDING,
		constants.AFTER_SALE_STATUS_RETURN,
		constants.AFTER_SALE_STATUS_REFUNDING,
	}
	afterSaleInfo, err = l.svcCtx.AfterSaleModel.FindOneByOrderSnSkuId(l.ctx, in.OrderSn, in.SkuId, unfinishedStatus)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		l.Logger.Errorf("查询售后单信息失败，error=%v", err)
		return nil, err
	}
	if afterSaleInfo != nil {
		return nil, errorx.NewBizError(response.ErrCodeAfterSaleAlreadyExist, "该商品已申请过售后")
	}

	// 参数校验
	if in.Reason == "" {
		return nil, errorx.NewBizError(response.ErrCodeAfterSaleReasonInvalid, "售后原因无效")
	}

	// 获取订单信息
	var orderInfo *model.OrderInfo
	orderInfo, err = l.svcCtx.OrderInfoModel.FindOneByOrderSn(l.ctx, in.OrderSn)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("订单号：%s 不存在", in.OrderSn)
			return nil, errorx.NewBizError(response.ErrCodeOrderNotFound, "订单不存在")
		}
		l.Logger.Errorf("获取订单信息失败，error=%v", err)
		return nil, err
	}
	// 订单状态校验
	allowedStatus := []int64{
		constants.ORDER_STATUS_PAID,
		constants.ORDER_STATUS_SHIPPED,
		constants.ORDER_STATUS_COMPLETED,
	}
	if !slices.Contains(allowedStatus, orderInfo.Status) {
		return nil, errorx.NewBizError(response.ErrCodeOrderStatusInvalid, "当前订单状态不可申请售后")
	}
	// 获取订单项
	var orderItem *model.OrderItem
	orderItem, err = l.svcCtx.OrderItemModel.FindOneByOrderSnSkuId(l.ctx, in.OrderSn, in.SkuId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("商品项不存在，orderSn=%s,skuId=%d", in.OrderSn, in.SkuId)
			return nil, errorx.NewBizError(response.ErrCodeSkuNotFound, "商品项不存在")
		}
		l.Logger.Errorf("获取订单项信息失败，error=%v", err)
		return nil, err
	}

	// 校验退款金额
	if in.ApplyAmount <= 0 || in.ApplyAmount > in.Quantity*orderItem.Price {
		return nil, errorx.NewBizError(response.ErrCodeAmountInvalid, "金额无效")
	}
	if in.Quantity > orderItem.Count {
		return nil, errorx.NewBizError(response.ErrCodeQuantityInvalid, "金额无效")
	}

	// 生成售后单号
	var afterSaleSn string
	afterSaleSn, err = l.svcCtx.IDGenerator.GenWithPrefix(constants.PREFIX_AFTER_SALE_SN)
	if err != nil {
		l.Logger.Errorf("生成售后单号失败，error=%v", err)
		return nil, err
	}
	var imageJson []byte
	if len(in.Images) != 0 {
		imageJson, err = json.Marshal(in.Images)
		if err != nil {
			l.Logger.Errorf("序列化图片列表失败，error=%v", err)
			return nil, err
		}
	}
	var afterSaleType int64
	if int64(in.Type) == 1 {
		afterSaleType = constants.AFTER_SALE_TYPE_REFUND_ONLY
	} else {
		afterSaleType = constants.AFTER_SALE_TYPE_RETURN_REFUND
	}
	afterSale := &model.AfterSale{
		AfterSaleSn:    afterSaleSn,
		IdempotencyKey: in.IdempotencyKey,
		OrderId:        orderInfo.Id,
		OrderSn:        in.OrderSn,
		UserId:         userId,
		Type:           afterSaleType,
		Status:         constants.AFTER_SALE_STATUS_PENDING,
		RefundStatus:   constants.REFUND_STATUS_PENDING_REFUND,
		ApplyAmount:    in.ApplyAmount,
		Reason:         in.Reason,
		Images:         sql.NullString{String: string(imageJson), Valid: true},
	}
	afterSaleItem := &model.AfterSaleItem{
		AfterSaleSn: afterSaleSn,
		OrderItemId: orderItem.Id,
		SkuId:       in.SkuId,
		SpuId:       orderItem.SpuId,
		ProductName: orderItem.SpuName,
		SkuName:     orderItem.SkuName,
		SkuPic:      orderItem.SkuPic,
		Price:       orderItem.Price,
		Quantity:    in.Quantity,
		TotalAmount: in.Quantity * orderItem.Price,
		CreatedAt:   time.Now(),
	}
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		_, err = l.svcCtx.AfterSaleModel.InsertTx(ctx, session, afterSale)
		if err != nil {
			logx.WithContext(ctx).Errorf("插入售后单信息失败，error=%v", err)
			return err
		}
		_, err = l.svcCtx.AfterSaleItemModel.InsertTx(ctx, session, afterSaleItem)
		if err != nil {
			logx.WithContext(ctx).Errorf("插入售后信息项失败，error=%v", err)
			return err
		}
		// 修改订单状态
		orderInfo.Status = constants.ORDER_STATUS_AFTER_SALE
		err = l.svcCtx.OrderInfoModel.UpdateTx(ctx, session, orderInfo)
		if err != nil {
			logx.WithContext(ctx).Errorf("更新订单状态失败，error=%v", err)
			return err
		}
		return nil
	})
	if err != nil {
		l.Logger.Error("申请售后失败")
		return nil, err
	}
	l.Logger.Info("申请售后成功")

	return &order.Empty{}, nil
}
