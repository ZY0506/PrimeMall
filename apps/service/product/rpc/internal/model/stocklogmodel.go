package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ StockLogModel = (*customStockLogModel)(nil)

type (
	// StockLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customStockLogModel.
	StockLogModel interface {
		stockLogModel
		customStockLog
		withSession(session sqlx.Session) StockLogModel
	}

	customStockLogModel struct {
		*defaultStockLogModel
	}

	customStockLog interface {
		FindList(ctx context.Context, skuId uint64, page, pageSize int64) ([]*StockLog, int64, error)
	}
)

// NewStockLogModel returns a model for the database table.
func NewStockLogModel(conn sqlx.SqlConn) StockLogModel {
	return &customStockLogModel{
		defaultStockLogModel: newStockLogModel(conn),
	}
}

func (m *customStockLogModel) withSession(session sqlx.Session) StockLogModel {
	return NewStockLogModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customStockLogModel) FindList(ctx context.Context, skuId uint64, page, pageSize int64) ([]*StockLog, int64, error) {
	if page <= 0 || pageSize <= 0 {
		return nil, 0, nil
	}

	var total int64
	var countQuery string
	var countArgs []interface{}

	if skuId > 0 {
		countQuery = fmt.Sprintf("select count(*) from %s where `sku_id` = ?", m.table)
		countArgs = append(countArgs, skuId)
	} else {
		countQuery = fmt.Sprintf("select count(*) from %s", m.table)
	}

	err := m.conn.QueryRowCtx(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*StockLog{}, 0, nil
	}

	offset := (page - 1) * pageSize
	var query string
	var args []interface{}

	if skuId > 0 {
		query = fmt.Sprintf("select %s from %s where `sku_id` = ? order by created_at desc limit ?, ?", stockLogRows, m.table)
		args = append(args, skuId, offset, pageSize)
	} else {
		query = fmt.Sprintf("select %s from %s order by created_at desc limit ?, ?", stockLogRows, m.table)
		args = append(args, offset, pageSize)
	}

	var resp []*StockLog
	err = m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return resp, total, nil
}
