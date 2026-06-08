package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ HelpArticleModel = (*customHelpArticleModel)(nil)

type (
	// HelpArticleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHelpArticleModel.
	HelpArticleModel interface {
		helpArticleModel
		withSession(session sqlx.Session) HelpArticleModel
	}

	customHelpArticleModel struct {
		*defaultHelpArticleModel
	}
)

// NewHelpArticleModel returns a model for the database table.
func NewHelpArticleModel(conn sqlx.SqlConn) HelpArticleModel {
	return &customHelpArticleModel{
		defaultHelpArticleModel: newHelpArticleModel(conn),
	}
}

func (m *customHelpArticleModel) withSession(session sqlx.Session) HelpArticleModel {
	return NewHelpArticleModel(sqlx.NewSqlConnFromSession(session))
}
