package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RolePermissionModel = (*customRolePermissionModel)(nil)

type (
	RolePermissionModel interface {
		rolePermissionModel
		FindByRoleId(ctx context.Context, roleId uint64) ([]*RolePermission, error)
		DeleteByRoleId(ctx context.Context, roleId uint64) error
		InsertTx(ctx context.Context, session sqlx.Session, data *RolePermission) (sql.Result, error)
		DeleteByRoleIdTx(ctx context.Context, session sqlx.Session, roleId uint64) error
		withSession(session sqlx.Session) RolePermissionModel
	}

	customRolePermissionModel struct {
		*defaultRolePermissionModel
	}
)

// NewRolePermissionModel returns a model for the database table.
func NewRolePermissionModel(conn sqlx.SqlConn) RolePermissionModel {
	return &customRolePermissionModel{
		defaultRolePermissionModel: newRolePermissionModel(conn),
	}
}

func (m *customRolePermissionModel) withSession(session sqlx.Session) RolePermissionModel {
	return NewRolePermissionModel(sqlx.NewSqlConnFromSession(session))
}

// FindByRoleId 根据角色ID查询所有权限关联
func (m *customRolePermissionModel) FindByRoleId(ctx context.Context, roleId uint64) ([]*RolePermission, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `role_id` = ?", rolePermissionRows, m.table)
	var rps []*RolePermission
	err := m.conn.QueryRowsCtx(ctx, &rps, query, roleId)
	if err != nil {
		return nil, err
	}
	return rps, nil
}

// DeleteByRoleId 根据角色ID删除所有权限关联
func (m *customRolePermissionModel) DeleteByRoleId(ctx context.Context, roleId uint64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE `role_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, roleId)
	return err
}

// InsertTx 事务插入角色权限
func (m *customRolePermissionModel) InsertTx(ctx context.Context, session sqlx.Session, data *RolePermission) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?)", m.table, rolePermissionRowsExpectAutoSet)
	ret, err := session.ExecCtx(ctx, query, data.RoleId, data.PermissionId)
	return ret, err
}

// DeleteByRoleIdTx 事务中根据角色ID删除所有权限关联
func (m *customRolePermissionModel) DeleteByRoleIdTx(ctx context.Context, session sqlx.Session, roleId uint64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE `role_id` = ?", m.table)
	_, err := session.ExecCtx(ctx, query, roleId)
	return err
}
