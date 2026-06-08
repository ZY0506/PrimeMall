package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ FreightTemplateModel = (*customFreightTemplateModel)(nil)

type (
	// FreightTemplateModel is an interface to be customized, add more methods here,
	// and implement the added methods in customFreightTemplateModel.
	FreightTemplateModel interface {
		freightTemplateModel
		customFreightTemplate
		withSession(session sqlx.Session) FreightTemplateModel
	}

	customFreightTemplateModel struct {
		*defaultFreightTemplateModel
	}

	customFreightTemplate interface {
		FindDefault(ctx context.Context) (*FreightTemplate, error)
		FindList(ctx context.Context, page, pageSize int64) ([]*FreightTemplate, int64, error)
	}
)

// NewFreightTemplateModel returns a model for the database table.
func NewFreightTemplateModel(conn sqlx.SqlConn) FreightTemplateModel {
	return &customFreightTemplateModel{
		defaultFreightTemplateModel: newFreightTemplateModel(conn),
	}
}

func (m *customFreightTemplateModel) withSession(session sqlx.Session) FreightTemplateModel {
	return NewFreightTemplateModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customFreightTemplateModel) FindDefault(ctx context.Context) (*FreightTemplate, error) {
	query := fmt.Sprintf("select %s from %s where `is_default` = 1 limit 1", freightTemplateRows, m.table)
	var resp FreightTemplate
	err := m.conn.QueryRowCtx(ctx, &resp, query)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customFreightTemplateModel) FindList(ctx context.Context, page, pageSize int64) ([]*FreightTemplate, int64, error) {
	if page <= 0 || pageSize <= 0 {
		return nil, 0, nil
	}

	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where `delete_at` IS NULL", m.table)
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*FreightTemplate{}, 0, nil
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where `delete_at` IS NULL order by created_at desc limit ?, ?", freightTemplateRows, m.table)
	var resp []*FreightTemplate
	err = m.conn.QueryRowsCtx(ctx, &resp, query, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	return resp, total, nil
}
