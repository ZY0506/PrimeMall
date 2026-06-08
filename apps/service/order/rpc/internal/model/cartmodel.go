package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
)

var _ CartModel = (*customCartModel)(nil)

type (
	// CartModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCartModel.
	CartModel interface {
		cartModel
		FindListByUserIdSkuIds(ctx context.Context, userId uint64, skuIds []uint64) ([]*Cart, error)
		BatchDelete(ctx context.Context, userId uint64, skuIds []uint64) error
		ClearCart(ctx context.Context, userId uint64) error
		FindListByUserId(ctx context.Context, userId uint64) ([]*Cart, error)
		BatchSelect(ctx context.Context, userId uint64, skuIds []uint64, selected, all bool) error
		withSession(session sqlx.Session) CartModel
	}

	customCartModel struct {
		*defaultCartModel
	}
)

// NewCartModel returns a model for the database table.
func NewCartModel(conn sqlx.SqlConn) CartModel {
	return &customCartModel{
		defaultCartModel: newCartModel(conn),
	}
}

func (m *customCartModel) withSession(session sqlx.Session) CartModel {
	return NewCartModel(sqlx.NewSqlConnFromSession(session))
}

// FindListByUserIdSkuIds 获取购物车列表
func (m *customCartModel) FindListByUserIdSkuIds(ctx context.Context, userId uint64, skuIds []uint64) ([]*Cart, error) {
	if len(skuIds) == 0 {
		return []*Cart{}, nil
	}
	var resp []*Cart
	// 生成占位符
	placeholders := strings.Repeat("?,", len(skuIds))
	placeholders = placeholders[:len(placeholders)-1] // 去掉末尾多余的逗号
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `sku_id` in (%s)", cartRows, m.table, placeholders)

	args := make([]interface{}, len(skuIds))
	for i, skuId := range skuIds {
		args[i] = skuId
	}
	err := m.conn.QueryRowsCtx(ctx, &resp, query, append([]interface{}{userId}, args...))
	return resp, err
}

// BatchDelete 批量删除购物车
func (m *customCartModel) BatchDelete(ctx context.Context, userId uint64, skuIds []uint64) error {
	if len(skuIds) == 0 {
		return nil
	}
	// 批量删除
	placeholders := strings.Repeat("?,", len(skuIds))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf("delete from %s where `user_id` = ? and `sku_id` in (%s)", m.table, placeholders)
	args := make([]interface{}, len(skuIds))
	for i, skuId := range skuIds {
		args[i] = skuId
	}
	_, err := m.conn.ExecCtx(ctx, query, append([]interface{}{userId}, args...)...)
	return err
}

// ClearCart 清空购物车
func (m *customCartModel) ClearCart(ctx context.Context, userId uint64) error {
	_, err := m.conn.ExecCtx(ctx, "delete from "+m.table+" where `user_id` = ?", userId)
	return err
}

// FindListByUserId 获取购物车列表
func (m *customCartModel) FindListByUserId(ctx context.Context, userId uint64) ([]*Cart, error) {
	var resp []*Cart
	query := fmt.Sprintf("select %s from %s where `user_id` = ?", cartRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	return resp, err
}

// BatchSelect 批量修改选中状态
func (m *customCartModel) BatchSelect(ctx context.Context, userId uint64, skuIds []uint64, selected, all bool) error {
	var query string
	if all {
		query = fmt.Sprintf("update %s set `selected` = ? where `user_id` = ?", m.table)
		_, err := m.conn.ExecCtx(ctx, query, selected, userId)
		return err
	}
	if len(skuIds) == 0 {
		return nil
	}
	// 拼接占位符
	placeholders := strings.Join(make([]string, len(skuIds)), "?,") + "?"
	query = fmt.Sprintf("update %s set `selected` = ? where `user_id` = ? and `sku_id` in (%s)", m.table, placeholders)
	args := make([]interface{}, len(skuIds))
	for i, skuId := range skuIds {
		args[i] = skuId
	}
	_, err := m.conn.ExecCtx(ctx, query, append([]interface{}{selected, userId}, args...))
	return err
}
