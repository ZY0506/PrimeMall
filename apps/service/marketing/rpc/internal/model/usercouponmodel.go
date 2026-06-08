package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserCouponModel = (*customUserCouponModel)(nil)

type (
	UserCouponModel interface {
		userCouponModel
		customUserCoupon
	}

	customUserCouponModel struct {
		*defaultUserCouponModel
	}

	customUserCoupon interface {
		FindByUserID(ctx context.Context, userID uint64, status int64, page, size int64) ([]*UserCoupon, int64, error)
		FindByUserAndCoupon(ctx context.Context, userID, couponID uint64) (*UserCoupon, error)
		FindAvailableByUser(ctx context.Context, userID uint64) ([]*UserCoupon, error)
		UpdateStatus(ctx context.Context, id uint64, status int64, orderSn string, usedTime time.Time) error
		CountByUserAndCoupon(ctx context.Context, userID, couponID uint64) (int64, error)
	}
)

func NewUserCouponModel(conn sqlx.SqlConn) UserCouponModel {
	return &customUserCouponModel{
		defaultUserCouponModel: newUserCouponModel(conn),
	}
}

func (m *customUserCouponModel) FindByUserID(ctx context.Context, userID uint64, status int64, page, size int64) ([]*UserCoupon, int64, error) {
	where := []string{"user_id = ?"}
	args := []any{userID}

	if status >= 0 && status <= 2 {
		where = append(where, "status = ?")
		args = append(args, status)
	}

	whereSQL := "WHERE " + strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, whereSQL)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*UserCoupon{}, 0, nil
	}

	offset := (page - 1) * size
	query := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY id DESC LIMIT ? OFFSET ?", userCouponRows, m.table, whereSQL)
	queryArgs := append(args, size, offset)
	var items []*UserCoupon
	if err := m.conn.QueryRowsCtx(ctx, &items, query, queryArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (m *customUserCouponModel) FindByUserAndCoupon(ctx context.Context, userID, couponID uint64) (*UserCoupon, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = ? AND coupon_id = ? LIMIT 1", userCouponRows, m.table)
	var resp UserCoupon
	err := m.conn.QueryRowCtx(ctx, &resp, query, userID, couponID)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, nil
	default:
		return nil, err
	}
}

func (m *customUserCouponModel) FindAvailableByUser(ctx context.Context, userID uint64) ([]*UserCoupon, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = ? AND status = 0 AND expire_time > NOW() ORDER BY id DESC", userCouponRows, m.table)
	var items []*UserCoupon
	if err := m.conn.QueryRowsCtx(ctx, &items, query, userID); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *customUserCouponModel) UpdateStatus(ctx context.Context, id uint64, status int64, orderSn string, usedTime time.Time) error {
	query := fmt.Sprintf("UPDATE %s SET status = ?, order_sn = ?, used_time = ? WHERE id = ? AND status = 0", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, orderSn, usedTime, id)
	return err
}

func (m *customUserCouponModel) CountByUserAndCoupon(ctx context.Context, userID, couponID uint64) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE user_id = ? AND coupon_id = ?", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, userID, couponID)
	return count, err
}
