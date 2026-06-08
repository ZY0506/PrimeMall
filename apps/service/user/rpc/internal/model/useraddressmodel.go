package model

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserAddressModel = (*customUserAddressModel)(nil)

type (
	// UserAddressModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserAddressModel.
	UserAddressModel interface {
		userAddressModel
		customUserAddress
		withSession(session sqlx.Session) UserAddressModel
	}

	customUserAddressModel struct {
		*defaultUserAddressModel
	}
	customUserAddress interface {
		InsertWithDefault(ctx context.Context, data *UserAddress) (sql.Result, error)
		FindPageByUserId(ctx context.Context, userId uint64, page, pageSize int64) ([]*UserAddress, int64, error)
		SetDefaultAddress(ctx context.Context, userId, addressId uint64) error
		UpdateAddress(ctx context.Context, data *UserAddress) error
	}
)

// NewUserAddressModel returns a model for the database table.
func NewUserAddressModel(conn sqlx.SqlConn) UserAddressModel {
	return &customUserAddressModel{
		defaultUserAddressModel: newUserAddressModel(conn),
	}
}

func (m *customUserAddressModel) withSession(session sqlx.Session) UserAddressModel {
	return NewUserAddressModel(sqlx.NewSqlConnFromSession(session))
}

// InsertWithDefault 插入地址，如果 is_default = 1 则自动处理默认地址冲突
func (m *defaultUserAddressModel) InsertWithDefault(ctx context.Context, data *UserAddress) (sql.Result, error) {
	// 如果不是设为默认，直接普通插入
	if data.IsDefault != 1 {
		return m.Insert(ctx, data)
	}

	var result sql.Result
	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 清除该用户现有的默认地址（如果有）
		clearQuery := fmt.Sprintf("update %s set `is_default` = 0 where `user_id` = ? and `is_default` = 1", m.table)
		_, err := session.ExecCtx(ctx, clearQuery, data.UserId)
		if err != nil {
			return err
		}

		// 2. 插入新地址（注意这里 is_default 已经是 1）
		insertQuery := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?)", m.table, userAddressRowsExpectAutoSet)
		ret, err := session.ExecCtx(ctx, insertQuery, data.UserId, data.Tag, data.ReceiverName, data.ReceiverPhone, data.IsDefault, data.Info)
		if err != nil {
			return err
		}
		result = ret
		return nil
	})
	return result, err
}

// FindPageByUserId 分页查询用户的地址列表
func (m *defaultUserAddressModel) FindPageByUserId(ctx context.Context, userId uint64, page, pageSize int64) ([]*UserAddress, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 1. 查询总记录数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where `user_id` = ?", m.table)
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, userId)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*UserAddress{}, 0, nil
	}

	// 2. 查询分页数据（按默认地址优先，再按创建时间倒序）
	query := fmt.Sprintf("select %s from %s where `user_id` = ? order by `is_default` desc, `created_at` desc limit ? offset ?",
		userAddressRows, m.table)
	var resp []*UserAddress
	err = m.conn.QueryRowsCtx(ctx, &resp, query, userId, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return resp, total, nil
}

// SetDefaultAddress 设置默认地址
func (m *defaultUserAddressModel) SetDefaultAddress(ctx context.Context, userId, addressId uint64) error {
	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 将所有地址置为0
		clearQuery := fmt.Sprintf("update %s set `is_default` = 0 where `user_id` = ? and `is_default` = 1", m.table)
		_, err := session.ExecCtx(ctx, clearQuery, userId)
		if err != nil {
			return err
		}

		// 设置当前地址为1
		updateQuery := fmt.Sprintf("update %s set `is_default` = 1 where `id` = ?", m.table)
		result, err := session.ExecCtx(ctx, updateQuery, addressId)
		if err != nil {
			return err
		}
		affects, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affects == 0 {
			return fmt.Errorf("address not found or does not belong to user")
		}
		return nil
	})
	return err
}

// UpdateAddress 修改地址信息，包括设置默认地址
func (m *defaultUserAddressModel) UpdateAddress(ctx context.Context, data *UserAddress) error {
	if data.IsDefault != 1 {
		return m.Update(ctx, data)
	}
	return m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 清除该用户所有现有默认地址
		clearQuery := fmt.Sprintf("UPDATE %s SET is_default = 0 WHERE user_id = ? AND is_default = 1", m.table)
		if _, err := session.ExecCtx(ctx, clearQuery, data.UserId); err != nil {
			return err
		}
		// 2. 完整更新地址（包含 is_default=1），同时校验归属
		query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ? AND user_id = ?", m.table, userAddressRowsWithPlaceHolder)
		_, err := session.ExecCtx(ctx, query,
			data.UserId, data.Tag, data.ReceiverName, data.ReceiverPhone, data.IsDefault, data.Info, data.Id, data.UserId)
		return err
	})
}
