package productadminlogic

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AdjustStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdjustStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdjustStockLogic {
	return &AdjustStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdjustStockLogic) AdjustStock(in *product.AdjustStockReq) (*product.StockChangeResp, error) {
	if len(in.Items) == 0 {
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "请至少选择一个SKU进行调整")
	}
	if in.Reason == "" {
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "请填写调整原因")
	}

	results := make([]*product.SkuStockResult, 0, len(in.Items))

	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		skuModel := l.svcCtx.ProductSkuModel.(interface {
			withSession(session sqlx.Session) model.ProductSkuModel
		}).withSession(session)

		for _, item := range in.Items {
			if item.Quantity == 0 {
				results = append(results, &product.SkuStockResult{
					SkuId:   item.SkuId,
					Success: false,
					Message: "调整数量不能为0",
				})
				continue
			}

			sku, err := l.svcCtx.ProductSkuModel.FindOne(ctx, item.SkuId)
			if err != nil {
				if errors.Is(err, model.ErrNotFound) {
					results = append(results, &product.SkuStockResult{
						SkuId:   item.SkuId,
						Success: false,
						Message: "SKU不存在",
					})
					continue
				}
				logx.WithContext(ctx).Errorf("查询SKU失败, skuId=%d, err=%v", item.SkuId, err)
				return err
			}

			beforeStock := sku.Stock
			afterStock := beforeStock + item.Quantity

			if afterStock < 0 {
				results = append(results, &product.SkuStockResult{
					SkuId:         item.SkuId,
					Success:       false,
					Message:       fmt.Sprintf("库存不足，当前库存%d，无法减少%d", beforeStock, -item.Quantity),
					CurrentStock:  beforeStock,
					CurrentLocked: sku.LockedStock,
				})
				continue
			}

			sku.Stock = afterStock
			sku.Version = sku.Version + 1

			err = skuModel.UpdateTx(ctx, session, sku)
			if err != nil {
				logx.WithContext(ctx).Errorf("更新SKU库存失败, skuId=%d, err=%v", item.SkuId, err)
				results = append(results, &product.SkuStockResult{
					SkuId:         item.SkuId,
					Success:       false,
					Message:       "更新失败",
					CurrentStock:  beforeStock,
					CurrentLocked: sku.LockedStock,
				})
				continue
			}

			remark := fmt.Sprintf("管理员调整库存: %s。调整前: %d, 调整后: %d", in.Reason, beforeStock, afterStock)
			if in.Operator != "" {
				remark = fmt.Sprintf("操作人: %s。%s", in.Operator, remark)
			}

			stockLog := &model.StockLog{
				SkuId:       item.SkuId,
				OrderSn:     "",
				ChangeType:  constants.STOCK_CHANGE_TYPE_ADMIN,
				Quantity:    item.Quantity,
				BeforeStock: beforeStock,
				AfterStock:  afterStock,
				Remark:      remark,
				CreatedAt:   time.Now(),
			}

			_, err = l.svcCtx.StockLogModel.Insert(ctx, stockLog)
			if err != nil {
				logx.WithContext(ctx).Errorf("记录库存日志失败, skuId=%d, err=%v", item.SkuId, err)
				return err
			}

			results = append(results, &product.SkuStockResult{
				SkuId:         item.SkuId,
				Success:       true,
				Message:       "调整成功",
				CurrentStock:  afterStock,
				CurrentLocked: sku.LockedStock,
			})

			logx.WithContext(ctx).Infof("管理员调整库存成功, skuId=%d, operator=%s, before=%d, after=%d, quantity=%d",
				item.SkuId, in.Operator, beforeStock, afterStock, item.Quantity)
		}

		return nil
	})

	if err != nil {
		l.Logger.Errorf("调整库存失败（已回滚）, err=%v", err)
		return nil, err
	}

	anyFailed := false
	for _, result := range results {
		if !result.Success {
			anyFailed = true
			break
		}
	}
// 清除被调整SKU的库存缓存
	for _, item := range in.Items {
		stockKey := constants.ProductStockKey + strconv.FormatUint(item.SkuId, 10)
		l.svcCtx.Client.Del(l.ctx, stockKey)
	}

	return &product.StockChangeResp{
		Success: !anyFailed,
		Results: results,
	}, nil
}
