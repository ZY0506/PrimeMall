package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
)

var _ CategoryModel = (*customCategoryModel)(nil)

// 生成带表别名 c. 的字段列表：c.`id`,c.`parent_id`,c.`name`...
var categoryRowsAliased = strings.Join(categoryFieldNames, ",c.")

type (
	// CategoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCategoryModel.
	CategoryModel interface {
		categoryModel
		customCategory
		withSession(session sqlx.Session) CategoryModel
	}

	customCategoryModel struct {
		*defaultCategoryModel
	}
	customCategory interface {
		FindAllChildren(ctx context.Context, id uint64) (*[]Category, error)
		FindALL(ctx context.Context) (*[]Category, error)
	}
)

// NewCategoryModel returns a model for the database table.
func NewCategoryModel(conn sqlx.SqlConn) CategoryModel {
	return &customCategoryModel{
		defaultCategoryModel: newCategoryModel(conn),
	}
}

func (m *customCategoryModel) withSession(session sqlx.Session) CategoryModel {
	return NewCategoryModel(sqlx.NewSqlConnFromSession(session))
}

func (m *defaultCategoryModel) FindAllChildren(ctx context.Context, id uint64) (*[]Category, error) {
	// 递归 CTE：查询所有后代（不含自身，若需要自身可去掉 id != ? 条件）
	query := fmt.Sprintf(`
        WITH RECURSIVE category_tree AS (
            SELECT %s FROM %s WHERE parent_id = ? AND status = 1
            UNION ALL
            SELECT c.%s FROM %s c
            INNER JOIN category_tree ct ON c.parent_id = ct.id
            WHERE c.status = 1
        )
        SELECT * FROM category_tree ORDER BY sort ASC, id ASC
    `, categoryRows, m.table, categoryRowsAliased, m.table)
	var resp []Category
	err := m.conn.QueryRowsCtx(ctx, &resp, query, id)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *defaultCategoryModel) FindALL(ctx context.Context) (*[]Category, error) {
	query := fmt.Sprintf("select %s from %s where `status` = 1 order by sort ASC,id ASC", categoryRows, m.table)
	var resp []Category
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
