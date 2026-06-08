package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CouponModel = (*customCouponModel)(nil)

type (
	CouponModel interface {
		couponModel
		customCoupon
	}

	customCouponModel struct {
		*defaultCouponModel
	}

	customCoupon interface {
		FindPageList(ctx context.Context, page, size int64) ([]*Coupon, int64, error)
		IncrUsedQuantity(ctx context.Context, id uint64, delta int64) error
	}
)

func NewCouponModel(conn sqlx.SqlConn) CouponModel {
	return &customCouponModel{
		defaultCouponModel: newCouponModel(conn),
	}
}

func (m *customCouponModel) FindPageList(ctx context.Context, page, size int64) ([]*Coupon, int64, error) {
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE status = 1", m.table)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*Coupon{}, 0, nil
	}

	offset := (page - 1) * size
	query := fmt.Sprintf("SELECT %s FROM %s WHERE status = 1 ORDER BY id DESC LIMIT ? OFFSET ?", couponRows, m.table)
	var items []*Coupon
	if err := m.conn.QueryRowsCtx(ctx, &items, query, size, offset); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (m *customCouponModel) IncrUsedQuantity(ctx context.Context, id uint64, delta int64) error {
	query := fmt.Sprintf("UPDATE %s SET used_quantity = used_quantity + ? WHERE id = ? AND used_quantity + ? <= total_quantity", m.table)
	result, err := m.conn.ExecCtx(ctx, query, delta, id, delta)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("coupon stock exhausted")
	}
	return nil
}

// FindActiveList 查找当前时间范围内可用的优惠券列表
func (m *customCouponModel) FindActiveList(ctx context.Context) ([]*Coupon, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE status = 1 AND start_time <= NOW() AND end_time >= NOW() ORDER BY id DESC", couponRows, m.table)
	var items []*Coupon
	if err := m.conn.QueryRowsCtx(ctx, &items, query); err != nil {
		return nil, err
	}
	return items, nil
}

// FindByIDs 批量查找
func (m *customCouponModel) FindByIDs(ctx context.Context, ids []uint64) ([]*Coupon, error) {
	if len(ids) == 0 {
		return []*Coupon{}, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id IN (%s)", couponRows, m.table, strings.Join(placeholders, ","))
	var items []*Coupon
	if err := m.conn.QueryRowsCtx(ctx, &items, query, args...); err != nil {
		return nil, err
	}
	return items, nil
}
