package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
	"time"
)

var _ ProductSkuModel = (*customProductSkuModel)(nil)

type (
	// ProductSkuModel is an interface to be customized, add more methods here,
	// and implement the added methods in customProductSkuModel.
	ProductSkuModel interface {
		productSkuModel
		customProductSku
		withSession(session sqlx.Session) ProductSkuModel
	}

	customProductSkuModel struct {
		*defaultProductSkuModel
	}
	customProductSku interface {
		FindByIds(ctx context.Context, spuIds []uint64) (*[]ProductSkuWithStatus, error)
		FindListBySpuId(ctx context.Context, spuId uint64) (*[]ProductSku, error)
		FindHotSkus(ctx context.Context, limit int) (*[]ProductSku, error)
		LockStock(ctx context.Context, items []*product.SkuStockItem, orderSn string) (*[]*product.SkuStockResult, error)
		UnlockStock(ctx context.Context, items []*product.SkuStockItem, orderSn string) (*[]*product.SkuStockResult, error)
		RollbackStock(ctx context.Context, items []*product.SkuStockItem, orderSn string) (*[]*product.SkuStockResult, error)
		DeductStock(ctx context.Context, items []*product.SkuStockItem, orderSn string) (*[]*product.SkuStockResult, error)
		RevertDeduct(ctx context.Context, items []*product.SkuStockItem, orderSn string) (*[]*product.SkuStockResult, error)
		CalculateSkusPrice(ctx context.Context, items []*product.SkuStockItem) (int64, error)
		InsertTx(ctx context.Context, session sqlx.Session, data *ProductSku) (sql.Result, error)
		BatchInsertTx(ctx context.Context, session sqlx.Session, data []*ProductSku) error
		UpdateTx(ctx context.Context, session sqlx.Session, data *ProductSku) error
		DeleteTx(ctx context.Context, session sqlx.Session, id uint64) error
	}
	ProductSkuWithStatus struct {
		ProductSku
		SpuStatus int64
	}
)

// NewProductSkuModel returns a model for the database table.
func NewProductSkuModel(conn sqlx.SqlConn) ProductSkuModel {
	return &customProductSkuModel{
		defaultProductSkuModel: newProductSkuModel(conn),
	}
}

func (m *customProductSkuModel) withSession(session sqlx.Session) ProductSkuModel {
	return NewProductSkuModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customProductSkuModel) InsertTx(ctx context.Context, session sqlx.Session, data *ProductSku) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, productSkuRowsExpectAutoSet)
	ret, err := session.ExecCtx(ctx, query, data.SpuId, data.SpuName, data.SkuCode, data.Price, data.MarketPrice, data.CostPrice, data.Stock, data.LockedStock, data.Version, data.SpecData, data.Images, data.Weight, data.Status, data.DeletedAt)
	return ret, err
}

func (m *customProductSkuModel) BatchInsertTx(ctx context.Context, session sqlx.Session, data []*ProductSku) error {
	if len(data) == 0 {
		return nil
	}

	query := fmt.Sprintf("insert into %s (%s) values ", m.table, productSkuRowsExpectAutoSet)
	var args []interface{}
	valueStrings := make([]string, 0, len(data))

	for _, item := range data {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		args = append(args, item.SpuId, item.SpuName, item.SkuCode, item.Price, item.MarketPrice, item.CostPrice, item.Stock, item.LockedStock, item.Version, item.SpecData, item.Images, item.Weight, item.Status, item.DeletedAt)
	}

	query += strings.Join(valueStrings, ", ")
	_, err := session.ExecCtx(ctx, query, args...)
	return err
}

func (m *customProductSkuModel) UpdateTx(ctx context.Context, session sqlx.Session, data *ProductSku) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, productSkuRowsWithPlaceHolder)
	_, err := session.ExecCtx(ctx, query, data.SpuId, data.SpuName, data.SkuCode, data.Price, data.MarketPrice, data.CostPrice, data.Stock, data.LockedStock, data.Version, data.SpecData, data.Images, data.Weight, data.Status, data.DeletedAt, data.Id)
	return err
}

func (m *customProductSkuModel) DeleteTx(ctx context.Context, session sqlx.Session, id uint64) error {
	query := fmt.Sprintf("update %s set deleted_at = ? where `id` = ?", m.table)
	_, err := session.ExecCtx(ctx, query, time.Now(), id)
	return err
}

func (m *defaultProductSkuModel) FindByIds(ctx context.Context, spuIds []uint64) (*[]ProductSkuWithStatus, error) {
	if len(spuIds) == 0 {
		return &[]ProductSkuWithStatus{}, nil
	}
	// 生成占位符
	placeholders := strings.Repeat("?,", len(spuIds))
	placeholders = placeholders[:len(placeholders)-1] // 去掉末尾多余的逗号

	// 修改查询：关联 spu 表 (假设表名为 product_spu，关联键为 spu_id)
	// 注意：必须给 sku 字段加前缀，避免与 spu 表的同名字段（如 id）冲突
	skuPrefixedRows := "sku." + strings.Join(productSkuFieldNames, ",sku.")
	query := fmt.Sprintf("select %s, spu.status as spu_status from %s sku join product_spu spu on sku.spu_id = spu.id where sku.status = 1 and sku.id in (%s)", skuPrefixedRows, m.table, placeholders)

	args := make([]interface{}, len(spuIds))
	for i, id := range spuIds {
		args[i] = id
	}

	var resp []ProductSkuWithStatus
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *defaultProductSkuModel) FindListBySpuId(ctx context.Context, spuId uint64) (*[]ProductSku, error) {
	query := fmt.Sprintf("select %s from %s where `spu_id` = ? and `status` = 1 and `deleted_at` IS NULL", productSkuRows, m.table)
	var resp []ProductSku
	err := m.conn.QueryRowsCtx(ctx, &resp, query, spuId)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// FindHotSkus 查询热销SKU（用于缓存预热）
func (m *defaultProductSkuModel) FindHotSkus(ctx context.Context, limit int) (*[]ProductSku, error) {
	query := fmt.Sprintf("select %s from %s where `status` = 1 and `deleted_at` IS NULL order by `stock` desc limit ?", productSkuRows, m.table)
	var resp []ProductSku
	err := m.conn.QueryRowsCtx(ctx, &resp, query, limit)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// LockStock 锁定库
// LockStock 锁定库存
func (m *defaultProductSkuModel) LockStock(ctx context.Context, items []*product.SkuStockItem, orderSn string) (*[]*product.SkuStockResult, error) {
	results := make([]*product.SkuStockResult, 0, len(items))

	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		anyFailed := false

		for _, item := range items {
			if anyFailed {
				break
			}
			result := &product.SkuStockResult{
				SkuId:   item.SkuId,
				Success: true,
			}
			var locked bool // 标记是否最终锁定了（成功）

			// 重试循环
			for retry := 0; retry < constants.MAX_RETRY_COUNT; retry++ {
				// ===================== 幂等校验 =====================
				exists, err := m.hasStockLog(ctx, session, orderSn, item.SkuId, constants.STOCK_CHANGE_TYPE_LOCKED)
				if err != nil {
					return err
				}
				if exists {
					result.Success = true
					result.Message = "已锁定（幂等）"
					locked = true
					logx.WithContext(ctx).Infof("库存已锁定（幂等）, orderSn=%s skuId=%d", orderSn, item.SkuId)
					break
				}
				var sku ProductSku
				queryOne := fmt.Sprintf("select %s from %s where `id` = ? and `status` = 1 and `deleted_at` IS NULL",
					productSkuRows, m.table)
				err = session.QueryRowCtx(ctx, &sku, queryOne, item.SkuId)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						logx.WithContext(ctx).Errorf("商品不存在, skuId=%d", item.SkuId)
						result.Success = false
						result.Message = "商品不存在"
						anyFailed = true
						locked = false
						break
					}
					logx.WithContext(ctx).Errorf("查询商品失败, error=%v", err)
					return err // 系统错误，直接返回（会触发整体回滚）
				}

				available := sku.Stock - sku.LockedStock
				if available < item.Quantity {
					logx.WithContext(ctx).Errorf("商品库存不足，skuId=%v", item.SkuId)
					result.Success = false
					result.Message = "商品库存不足"
					result.CurrentStock = sku.Stock
					result.CurrentLocked = sku.LockedStock
					anyFailed = true
					locked = false
					break
				}

				queryUpdate := fmt.Sprintf(`
                    UPDATE %s 
                    SET locked_stock = locked_stock + ?, version = version + 1 
                    WHERE id = ? AND version = ? AND (stock - locked_stock) >= ? 
                      AND status = 1 AND deleted_at IS NULL
                `, m.table)
				ret, err := session.ExecCtx(ctx, queryUpdate, item.Quantity, item.SkuId, sku.Version, item.Quantity)
				if err != nil {
					logx.WithContext(ctx).Errorf("更新商品库存失败, error=%v", err)
					return err
				}
				affected, _ := ret.RowsAffected()
				if affected > 0 {
					_, err = session.ExecCtx(ctx,
						`INSERT INTO stock_log 
                         (sku_id, order_sn, change_type, quantity, before_stock, after_stock, remark, created_at) 
                         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
						item.SkuId, orderSn, constants.STOCK_CHANGE_TYPE_LOCKED, item.Quantity,
						sku.Stock, sku.Stock, "用户下单，预扣减库存", time.Now(),
					)
					if err != nil {
						logx.WithContext(ctx).Errorf("记录库存日志失败, error=%v", err)
						return err
					}
					result.CurrentStock = sku.Stock
					result.CurrentLocked = sku.LockedStock + item.Quantity
					result.Message = "锁定成功"
					locked = true
					break
				}
				// 乐观锁冲突，重试
				logx.WithContext(ctx).Infof("乐观锁冲突，重试 %d/%d, skuId=%d", retry+1, constants.MAX_RETRY_COUNT, item.SkuId)
			}

			if !locked && result.Message == "" {
				logx.WithContext(ctx).Error("乐观锁冲突，重试失败")
				result.Success = false
				result.Message = "业务繁忙，请稍后重试"
				anyFailed = true
			}

			results = append(results, result)
		}

		// 如果有任何失败，整体回滚
		if anyFailed {
			for i := range results {
				if results[i].Success {
					results[i].Success = false
					results[i].Message = "因其他商品原因，下单失败"
				}
			}
			return errorx.NewBizError(response.ErrCodePreOrderFailed, "部分商品锁定失败，已整体回滚")
		}
		return nil
	})

	if err != nil {
		return &results, err
	}
	return &results, nil
}

// UnlockStock 解锁库存（订单超时、未支付取消订单）
func (m *defaultProductSkuModel) UnlockStock(ctx context.Context, items []*product.SkuStockItem, orderSn string) (*[]*product.SkuStockResult, error) {
	results := make([]*product.SkuStockResult, 0, len(items))

	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		anyFailed := false

		for _, item := range items {
			if anyFailed {
				// 只要有一个失败，后续的SKU不再处理（与LockStock保持一致）
				result := &product.SkuStockResult{
					SkuId:   item.SkuId,
					Success: false,
					Message: "因其他商品失败，整体回滚",
				}
				results = append(results, result)
				continue
			}

			result := &product.SkuStockResult{
				SkuId:   item.SkuId,
				Success: true,
			}
			var unlocked bool

			for retry := 0; retry < constants.MAX_RETRY_COUNT; retry++ {
				// ===================== 幂等校验 =====================
				exists, err := m.hasStockLog(ctx, session, orderSn, item.SkuId, constants.STOCK_CHANGE_TYPE_UNLOCKED)
				if err != nil {
					return err
				}
				if exists {
					result.Success = true
					result.Message = "已解锁（幂等）"
					unlocked = true
					logx.WithContext(ctx).Infof("库存已解锁（幂等）, orderSn=%s skuId=%d", orderSn, item.SkuId)
					break
				}
				// 1. 查询当前 SKU 最新数据
				var sku ProductSku
				queryOne := fmt.Sprintf("select %s from %s where `id` = ? and `status` = 1 and `deleted_at` IS NULL",
					productSkuRows, m.table)
				err = session.QueryRowCtx(ctx, &sku, queryOne, item.SkuId)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						logx.WithContext(ctx).Errorf("商品不存在, skuId=%d", item.SkuId)
						result.Success = false
						result.Message = "商品不存在"
						anyFailed = true
						unlocked = false
						break
					}
					logx.WithContext(ctx).Errorf("查询商品失败, error=%v", err)
					return err // 系统错误，直接回滚
				}

				// 2. 乐观锁更新：解锁 locked_stock
				queryUpdate := fmt.Sprintf(`
					UPDATE %s 
					SET locked_stock = locked_stock - ?, version = version + 1 
					WHERE id = ? AND version = ? AND locked_stock >= ?
					  AND status = 1 AND deleted_at IS NULL
				`, m.table)
				res, err := session.ExecCtx(ctx, queryUpdate, item.Quantity, item.SkuId, sku.Version, item.Quantity)
				if err != nil {
					logx.WithContext(ctx).Errorf("更新商品库存失败, error=%v", err)
					return err
				}
				affected, _ := res.RowsAffected()
				if affected > 0 {
					// 解锁成功，记录日志
					_, err = session.ExecCtx(ctx,
						`INSERT INTO stock_log 
                         (sku_id, order_sn, change_type, quantity, before_stock, after_stock, remark, created_at) 
                         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
						item.SkuId, orderSn, constants.STOCK_CHANGE_TYPE_UNLOCKED, item.Quantity,
						sku.Stock, sku.Stock, "订单超时，解锁库存", time.Now(),
					)
					if err != nil {
						logx.WithContext(ctx).Errorf("记录库存日志失败, error=%v", err)
						return err
					}
					result.CurrentStock = sku.Stock
					result.CurrentLocked = sku.LockedStock - item.Quantity
					result.Message = "解锁成功"
					unlocked = true
					break
				}
				// 乐观锁冲突，重试
				logx.WithContext(ctx).Infof("乐观锁冲突，重试 %d/%d, skuId=%d", retry+1, constants.MAX_RETRY_COUNT, item.SkuId)
			}

			if !unlocked && result.Message == "" {
				logx.WithContext(ctx).Errorf("乐观锁冲突，重试失败, skuId=%d", item.SkuId)
				result.Success = false
				result.Message = "业务繁忙，请稍后重试"
				anyFailed = true
			}

			results = append(results, result)
		}

		if anyFailed {
			// 将所有成功的结果标记为失败（整体回滚）
			for i := range results {
				if results[i].Success {
					results[i].Success = false
					results[i].Message = "因其他商品原因，操作失败"
				}
			}
			return errorx.NewBizError(response.ErrCodePreOrderFailed, "部分商品操作失败，已整体回滚")
		}
		return nil
	})

	if err != nil {
		return &results, err
	}
	return &results, nil
}

// RollbackStock 回滚库存（订单退款退货）
func (m *defaultProductSkuModel) RollbackStock(ctx context.Context, items []*product.SkuStockItem, orderSn string) (*[]*product.SkuStockResult, error) {
	results := make([]*product.SkuStockResult, 0, len(items))

	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		anyFailed := false

		for _, item := range items {
			if anyFailed {
				result := &product.SkuStockResult{
					SkuId:   item.SkuId,
					Success: false,
					Message: "因其他商品失败，整体回滚",
				}
				results = append(results, result)
				continue
			}

			result := &product.SkuStockResult{
				SkuId:   item.SkuId,
				Success: true,
			}
			var rolledBack bool

			for retry := 0; retry < constants.MAX_RETRY_COUNT; retry++ {
				var sku ProductSku
				queryOne := fmt.Sprintf("select %s from %s where `id` = ? and `status` = 1 and `deleted_at` IS NULL",
					productSkuRows, m.table)
				err := session.QueryRowCtx(ctx, &sku, queryOne, item.SkuId)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						logx.WithContext(ctx).Errorf("商品不存在, skuId=%d", item.SkuId)
						result.Success = false
						result.Message = "商品不存在"
						anyFailed = true
						rolledBack = false
						break
					}
					logx.WithContext(ctx).Errorf("查询商品失败, error=%v", err)
					return err
				}

				// 回滚操作：增加 stock（不修改 locked_stock，因为退款时货物已发货，locked_stock 应该是0或已扣减）
				// 注意：订单支付后，locked_stock 已减为0，这里只需回退 stock
				queryUpdate := fmt.Sprintf(`
                    UPDATE %s SET stock = stock + ?, version = version + 1 
                    WHERE id = ? AND version = ? AND status = 1 AND deleted_at IS NULL
                `, m.table)
				res, err := session.ExecCtx(ctx, queryUpdate, item.Quantity, item.SkuId, sku.Version)
				if err != nil {
					logx.WithContext(ctx).Errorf("更新商品库存失败, error=%v", err)
					return err
				}
				affected, _ := res.RowsAffected()
				if affected > 0 {
					// 回滚成功，记录日志
					_, err = session.ExecCtx(ctx,
						`INSERT INTO stock_log 
                         (sku_id, order_sn, change_type, quantity, before_stock, after_stock, remark, created_at) 
                         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
						item.SkuId, orderSn, constants.STOCK_CHANGE_TYPE_ROLLBACK, item.Quantity,
						sku.Stock, sku.Stock+item.Quantity, "订单取消/退款，库存回滚", time.Now(),
					)
					if err != nil {
						logx.WithContext(ctx).Errorf("记录库存日志失败, error=%v", err)
						return err
					}
					result.CurrentStock = sku.Stock + item.Quantity
					result.CurrentLocked = sku.LockedStock
					result.Message = "回滚成功"
					rolledBack = true
					break
				}
				logx.WithContext(ctx).Infof("乐观锁冲突，重试 %d/%d, skuId=%d", retry+1, constants.MAX_RETRY_COUNT, item.SkuId)
			}

			if !rolledBack && result.Message == "" {
				logx.WithContext(ctx).Errorf("库存回滚最终失败, skuId=%d", item.SkuId)
				result.Success = false
				result.Message = "业务繁忙，请稍后重试"
				anyFailed = true
			}

			results = append(results, result)
		}

		if anyFailed {
			for i := range results {
				if results[i].Success {
					results[i].Success = false
					results[i].Message = "因其他商品原因，操作失败"
				}
			}
			return errorx.NewBizError(response.ErrCodePreOrderFailed, "部分商品操作失败，已整体回滚")
		}
		return nil
	})

	if err != nil {
		return &results, err
	}
	return &results, nil
}

// DeductStock 扣减库存（订单支付/创建订单时）
func (m *defaultProductSkuModel) DeductStock(ctx context.Context, items []*product.SkuStockItem, orderSn string) (*[]*product.SkuStockResult, error) {
	results := make([]*product.SkuStockResult, 0, len(items))

	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		anyFailed := false

		for _, item := range items {
			if anyFailed {
				result := &product.SkuStockResult{
					SkuId:   item.SkuId,
					Success: false,
					Message: "因其他商品失败，整体回滚",
				}
				results = append(results, result)
				continue
			}

			result := &product.SkuStockResult{
				SkuId:   item.SkuId,
				Success: true,
			}
			var deducted bool

			for retry := 0; retry < constants.MAX_RETRY_COUNT; retry++ {
				// ===================== 幂等校验 =====================
				exists, err := m.hasStockLog(ctx, session, orderSn, item.SkuId, constants.STOCK_CHANGE_TYPE_DEDUCT)
				if err != nil {
					return err
				}
				if exists {
					result.Success = true
					result.Message = "已扣减（幂等）"
					deducted = true
					logx.WithContext(ctx).Infof("库存已扣减（幂等）, orderSn=%s skuId=%d", orderSn, item.SkuId)
					break
				}
				var sku ProductSku
				queryOne := fmt.Sprintf("select %s from %s where `id` = ? and `status` = 1 and `deleted_at` IS NULL",
					productSkuRows, m.table)
				err = session.QueryRowCtx(ctx, &sku, queryOne, item.SkuId)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						logx.WithContext(ctx).Errorf("商品不存在, skuId=%d", item.SkuId)
						result.Success = false
						result.Message = "商品不存在"
						anyFailed = true
						deducted = false
						break
					}
					logx.WithContext(ctx).Errorf("查询商品失败, error=%v", err)
					return err
				}

				// 扣减操作：stock - quantity, locked_stock - quantity
				available := sku.Stock - sku.LockedStock
				if available < item.Quantity {
					logx.WithContext(ctx).Infof("商品库存不足, skuId=%d, quantity=%d, available=%d", item.SkuId, item.Quantity, available)
					result.Success = false
					result.Message = "商品库存不足"
					result.CurrentStock = sku.Stock
					result.CurrentLocked = sku.LockedStock
					anyFailed = true
					deducted = false
					break
				}

				queryUpdate := fmt.Sprintf(`
                    UPDATE %s SET stock = stock - ?, locked_stock = locked_stock - ?, version = version + 1 
                    WHERE id = ? AND version = ? AND locked_stock >= ?
                      AND status = 1 AND deleted_at IS NULL
                `, m.table)
				res, err := session.ExecCtx(ctx, queryUpdate, item.Quantity, item.Quantity, item.SkuId, sku.Version, item.Quantity)
				if err != nil {
					logx.WithContext(ctx).Errorf("更新商品库存失败, error=%v", err)
					return err
				}
				affected, _ := res.RowsAffected()
				if affected > 0 {
					// 扣减成功，记录日志
					_, err = session.ExecCtx(ctx,
						`INSERT INTO stock_log 
                         (sku_id, order_sn, change_type, quantity, before_stock, after_stock, remark, created_at) 
                         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
						item.SkuId, orderSn, constants.STOCK_CHANGE_TYPE_DEDUCT, item.Quantity,
						sku.Stock, sku.Stock-item.Quantity, "用户提交订单，扣减库存", time.Now(),
					)
					if err != nil {
						logx.WithContext(ctx).Errorf("记录库存日志失败, error=%v", err)
						return err
					}
					result.CurrentStock = sku.Stock - item.Quantity
					result.CurrentLocked = sku.LockedStock - item.Quantity
					result.Message = "扣减成功"
					deducted = true
					break
				}
				logx.WithContext(ctx).Infof("乐观锁冲突，重试 %d/%d, skuId=%d", retry+1, constants.MAX_RETRY_COUNT, item.SkuId)
			}

			if !deducted && result.Message == "" {
				logx.WithContext(ctx).Errorf("扣减库存最终失败, skuId=%d", item.SkuId)
				result.Success = false
				result.Message = "业务繁忙，请稍后重试"
				anyFailed = true
			}

			results = append(results, result)
		}

		if anyFailed {
			for i := range results {
				if results[i].Success {
					results[i].Success = false
					results[i].Message = "因其他商品原因，操作失败"
				}
			}
			return errorx.NewBizError(response.ErrCodePreOrderFailed, "部分商品操作失败，已整体回滚")
		}
		return nil
	})

	if err != nil {
		return &results, err
	}
	return &results, nil
}

// RevertDeduct 回滚 DeductStock 操作（创建订单事务失败时调用）
func (m *defaultProductSkuModel) RevertDeduct(ctx context.Context, items []*product.SkuStockItem, orderSn string) (*[]*product.SkuStockResult, error) {
	results := make([]*product.SkuStockResult, 0, len(items))

	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		anyFailed := false

		for _, item := range items {
			if anyFailed {
				results = append(results, &product.SkuStockResult{
					SkuId:   item.SkuId,
					Success: false,
					Message: "因其他商品失败，整体回滚",
				})
				continue
			}

			result := &product.SkuStockResult{SkuId: item.SkuId, Success: true}
			var reverted bool

			for retry := 0; retry < constants.MAX_RETRY_COUNT; retry++ {
				// 1. 查询当前 SKU
				var sku ProductSku
				queryOne := fmt.Sprintf("select %s from %s where `id` = ? and `status` = 1 and `deleted_at` IS NULL",
					productSkuRows, m.table)
				err := session.QueryRowCtx(ctx, &sku, queryOne, item.SkuId)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						result.Success = false
						result.Message = "商品不存在"
						anyFailed = true
						break
					}
					return err
				}

				// 2. 同时恢复 stock 和 locked_stock
				queryUpdate := fmt.Sprintf(`
                    UPDATE %s 
                    SET stock = stock + ?, locked_stock = locked_stock + ?, version = version + 1 
                    WHERE id = ? AND version = ? AND status = 1 AND deleted_at IS NULL
                `, m.table)
				res, err := session.ExecCtx(ctx, queryUpdate, item.Quantity, item.Quantity, item.SkuId, sku.Version)
				if err != nil {
					return err
				}
				affected, _ := res.RowsAffected()
				if affected > 0 {
					// 记录日志
					_, err = session.ExecCtx(ctx,
						`INSERT INTO stock_log 
                         (sku_id, order_sn, change_type, quantity, before_stock, after_stock, remark, created_at) 
                         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
						item.SkuId, orderSn, constants.STOCK_CHANGE_TYPE_REVERT_DEDUCT, item.Quantity,
						sku.Stock, sku.Stock+item.Quantity, "创建订单失败，回滚扣减", time.Now(),
					)
					if err != nil {
						return err
					}
					result.CurrentStock = sku.Stock + item.Quantity
					result.CurrentLocked = sku.LockedStock + item.Quantity
					result.Message = "回滚成功"
					reverted = true
					break
				}
				logx.Infof("乐观锁冲突，重试 %d/%d, skuId=%d", retry+1, constants.MAX_RETRY_COUNT, item.SkuId)
			}

			if !reverted && result.Message == "" {
				result.Success = false
				result.Message = "业务繁忙，请稍后重试"
				anyFailed = true
			}
			results = append(results, result)
		}

		if anyFailed {
			for i := range results {
				if results[i].Success {
					results[i].Success = false
					results[i].Message = "因其他商品原因，操作失败"
				}
			}
			return errorx.NewBizError(response.ErrCodePreOrderFailed, "部分商品回滚失败")
		}
		return nil
	})

	if err != nil {
		return &results, err
	}
	return &results, nil
}

// CalculateSkusPrice 计算价格
func (m *defaultProductSkuModel) CalculateSkusPrice(ctx context.Context, items []*product.SkuStockItem) (int64, error) {
	var total int64
	for _, item := range items {
		sku, err := m.FindOne(ctx, item.SkuId)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				msg := fmt.Sprintf("skuId=%d,商品不存在", item.SkuId)
				logx.WithContext(ctx).Errorf("商品不存在, skuId=%d", item.SkuId)
				return 0, errorx.NewBizError(response.ErrCodeSkuNotFound, msg)
			}
			logx.WithContext(ctx).Errorf("查询商品失败, error=%v", err)
			return 0, err

		}
		if sku != nil {
			total += sku.Price * item.Quantity
		}
	}
	return total, nil
}

func (m *defaultProductSkuModel) hasStockLog(ctx context.Context, session sqlx.Session, orderSn string, skuId uint64, changeType int64) (bool, error) {
	var count int64
	err := session.QueryRowCtx(ctx, &count,
		`SELECT COUNT(1) FROM stock_log WHERE order_sn = ? AND sku_id = ? AND change_type = ?`,
		orderSn, skuId, changeType,
	)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
