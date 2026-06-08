package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"time"
)

var _ UserLoginLogModel = (*customUserLoginLogModel)(nil)

type (
	// UserLoginLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserLoginLogModel.
	UserLoginLogModel interface {
		userLoginLogModel
		FindLimitListByUserId(ctx context.Context, userId uint64, limit int64) ([]*UserLoginLog, error)
		FindListByPage(ctx context.Context, filter *LoginLogListFilter) ([]*UserLoginLog, int64, error)
		withSession(session sqlx.Session) UserLoginLogModel
	}

	customUserLoginLogModel struct {
		*defaultUserLoginLogModel
	}
	// LoginLogListFilter 登录日志查询筛选条件
	LoginLogListFilter struct {
		Page      int64     // 页码（从1开始）
		PageSize  int64     // 每页条数
		UserId    uint64    // 用户ID（0表示不过滤）
		LoginType string    // 登录方式：password/captcha（空字符串表示不过滤）
		Status    int64     // 状态：1-成功，2-失败（0表示不过滤）
		StartTime time.Time // 开始时间（零值不过滤）
		EndTime   time.Time // 结束时间（零值不过滤）
	}
)

// NewUserLoginLogModel returns a model for the database table.
func NewUserLoginLogModel(conn sqlx.SqlConn) UserLoginLogModel {
	return &customUserLoginLogModel{
		defaultUserLoginLogModel: newUserLoginLogModel(conn),
	}
}

func (m *customUserLoginLogModel) withSession(session sqlx.Session) UserLoginLogModel {
	return NewUserLoginLogModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUserLoginLogModel) FindLimitListByUserId(ctx context.Context, userId uint64, limit int64) ([]*UserLoginLog, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? order by `id` desc limit ?", userLoginLogRows, m.table)
	var resp []*UserLoginLog
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, limit)
	return resp, err
}

// FindListByPage 分页查询登录日志列表
// 返回值：日志列表、总记录数、错误
func (m *customUserLoginLogModel) FindListByPage(ctx context.Context, filter *LoginLogListFilter) ([]*UserLoginLog, int64, error) {
	// 1. 构建 WHERE 条件
	whereClause := "WHERE 1=1"
	args := make([]interface{}, 0)

	if filter.UserId > 0 {
		whereClause += " AND `user_id` = ?"
		args = append(args, filter.UserId)
	}
	if filter.LoginType != "" {
		whereClause += " AND `login_type` = ?"
		args = append(args, filter.LoginType)
	}
	if filter.Status > 0 {
		whereClause += " AND `status` = ?"
		args = append(args, filter.Status)
	}
	if !filter.StartTime.IsZero() {
		whereClause += " AND `created_at` >= ?"
		args = append(args, filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		whereClause += " AND `created_at` <= ?"
		args = append(args, filter.EndTime)
	}

	// 2. 排序（固定按创建时间倒序，最新在前）
	orderClause := "ORDER BY `created_at` DESC"

	// 3. 分页参数处理
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20 // 默认每页20条
	}
	offset := (filter.Page - 1) * filter.PageSize

	// 4. 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, whereClause)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*UserLoginLog{}, 0, nil
	}

	// 5. 查询列表
	query := fmt.Sprintf("SELECT %s FROM %s %s %s LIMIT ?, ?",
		userLoginLogRows, m.table, whereClause, orderClause)
	// 参数：WHERE条件参数 + offset + pageSize
	queryArgs := append(args, offset, filter.PageSize)
	var logs []*UserLoginLog
	err = m.conn.QueryRowsCtx(ctx, &logs, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
