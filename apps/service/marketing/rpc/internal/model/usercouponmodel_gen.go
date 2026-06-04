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
	userCouponFieldNames          = builder.RawFieldNames(&UserCoupon{})
	userCouponRows                = strings.Join(userCouponFieldNames, ",")
	userCouponRowsExpectAutoSet   = strings.Join(stringx.Remove(userCouponFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`update_at`", "`update_time`", "`updated_at`"), ",")
	userCouponRowsWithPlaceHolder = strings.Join(stringx.Remove(userCouponFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`update_at`", "`update_time`", "`updated_at`"), "=?,") + "=?"
)

type (
	userCouponModel interface {
		Insert(ctx context.Context, data *UserCoupon) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*UserCoupon, error)
		Update(ctx context.Context, data *UserCoupon) error
		Delete(ctx context.Context, id uint64) error
	}

	defaultUserCouponModel struct {
		conn  sqlx.SqlConn
		table string
	}

	UserCoupon struct {
		Id         uint64       `db:"id"`
		CouponId   uint64       `db:"coupon_id"`
		UserId     uint64       `db:"user_id"`
		OrderSn    string       `db:"order_sn"`
		Status     int64        `db:"status"`
		UsedTime   sql.NullTime `db:"used_time"`
		Source     string       `db:"source"`
		CreatedAt  time.Time    `db:"created_at"`
		ExpireTime time.Time    `db:"expire_time"`
	}
)

func newUserCouponModel(conn sqlx.SqlConn) *defaultUserCouponModel {
	return &defaultUserCouponModel{
		conn:  conn,
		table: "user_coupon",
	}
}

func (m *defaultUserCouponModel) Insert(ctx context.Context, data *UserCoupon) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?)", m.table, userCouponRowsExpectAutoSet)
	ret, err := m.conn.ExecCtx(ctx, query, data.CouponId, data.UserId, data.OrderSn,
		data.Status, data.UsedTime, data.Source, data.CreatedAt, data.ExpireTime)
	return ret, err
}

func (m *defaultUserCouponModel) FindOne(ctx context.Context, id uint64) (*UserCoupon, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", userCouponRows, m.table)
	var resp UserCoupon
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

func (m *defaultUserCouponModel) Update(ctx context.Context, data *UserCoupon) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, userCouponRowsWithPlaceHolder)
	_, err := m.conn.ExecCtx(ctx, query, data.CouponId, data.UserId, data.OrderSn,
		data.Status, data.UsedTime, data.Source, data.ExpireTime, data.Id)
	return err
}

func (m *defaultUserCouponModel) Delete(ctx context.Context, id uint64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}
