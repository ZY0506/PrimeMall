package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
	"time"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
		UpdateTx(ctx context.Context, session sqlx.Session, newData *User) error
		FindListByPage(ctx context.Context, filter *UserListFilter) ([]*User, int64, error)
		withSession(session sqlx.Session) UserModel
	}

	customUserModel struct {
		*defaultUserModel
	}

	UserListFilter struct {
		Page      int64     // 页码，从1开始
		PageSize  int64     // 每页条数
		Phone     string    // 手机号模糊匹配
		Nickname  string    // 昵称模糊匹配
		Status    int64     // 状态：0-不过滤，1-正常，2-限制下单，3-封禁
		StartTime time.Time // 注册开始时间（零值不过滤）
		EndTime   time.Time // 注册结束时间（零值不过滤）
		SortField string    // 排序字段：id, created_at, last_login_time
		SortOrder string    // 排序方向：asc, desc（不区分大小写）
	}
)

// NewUserModel returns a model for the database table.
func NewUserModel(conn sqlx.SqlConn) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn),
	}
}

func (m *customUserModel) withSession(session sqlx.Session) UserModel {
	return NewUserModel(sqlx.NewSqlConnFromSession(session))
}

// FindListByPage 分页查询用户列表（支持动态筛选和排序）
// 返回：用户列表、总记录数、错误
func (m *customUserModel) FindListByPage(ctx context.Context, filter *UserListFilter) ([]*User, int64, error) {
	// 1. 构建 WHERE 条件（软删除过滤）
	whereClause := "WHERE `deleted_at` IS NULL"
	args := make([]interface{}, 0)

	if filter.Phone != "" {
		whereClause += " AND `phone` LIKE ?"
		args = append(args, "%"+filter.Phone+"%")
	}
	if filter.Nickname != "" {
		whereClause += " AND `nickname` LIKE ?"
		args = append(args, "%"+filter.Nickname+"%")
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

	// 2. 排序（白名单防注入）
	// 允许排序的字段映射
	sortFieldMap := map[string]string{
		"id":              "id",
		"created_at":      "created_at",
		"last_login_time": "last_login_time",
	}
	// 默认排序（按 id 降序）
	dbSortField := "id"
	if field, ok := sortFieldMap[filter.SortField]; ok {
		dbSortField = field
	}
	sortOrder := "DESC"
	if filter.SortOrder != "" {
		order := strings.ToUpper(filter.SortOrder)
		if order == "ASC" {
			sortOrder = "ASC"
		}
	}
	orderClause := fmt.Sprintf("ORDER BY %s %s", dbSortField, sortOrder)

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
		return []*User{}, 0, nil
	}

	// 5. 查询列表（使用 LIMIT offset, limit 语法）
	// 注意：参数顺序为 offset, limit
	query := fmt.Sprintf("SELECT %s FROM %s %s %s LIMIT ?, ?", userRows, m.table, whereClause, orderClause)
	// 构建完整参数：WHERE 参数 + offset + limit
	queryArgs := append(args, offset, filter.PageSize)
	var users []*User
	err = m.conn.QueryRowsCtx(ctx, &users, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// UpdateTx 更新用户信息
func (m *defaultUserModel) UpdateTx(ctx context.Context, session sqlx.Session, newData *User) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, userRowsWithPlaceHolder)
	_, err := session.ExecCtx(ctx, query, newData.Phone, newData.Password, newData.Nickname, newData.Avatar, newData.Gender, newData.Birthday, newData.Status, newData.LastLoginTime, newData.LastLoginIp, newData.ExtInfo, newData.DeletedAt, newData.Id)
	return err
}
