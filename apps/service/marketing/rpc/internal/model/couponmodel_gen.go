package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	couponFieldNames          = builder.RawFieldNames(&Coupon{})
	couponRows                = strings.Join(couponFieldNames, ",")
	couponRowsExpectAutoSet   = strings.Join(stringx.Remove(couponFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`update_at`", "`update_time`", "`updated_at`"), ",")
	couponRowsWithPlaceHolder = strings.Join(stringx.Remove(couponFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`update_at`", "`update_time`", "`updated_at`"), "=?,") + "=?"
)

type (
	couponModel interface {
		Insert(ctx context.Context, data *Coupon) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*Coupon, error)
		Update(ctx context.Context, data *Coupon) error
		Delete(ctx context.Context, id uint64) error
	}

	defaultCouponModel struct {
		conn  sqlx.SqlConn
		table string
	}

	Coupon struct {
		Id                uint64    `db:"id"`
		Name              string    `db:"name"`
		Type              int64     `db:"type"`
		ThresholdAmount   int64     `db:"threshold_amount"`
		ReduceAmount      int64     `db:"reduce_amount"`
		DiscountRate      int64     `db:"discount_rate"`
		MaxDiscountAmount int64     `db:"max_discount_amount"`
		TotalQuantity     int64     `db:"total_quantity"`
		UsedQuantity      int64     `db:"used_quantity"`
		PerUserLimit      int64     `db:"per_user_limit"`
		StartTime         time.Time `db:"start_time"`
		EndTime           time.Time `db:"end_time"`
		Status            int64     `db:"status"`
		Description       string    `db:"description"`
		CreatedAt         time.Time `db:"created_at"`
		UpdatedAt         time.Time `db:"updated_at"`
	}
)

func newCouponModel(conn sqlx.SqlConn) *defaultCouponModel {
	return &defaultCouponModel{
		conn:  conn,
		table: "coupon",
	}
}

func (m *defaultCouponModel) Insert(ctx context.Context, data *Coupon) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, couponRowsExpectAutoSet)
	ret, err := m.conn.ExecCtx(ctx, query, data.Name, data.Type, data.ThresholdAmount, data.ReduceAmount,
		data.DiscountRate, data.MaxDiscountAmount, data.TotalQuantity, data.UsedQuantity,
		data.PerUserLimit, data.StartTime, data.EndTime, data.Status, data.Description)
	return ret, err
}

func (m *defaultCouponModel) FindOne(ctx context.Context, id uint64) (*Coupon, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", couponRows, m.table)
	var resp Coupon
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultCouponModel) Update(ctx context.Context, data *Coupon) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, couponRowsWithPlaceHolder)
	_, err := m.conn.ExecCtx(ctx, query, data.Name, data.Type, data.ThresholdAmount, data.ReduceAmount,
		data.DiscountRate, data.MaxDiscountAmount, data.TotalQuantity, data.UsedQuantity,
		data.PerUserLimit, data.StartTime, data.EndTime, data.Status, data.Description, data.Id)
	return err
}

func (m *defaultCouponModel) Delete(ctx context.Context, id uint64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}
