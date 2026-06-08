package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ PaymentCallbackLogModel = (*customPaymentCallbackLogModel)(nil)

type (
	// PaymentCallbackLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPaymentCallbackLogModel.
	PaymentCallbackLogModel interface {
		paymentCallbackLogModel
		withSession(session sqlx.Session) PaymentCallbackLogModel
	}

	customPaymentCallbackLogModel struct {
		*defaultPaymentCallbackLogModel
	}
)

// NewPaymentCallbackLogModel returns a model for the database table.
func NewPaymentCallbackLogModel(conn sqlx.SqlConn) PaymentCallbackLogModel {
	return &customPaymentCallbackLogModel{
		defaultPaymentCallbackLogModel: newPaymentCallbackLogModel(conn),
	}
}

func (m *customPaymentCallbackLogModel) withSession(session sqlx.Session) PaymentCallbackLogModel {
	return NewPaymentCallbackLogModel(sqlx.NewSqlConnFromSession(session))
}
