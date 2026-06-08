package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PaymentModel = (*customPaymentModel)(nil)

type (
	// PaymentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPaymentModel.
	PaymentModel interface {
		paymentModel
		FindListByPage(ctx context.Context, filter *PaymentListFilter) ([]*Payment, int64, error)
		withSession(session sqlx.Session) PaymentModel
	}

	customPaymentModel struct {
		*defaultPaymentModel
	}

	PaymentListFilter struct {
		Page      int64
		PageSize  int64
		Status    int64
		Channel   string
		OrderSn   string
		StartTime time.Time
		EndTime   time.Time
	}
)

// NewPaymentModel returns a model for the database table.
func NewPaymentModel(conn sqlx.SqlConn) PaymentModel {
	return &customPaymentModel{
		defaultPaymentModel: newPaymentModel(conn),
	}
}

func (m *customPaymentModel) withSession(session sqlx.Session) PaymentModel {
	return NewPaymentModel(sqlx.NewSqlConnFromSession(session))
}

// FindListByPage 分页查询支付记录
func (m *customPaymentModel) FindListByPage(ctx context.Context, filter *PaymentListFilter) ([]*Payment, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	where := "where 1=1"
	var args []interface{}

	if filter.Status > 0 {
		where += " and `status` = ?"
		args = append(args, filter.Status)
	}
	if filter.Channel != "" {
		where += " and `channel` = ?"
		args = append(args, filter.Channel)
	}
	if filter.OrderSn != "" {
		where += " and `order_sn` like ?"
		args = append(args, "%"+filter.OrderSn+"%")
	}
	if !filter.StartTime.IsZero() {
		where += " and `created_at` >= ?"
		args = append(args, filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		where += " and `created_at` <= ?"
		args = append(args, filter.EndTime)
	}

	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s %s", m.table, where)
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*Payment{}, 0, nil
	}

	offset := (filter.Page - 1) * filter.PageSize
	dataQuery := fmt.Sprintf("select %s from %s %s order by `created_at` desc limit ?, ?", paymentRows, m.table, where)
	args = append(args, offset, filter.PageSize)

	var list []*Payment
	err = m.conn.QueryRowsCtx(ctx, &list, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
