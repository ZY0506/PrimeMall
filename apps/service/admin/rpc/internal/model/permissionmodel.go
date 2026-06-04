package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PermissionModel = (*customPermissionModel)(nil)

type (
	PermissionModel interface {
		permissionModel
		FindAll(ctx context.Context) ([]*Permission, error)
		FindByModule(ctx context.Context, module string) ([]*Permission, error)
		withSession(session sqlx.Session) PermissionModel
	}

	customPermissionModel struct {
		*defaultPermissionModel
	}
)

// NewPermissionModel returns a model for the database table.
func NewPermissionModel(conn sqlx.SqlConn) PermissionModel {
	return &customPermissionModel{
		defaultPermissionModel: newPermissionModel(conn),
	}
}

func (m *customPermissionModel) withSession(session sqlx.Session) PermissionModel {
	return NewPermissionModel(sqlx.NewSqlConnFromSession(session))
}

// FindAll 查询所有权限
func (m *customPermissionModel) FindAll(ctx context.Context) ([]*Permission, error) {
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY `id` ASC", permissionRows, m.table)
	var permissions []*Permission
	err := m.conn.QueryRowsCtx(ctx, &permissions, query)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// FindByModule 按模块查询权限
func (m *customPermissionModel) FindByModule(ctx context.Context, module string) ([]*Permission, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `module` = ? ORDER BY `id` ASC", permissionRows, m.table)
	var permissions []*Permission
	err := m.conn.QueryRowsCtx(ctx, &permissions, query, module)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}
