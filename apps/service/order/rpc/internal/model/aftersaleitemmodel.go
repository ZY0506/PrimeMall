package model

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
)

var _ AfterSaleItemModel = (*customAfterSaleItemModel)(nil)

type (
	// AfterSaleItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAfterSaleItemModel.
	AfterSaleItemModel interface {
		afterSaleItemModel
		InsertTx(ctx context.Context, session sqlx.Session, data *AfterSaleItem) (sql.Result, error)
		FindOneByAfterSaleSn(ctx context.Context, afterSaleSn string) (*AfterSaleItem, error)
		FindMapByAfterSaleSns(ctx context.Context, afterSaleSns []string) (map[string]*AfterSaleItem, error)
		withSession(session sqlx.Session) AfterSaleItemModel
	}

	customAfterSaleItemModel struct {
		*defaultAfterSaleItemModel
	}
)

// NewAfterSaleItemModel returns a model for the database table.
func NewAfterSaleItemModel(conn sqlx.SqlConn) AfterSaleItemModel {
	return &customAfterSaleItemModel{
		defaultAfterSaleItemModel: newAfterSaleItemModel(conn),
	}
}

func (m *customAfterSaleItemModel) withSession(session sqlx.Session) AfterSaleItemModel {
	return NewAfterSaleItemModel(sqlx.NewSqlConnFromSession(session))
}

func (m *defaultAfterSaleItemModel) InsertTx(ctx context.Context, session sqlx.Session, data *AfterSaleItem) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, afterSaleItemRowsExpectAutoSet)
	ret, err := session.ExecCtx(ctx, query, data.AfterSaleSn, data.OrderItemId, data.SkuId, data.SpuId, data.ProductName, data.SkuName, data.SkuPic, data.Price, data.Quantity, data.TotalAmount, data.RefundAmount)
	return ret, err
}

func (m *customAfterSaleItemModel) FindOneByAfterSaleSn(ctx context.Context, afterSaleSn string) (*AfterSaleItem, error) {
	query := fmt.Sprintf("select %s from %s where after_sale_sn = ? limit 1", afterSaleItemRows, m.table)
	var resp *AfterSaleItem
	err := m.conn.QueryRowCtx(ctx, &resp, query, afterSaleSn)
	return resp, err
}

// FindMapByAfterSaleSns 根据售后单号列表批量查询商品，返回map结构 key:after_sale_sn
func (m *customAfterSaleItemModel) FindMapByAfterSaleSns(ctx context.Context, afterSaleSns []string) (map[string]*AfterSaleItem, error) {
	// 空列表直接返回空map
	if len(afterSaleSns) == 0 {
		return make(map[string]*AfterSaleItem), nil
	}

	placeholders := strings.Join(make([]string, len(afterSaleSns)), "?,") + "?"
	args := make([]any, len(afterSaleSns))
	for i, v := range afterSaleSns {
		args[i] = v
	}

	// 2. 拼接SQL
	query := fmt.Sprintf("select %s from %s where after_sale_sn in (%s)", afterSaleItemRows, m.table, placeholders)

	// 4. go-zero 原生查询
	var items []*AfterSaleItem
	err := m.conn.QueryRowsCtx(ctx, &items, query, args...)
	if err != nil {
		return nil, err
	}

	// 5. 组装 map，key = 售后单号，快速匹配
	resultMap := make(map[string]*AfterSaleItem, len(items))
	for _, item := range items {
		resultMap[item.AfterSaleSn] = item
	}

	return resultMap, nil
}
