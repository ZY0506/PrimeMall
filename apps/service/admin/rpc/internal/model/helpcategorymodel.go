package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ HelpCategoryModel = (*customHelpCategoryModel)(nil)

type (
	// HelpCategoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHelpCategoryModel.
	HelpCategoryModel interface {
		helpCategoryModel
		withSession(session sqlx.Session) HelpCategoryModel
	}

	customHelpCategoryModel struct {
		*defaultHelpCategoryModel
	}
)

// NewHelpCategoryModel returns a model for the database table.
func NewHelpCategoryModel(conn sqlx.SqlConn) HelpCategoryModel {
	return &customHelpCategoryModel{
		defaultHelpCategoryModel: newHelpCategoryModel(conn),
	}
}

func (m *customHelpCategoryModel) withSession(session sqlx.Session) HelpCategoryModel {
	return NewHelpCategoryModel(sqlx.NewSqlConnFromSession(session))
}
