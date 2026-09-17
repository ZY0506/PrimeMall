package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, productSkuRowsExpectAutoSet)
	ret, err := session.ExecCtx(ctx, query, data.SpuId, data.SpuName, data.SkuCode, data.Price, data.MarketPrice, data.CostPrice, data.Stock, data.LockedStock, data.SpecData, data.Images, data.Weight, data.Status, data.DeletedAt)
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
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		args = append(args, item.SpuId, item.SpuName, item.SkuCode, item.Price, item.MarketPrice, item.CostPrice, item.Stock, item.LockedStock, item.SpecData, item.Images, item.Weight, item.Status, item.DeletedAt)
	}

	query += strings.Join(valueStrings, ", ")
	_, err := session.ExecCtx(ctx, query, args...)
	return err
}

func (m *customProductSkuModel) UpdateTx(ctx context.Context, session sqlx.Session, data *ProductSku) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, productSkuRowsWithPlaceHolder)
	_, err := session.ExecCtx(ctx, query, data.SpuId, data.SpuName, data.SkuCode, data.Price, data.MarketPrice, data.CostPrice, data.Stock, data.LockedStock, data.SpecData, data.Images, data.Weight, data.Status, data.DeletedAt, data.Id)
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

// ==================== 库存变更 ====================
//
// 并发控制采用两层方案：
//  1. 应用层：RPC 入口按 skuId 加 Redis 分布式锁（见 productinternal/stocklock.go），
//     把同一 SKU 的并发请求挡在 DB 之前；
//  2. DB 层：单条带 WHERE 条件的原子 UPDATE 兜底。
//
// 不再使用 version 乐观锁：InnoDB 默认 REPEATABLE READ 下，事务内的普通 SELECT 是一致性
// 快照读，重试时读到的仍是事务开始时的旧 version，重试必然持续冲突，因此该机制无效。
// 条件原子 UPDATE 的 RowsAffected == 0 即表示"库存不足/已下架"，无需重试。

// stockChangeSpec 描述一次库存变更中随业务类型而异的参数。
//
// 5 个库存操作（锁定/解锁/回滚/扣减/恢复扣减）共用同一套
// 「幂等校验 → 快照读 → 条件原子更新 → 写流水」流程，差异全部收敛到这里。
type stockChangeSpec struct {
	// changeType 写入 stock_log.change_type
	changeType int64
	// idempotent 是否在更新前做 stock_log 幂等校验
	idempotent bool
	// setClause UPDATE 的 SET 子句（? 为占位符）
	setClause string
	// guardClause UPDATE 的 WHERE 附加条件（不含前导 AND），空串表示无附加条件
	guardClause string
	// updateArgs 组装 UPDATE 参数：SET 占位符参数 + skuId + guard 占位符参数
	updateArgs func(qty int64, skuId uint64) []interface{}
	// successMsg 更新成功时的提示
	successMsg string
	// depletedMsg 条件更新命中 0 行时的提示
	depletedMsg string
	// remark 写入 stock_log 的备注
	remark string
	// stockDelta 本次变更对 stock 列的增量
	stockDelta func(qty int64) int64
	// lockedDelta 本次变更对 locked_stock 列的增量
	lockedDelta func(qty int64) int64
	// abortMsg 整体回滚时返回的业务错误信息
	abortMsg string
	// flipMsg 整体回滚时改写成功结果的信息
	flipMsg string
}

// sortableStockItem 库存变更条目及其在原始请求中的下标
type sortableStockItem struct {
	item *product.SkuStockItem
	pos  int
}

// sortItemsBySkuId 按 SkuId 升序返回带原始下标的条目。
// 所有多 SKU 库存操作都按同一顺序获取行锁，避免交叉顺序导致的死锁。
func sortItemsBySkuId(items []*product.SkuStockItem) []sortableStockItem {
	sorted := make([]sortableStockItem, len(items))
	for i, item := range items {
		sorted[i] = sortableStockItem{item: item, pos: i}
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].item.SkuId < sorted[j].item.SkuId })
	return sorted
}

// applyStockChange 库存变更的统一实现。
// 返回的结果切片与入参 items 顺序一一对应。
func (m *defaultProductSkuModel) applyStockChange(ctx context.Context, items []*product.SkuStockItem, orderSn string, spec stockChangeSpec) (*[]*product.SkuStockResult, error) {
	results := make([]*product.SkuStockResult, len(items))

	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		anyFailed := false

		for _, cur := range sortItemsBySkuId(items) {
			result := &product.SkuStockResult{
				SkuId:   cur.item.SkuId,
				Success: true,
			}
			results[cur.pos] = result

			// 只要有一个 SKU 失败，后续的 SKU 不再处理（整体回滚）
			if anyFailed {
				result.Success = false
				result.Message = "因其他商品失败，整体回滚"
				continue
			}

			ok, err := m.applyOneStockChange(ctx, session, orderSn, cur.item, spec, result)
			if err != nil {
				return err
			}
			if !ok {
				anyFailed = true
			}
		}

		if anyFailed {
			// 将所有成功的结果标记为失败（整体回滚）
			for i := range results {
				if results[i].Success {
					results[i].Success = false
					results[i].Message = spec.flipMsg
				}
			}
			return errorx.NewBizError(response.ErrCodePreOrderFailed, spec.abortMsg)
		}
		return nil
	})

	if err != nil {
		return &results, err
	}
	return &results, nil
}

// applyOneStockChange 处理单个 SKU 的库存变更，返回 false 表示业务失败（需要整体回滚）。
// 返回的 error 仅用于系统级错误（会直接中断事务）。
func (m *defaultProductSkuModel) applyOneStockChange(ctx context.Context, session sqlx.Session, orderSn string, item *product.SkuStockItem, spec stockChangeSpec, result *product.SkuStockResult) (bool, error) {
	// ===================== 1. 幂等校验 =====================
	if spec.idempotent {
		exists, err := m.hasStockLog(ctx, session, orderSn, item.SkuId, spec.changeType)
		if err != nil {
			return false, err
		}
		if exists {
			result.Message = "已处理（幂等）"
			logx.WithContext(ctx).Infof("库存变更已处理（幂等）, orderSn=%s skuId=%d changeType=%d", orderSn, item.SkuId, spec.changeType)
			return true, nil
		}
	}

	// ===================== 2. 快照读：校验 SKU 有效 + 取流水所需值 =====================
	var sku ProductSku
	queryOne := fmt.Sprintf("select %s from %s where `id` = ? and `status` = 1 and `deleted_at` IS NULL",
		productSkuRows, m.table)
	if err := session.QueryRowCtx(ctx, &sku, queryOne, item.SkuId); err != nil {
		if errors.Is(err, ErrNotFound) {
			logx.WithContext(ctx).Errorf("商品不存在, skuId=%d", item.SkuId)
			result.Success = false
			result.Message = "商品不存在"
			return false, nil
		}
		logx.WithContext(ctx).Errorf("查询商品失败, error=%v", err)
		return false, err
	}

	// ===================== 3. 条件原子更新 =====================
	where := "id = ?"
	if spec.guardClause != "" {
		where += " AND " + spec.guardClause
	}
	queryUpdate := fmt.Sprintf("UPDATE %s SET %s WHERE %s AND status = 1 AND deleted_at IS NULL",
		m.table, spec.setClause, where)
	res, err := session.ExecCtx(ctx, queryUpdate, spec.updateArgs(item.Quantity, item.SkuId)...)
	if err != nil {
		logx.WithContext(ctx).Errorf("更新商品库存失败, error=%v", err)
		return false, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		logx.WithContext(ctx).Errorf("库存条件更新未命中（%s）, skuId=%d quantity=%d", spec.depletedMsg, item.SkuId, item.Quantity)
		result.Success = false
		result.Message = spec.depletedMsg
		result.CurrentStock = sku.Stock
		result.CurrentLocked = sku.LockedStock
		return false, nil
	}

	// ===================== 4. 写库存流水 =====================
	_, err = session.ExecCtx(ctx,
		`INSERT INTO stock_log
         (sku_id, order_sn, change_type, quantity, before_stock, after_stock, remark, created_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		item.SkuId, orderSn, spec.changeType, item.Quantity,
		sku.Stock, sku.Stock+spec.stockDelta(item.Quantity), spec.remark, time.Now(),
	)
	if err != nil {
		logx.WithContext(ctx).Errorf("记录库存日志失败, error=%v", err)
		return false, err
	}

	result.CurrentStock = sku.Stock + spec.stockDelta(item.Quantity)
	result.CurrentLocked = sku.LockedStock + spec.lockedDelta(item.Quantity)
	result.Message = spec.successMsg
	return true, nil
}

// LockStock 锁定库存（创建订单预占）
func (m *defaultProductSkuModel) LockStock(ctx context.Context, items []*product.SkuStockItem, orderSn string) (*[]*product.SkuStockResult, error) {
	return m.applyStockChange(ctx, items, orderSn, stockChangeSpec{
		changeType:  constants.STOCK_CHANGE_TYPE_LOCKED,
		idempotent:  true,
		setClause:   "locked_stock = locked_stock + ?",
		guardClause: "(stock - locked_stock) >= ?",
		updateArgs: func(qty int64, skuId uint64) []interface{} {
			return []interface{}{qty, skuId, qty}
		},
		successMsg:  "锁定成功",
		depletedMsg: "商品库存不足",
		remark:      "用户下单，预扣减库存",
		stockDelta:  func(int64) int64 { return 0 },
		lockedDelta: func(qty int64) int64 { return qty },
		abortMsg:    "部分商品锁定失败，已整体回滚",
		flipMsg:     "因其他商品原因，下单失败",
	})
}

// UnlockStock 解锁库存（订单超时、未支付取消订单）
func (m *defaultProductSkuModel) UnlockStock(ctx context.Context, items []*product.SkuStockItem, orderSn string) (*[]*product.SkuStockResult, error) {
	return m.applyStockChange(ctx, items, orderSn, stockChangeSpec{
		changeType:  constants.STOCK_CHANGE_TYPE_UNLOCKED,
		idempotent:  true,
		setClause:   "locked_stock = locked_stock - ?",
		guardClause: "locked_stock >= ?",
		updateArgs: func(qty int64, skuId uint64) []interface{} {
			return []interface{}{qty, skuId, qty}
		},
		successMsg:  "解锁成功",
		depletedMsg: "可解锁库存不足",
		remark:      "订单超时，解锁库存",
		stockDelta:  func(int64) int64 { return 0 },
		lockedDelta: func(qty int64) int64 { return -qty },
		abortMsg:    "部分商品操作失败，已整体回滚",
		flipMsg:     "因其他商品原因，操作失败",
	})
}

// RollbackStock 回滚库存（订单退款退货）
func (m *defaultProductSkuModel) RollbackStock(ctx context.Context, items []*product.SkuStockItem, orderSn string) (*[]*product.SkuStockResult, error) {
	return m.applyStockChange(ctx, items, orderSn, stockChangeSpec{
		changeType: constants.STOCK_CHANGE_TYPE_ROLLBACK,
		// 回滚只增加 stock，不修改 locked_stock：订单支付后 locked_stock 已扣减为 0
		setClause:   "stock = stock + ?",
		updateArgs:  func(qty int64, skuId uint64) []interface{} { return []interface{}{qty, skuId} },
		successMsg:  "回滚成功",
		depletedMsg: "商品不存在或已下架",
		remark:      "订单取消/退款，库存回滚",
		stockDelta:  func(qty int64) int64 { return qty },
		lockedDelta: func(int64) int64 { return 0 },
		abortMsg:    "部分商品操作失败，已整体回滚",
		flipMsg:     "因其他商品原因，操作失败",
	})
}

// DeductStock 扣减库存（订单支付）
func (m *defaultProductSkuModel) DeductStock(ctx context.Context, items []*product.SkuStockItem, orderSn string) (*[]*product.SkuStockResult, error) {
	return m.applyStockChange(ctx, items, orderSn, stockChangeSpec{
		changeType:  constants.STOCK_CHANGE_TYPE_DEDUCT,
		idempotent:  true,
		setClause:   "stock = stock - ?, locked_stock = locked_stock - ?",
		guardClause: "locked_stock >= ?",
		updateArgs: func(qty int64, skuId uint64) []interface{} {
			return []interface{}{qty, qty, skuId, qty}
		},
		successMsg:  "扣减成功",
		depletedMsg: "商品库存不足",
		remark:      "用户提交订单，扣减库存",
		stockDelta:  func(qty int64) int64 { return -qty },
		lockedDelta: func(qty int64) int64 { return -qty },
		abortMsg:    "部分商品操作失败，已整体回滚",
		flipMsg:     "因其他商品原因，操作失败",
	})
}

// RevertDeduct 回滚 DeductStock 操作（创建订单事务失败时调用）
func (m *defaultProductSkuModel) RevertDeduct(ctx context.Context, items []*product.SkuStockItem, orderSn string) (*[]*product.SkuStockResult, error) {
	return m.applyStockChange(ctx, items, orderSn, stockChangeSpec{
		changeType:  constants.STOCK_CHANGE_TYPE_REVERT_DEDUCT,
		setClause:   "stock = stock + ?, locked_stock = locked_stock + ?",
		updateArgs:  func(qty int64, skuId uint64) []interface{} { return []interface{}{qty, qty, skuId} },
		successMsg:  "回滚成功",
		depletedMsg: "商品不存在或已下架",
		remark:      "创建订单失败，回滚扣减",
		stockDelta:  func(qty int64) int64 { return qty },
		lockedDelta: func(qty int64) int64 { return qty },
		abortMsg:    "部分商品操作失败，已整体回滚",
		flipMsg:     "因其他商品原因，操作失败",
	})
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
