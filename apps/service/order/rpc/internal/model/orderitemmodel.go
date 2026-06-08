package model

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ OrderItemModel = (*customOrderItemModel)(nil)

type (
	// OrderItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customOrderItemModel.
	OrderItemModel interface {
		orderItemModel
		InsertTx(ctx context.Context, session sqlx.Session, data *OrderItem) (sql.Result, error)
		FindListByOrderId(ctx context.Context, orderId uint64) ([]*OrderItem, error)
		FindOneByOrderSnSkuId(ctx context.Context, orderSn string, skuId uint64) (*OrderItem, error)
		withSession(session sqlx.Session) OrderItemModel
	}

	customOrderItemModel struct {
		*defaultOrderItemModel
	}
)

// NewOrderItemModel returns a model for the database table.
func NewOrderItemModel(conn sqlx.SqlConn) OrderItemModel {
	return &customOrderItemModel{
		defaultOrderItemModel: newOrderItemModel(conn),
	}
}

func (m *customOrderItemModel) withSession(session sqlx.Session) OrderItemModel {
	return NewOrderItemModel(sqlx.NewSqlConnFromSession(session))
}

// InsertTx 事务操作
func (m *customOrderItemModel) InsertTx(ctx context.Context, session sqlx.Session, data *OrderItem) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, orderItemRowsExpectAutoSet)
	ret, err := session.ExecCtx(ctx, query, data.Id, data.OrderId, data.OrderSn, data.SkuId, data.SpuId, data.SpuName, data.SkuName, data.SkuPic, data.Price, data.Count, data.TotalAmount, data.IsSeckill)
	return ret, err
}

// FindListByOrderId 获取订单项列表
func (m *customOrderItemModel) FindListByOrderId(ctx context.Context, orderId uint64) ([]*OrderItem, error) {
	query := fmt.Sprintf("select %s from %s where order_id = ?", orderItemRows, m.table)
	var resp []*OrderItem
	err := m.conn.QueryRowsCtx(ctx, &resp, query, orderId)
	return resp, err
}

// FindOneByOrderSnSkuId 通过订单号和skuId查询
func (m *customOrderItemModel) FindOneByOrderSnSkuId(ctx context.Context, orderSn string, skuId uint64) (*OrderItem, error) {
	query := fmt.Sprintf("select %s from %s where `order_sn` = ? and `sku_id` = ? limit 1", orderItemRows, m.table)
	var resp OrderItem
	err := m.conn.QueryRowCtx(ctx, &resp, query, orderSn, skuId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
