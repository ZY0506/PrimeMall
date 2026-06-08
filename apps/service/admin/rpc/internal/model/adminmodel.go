package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminModel = (*customAdminModel)(nil)

type (
	// AdminListFilter 管理员列表查询过滤器
	AdminListFilter struct {
		Page     int64
		PageSize int64
		Username string
		Status   int64
	}

	// AdminModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminModel.
	AdminModel interface {
		adminModel
		FindListByPage(ctx context.Context, filter *AdminListFilter) ([]*Admin, int64, error)
		UpdateTx(ctx context.Context, session sqlx.Session, newData *Admin) error
		UpdateLoginInfo(ctx context.Context, id uint64, ip string) error
		withSession(session sqlx.Session) AdminModel
	}

	customAdminModel struct {
		*defaultAdminModel
	}
)

// NewAdminModel returns a model for the database table.
func NewAdminModel(conn sqlx.SqlConn) AdminModel {
	return &customAdminModel{
		defaultAdminModel: newAdminModel(conn),
	}
}

func (m *customAdminModel) withSession(session sqlx.Session) AdminModel {
	return NewAdminModel(sqlx.NewSqlConnFromSession(session))
}

// FindListByPage 分页查询管理员列表
func (m *customAdminModel) FindListByPage(ctx context.Context, filter *AdminListFilter) ([]*Admin, int64, error) {
	whereClause := "WHERE 1=1"
	args := make([]interface{}, 0)

	if filter.Username != "" {
		whereClause += " AND `username` LIKE ?"
		args = append(args, "%"+filter.Username+"%")
	}
	if filter.Status > 0 {
		whereClause += " AND `status` = ?"
		args = append(args, filter.Status)
	}

	// 查总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, whereClause)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 查列表
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	offset := (filter.Page - 1) * filter.PageSize

	query := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY `id` DESC LIMIT ?,?", adminRows, m.table, whereClause)
	queryArgs := append(args, offset, filter.PageSize)

	var admins []*Admin
	err = m.conn.QueryRowsCtx(ctx, &admins, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}

	return admins, total, nil
}

// UpdateTx 事务更新管理员
func (m *customAdminModel) UpdateTx(ctx context.Context, session sqlx.Session, newData *Admin) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, adminRowsWithPlaceHolder)
	_, err := session.ExecCtx(ctx, query, newData.Username, newData.Password, newData.RealName, newData.Avatar, newData.RoleId, newData.Status, newData.LastLoginTime, newData.LastLoginIp, newData.Id)
	return err
}

// UpdateLoginInfo 更新管理员登录信息
func (m *customAdminModel) UpdateLoginInfo(ctx context.Context, id uint64, ip string) error {
	now := sql.NullTime{Time: time.Now(), Valid: true}
	query := fmt.Sprintf("UPDATE %s SET `last_login_time` = ?, `last_login_ip` = ? WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, now, ip, id)
	return err
}
