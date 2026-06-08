package model

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"time"
)

var _ UserPunishLogModel = (*customUserPunishLogModel)(nil)

type (
	// UserPunishLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserPunishLogModel.
	UserPunishLogModel interface {
		userPunishLogModel
		InsertTx(ctx context.Context, session sqlx.Session, data *UserPunishLog) (sql.Result, error)
		UpdateTx(ctx context.Context, session sqlx.Session, data *UserPunishLog) error
		FindListByUserId(ctx context.Context, userId uint64) ([]*UserPunishLog, error)
		FindListByPage(ctx context.Context, filter *PunishLogListFilter) ([]*UserPunishLog, int64, error)
		FindLastLogByUserId(ctx context.Context, userId uint64) (*UserPunishLog, error)
		withSession(session sqlx.Session) UserPunishLogModel
	}

	customUserPunishLogModel struct {
		*defaultUserPunishLogModel
	}
	// PunishLogListFilter 风控日志查询筛选条件
	PunishLogListFilter struct {
		Page       int64     // 页码（从1开始）
		PageSize   int64     // 每页条数
		UserId     uint64    // 用户ID（0表示不过滤）
		ActionType int64     // 处罚类型：1-禁止下单，2-禁止登录（0表示不过滤）
		Operator   string    // 操作人：SYSTEM/ADMIN（空字符串表示不过滤）
		StartTime  time.Time // 开始时间（零值不过滤）
		EndTime    time.Time // 结束时间（零值不过滤）
	}
)

// NewUserPunishLogModel returns a model for the database table.
func NewUserPunishLogModel(conn sqlx.SqlConn) UserPunishLogModel {
	return &customUserPunishLogModel{
		defaultUserPunishLogModel: newUserPunishLogModel(conn),
	}
}

func (m *customUserPunishLogModel) withSession(session sqlx.Session) UserPunishLogModel {
	return NewUserPunishLogModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUserPunishLogModel) FindListByUserId(ctx context.Context, userId uint64) ([]*UserPunishLog, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? order by id desc", userPunishLogRows, m.table)
	var resp []*UserPunishLog
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	return resp, err
}

// FindListByPage 分页查询风控日志列表
func (m *customUserPunishLogModel) FindListByPage(ctx context.Context, filter *PunishLogListFilter) ([]*UserPunishLog, int64, error) {
	// 1. 构建 WHERE 条件
	whereClause := "WHERE 1=1"
	args := make([]interface{}, 0)

	if filter.UserId > 0 {
		whereClause += " AND `user_id` = ?"
		args = append(args, filter.UserId)
	}
	if filter.ActionType > 0 {
		whereClause += " AND `action_type` = ?"
		args = append(args, filter.ActionType)
	}
	if filter.Operator != "" {
		whereClause += " AND `operator` = ?"
		args = append(args, filter.Operator)
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
		return []*UserPunishLog{}, 0, nil
	}

	// 5. 查询列表
	query := fmt.Sprintf("SELECT %s FROM %s %s %s LIMIT ?, ?",
		userPunishLogRows, m.table, whereClause, orderClause)
	queryArgs := append(args, offset, filter.PageSize)
	var logs []*UserPunishLog
	err = m.conn.QueryRowsCtx(ctx, &logs, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// FindLastLogByUserId 查询指定用户ID的最新风控日志
func (m *customUserPunishLogModel) FindLastLogByUserId(ctx context.Context, userId uint64) (*UserPunishLog, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? order by `created_at` desc limit 1", userPunishLogRows, m.table)
	var resp UserPunishLog
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// InsertTx 插入数据
func (m *defaultUserPunishLogModel) InsertTx(ctx context.Context, session sqlx.Session, data *UserPunishLog) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?)", m.table, userPunishLogRowsExpectAutoSet)
	ret, err := session.ExecCtx(ctx, query, data.UserId, data.Phone, data.ActionType, data.Reason, data.BannedBy, data.UnbannedBy, data.StartTime, data.EndTime)
	return ret, err
}

// UpdateTx 更新数据
func (m *defaultUserPunishLogModel) UpdateTx(ctx context.Context, session sqlx.Session, data *UserPunishLog) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, userPunishLogRowsWithPlaceHolder)
	_, err := session.ExecCtx(ctx, query, data.UserId, data.Phone, data.ActionType, data.Reason, data.BannedBy, data.UnbannedBy, data.StartTime, data.EndTime, data.Id)
	return err
}
