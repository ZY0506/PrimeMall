package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RoleModel = (*customRoleModel)(nil)

type (
	RoleModel interface {
		roleModel
		FindAll(ctx context.Context) ([]*Role, error)
		withSession(session sqlx.Session) RoleModel
	}

	customRoleModel struct {
		*defaultRoleModel
	}
)

// NewRoleModel returns a model for the database table.
func NewRoleModel(conn sqlx.SqlConn) RoleModel {
	return &customRoleModel{
		defaultRoleModel: newRoleModel(conn),
	}
}

func (m *customRoleModel) withSession(session sqlx.Session) RoleModel {
	return NewRoleModel(sqlx.NewSqlConnFromSession(session))
}

// FindAll 查询所有角色
func (m *customRoleModel) FindAll(ctx context.Context) ([]*Role, error) {
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY `id` ASC", roleRows, m.table)
	var roles []*Role
	err := m.conn.QueryRowsCtx(ctx, &roles, query)
	if err != nil {
		return nil, err
	}
	return roles, nil
}
