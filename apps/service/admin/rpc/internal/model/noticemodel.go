package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ NoticeModel = (*customNoticeModel)(nil)

type (
	// NoticeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customNoticeModel.
	NoticeModel interface {
		noticeModel
		withSession(session sqlx.Session) NoticeModel
	}

	customNoticeModel struct {
		*defaultNoticeModel
	}
)

// NewNoticeModel returns a model for the database table.
func NewNoticeModel(conn sqlx.SqlConn) NoticeModel {
	return &customNoticeModel{
		defaultNoticeModel: newNoticeModel(conn),
	}
}

func (m *customNoticeModel) withSession(session sqlx.Session) NoticeModel {
	return NewNoticeModel(sqlx.NewSqlConnFromSession(session))
}
