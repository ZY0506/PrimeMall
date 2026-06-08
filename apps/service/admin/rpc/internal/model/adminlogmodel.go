package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminLogModel = (*customAdminLogModel)(nil)

type (
	// AdminLogListFilter 管理员日志查询过滤器
	AdminLogListFilter struct {
		Page     int64
		PageSize int64
		AdminId  uint64
		Module   string
	}

	AdminLogModel interface {
		adminLogModel
		FindListByPage(ctx context.Context, filter *AdminLogListFilter) ([]*AdminLog, int64, error)
		withSession(session sqlx.Session) AdminLogModel
	}

	customAdminLogModel struct {
		*defaultAdminLogModel
	}
)

// NewAdminLogModel returns a model for the database table.
func NewAdminLogModel(conn sqlx.SqlConn) AdminLogModel {
	return &customAdminLogModel{
		defaultAdminLogModel: newAdminLogModel(conn),
	}
}

func (m *customAdminLogModel) withSession(session sqlx.Session) AdminLogModel {
	return NewAdminLogModel(sqlx.NewSqlConnFromSession(session))
}

// FindListByPage 分页查询管理员日志
func (m *customAdminLogModel) FindListByPage(ctx context.Context, filter *AdminLogListFilter) ([]*AdminLog, int64, error) {
	whereClause := "WHERE 1=1"
	args := make([]interface{}, 0)

	if filter.AdminId > 0 {
		whereClause += " AND `admin_id` = ?"
		args = append(args, filter.AdminId)
	}
	if filter.Module != "" {
		whereClause += " AND `module` = ?"
		args = append(args, filter.Module)
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

	query := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY `id` DESC LIMIT ?,?", adminLogRows, m.table, whereClause)
	queryArgs := append(args, offset, filter.PageSize)

	var logs []*AdminLog
	err = m.conn.QueryRowsCtx(ctx, &logs, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// InsertTx 事务插入管理员日志
func (m *customAdminLogModel) InsertTx(ctx context.Context, session sqlx.Session, data *AdminLog) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, adminLogRowsExpectAutoSet)
	ret, err := session.ExecCtx(ctx, query, data.AdminId, data.Username, data.Module, data.Action, data.RequestMethod, data.RequestUrl, data.RequestParams, data.ResponseResult, data.Ip, data.DurationMs)
	return ret, err
}
