package model

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ OrderInfoModel = (*customOrderInfoModel)(nil)

type (
	// OrderInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customOrderInfoModel.
	OrderInfoModel interface {
		orderInfoModel
		InsertTx(ctx context.Context, session sqlx.Session, data *OrderInfo) (sql.Result, error)
		UpdateTx(ctx context.Context, session sqlx.Session, newData *OrderInfo) error
		FindList(ctx context.Context, userId uint64, status, page, pageSize int64) ([]*OrderInfo, int64, error)
		withSession(session sqlx.Session) OrderInfoModel
	}

	customOrderInfoModel struct {
		*defaultOrderInfoModel
	}
)

// NewOrderInfoModel returns a model for the database table.
func NewOrderInfoModel(conn sqlx.SqlConn) OrderInfoModel {
	return &customOrderInfoModel{
		defaultOrderInfoModel: newOrderInfoModel(conn),
	}
}

func (m *customOrderInfoModel) withSession(session sqlx.Session) OrderInfoModel {
	return NewOrderInfoModel(sqlx.NewSqlConnFromSession(session))
}

// InsertTx 事务操作插入
func (m *customOrderInfoModel) InsertTx(ctx context.Context, session sqlx.Session, data *OrderInfo) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, orderInfoRowsExpectAutoSet)
	ret, err := session.ExecCtx(ctx, query, data.Id, data.OrderSn, data.OrderType, data.IdempotencyKey, data.UserId, data.Status, data.AddressSnap, data.TotalAmount, data.FreightAmount, data.CouponId, data.CouponDiscount, data.PayAmount, data.Remark, data.PayTime, data.PayType, data.DeliveryTime, data.DeliverySn, data.DeliveryCorp, data.ReceiveTime, data.CancelTime, data.CancelReasonType, data.CancelReason, data.SeckillActivityId, data.SeckillPrice, data.ExpireTime)
	return ret, err
}

// UpdateTx 事务操作更新
func (m *customOrderInfoModel) UpdateTx(ctx context.Context, session sqlx.Session, newData *OrderInfo) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, orderInfoRowsWithPlaceHolder)
	_, err := session.ExecCtx(ctx, query, newData.OrderSn, newData.OrderType, newData.IdempotencyKey, newData.UserId, newData.Status, newData.AddressSnap, newData.TotalAmount, newData.FreightAmount, newData.CouponId, newData.CouponDiscount, newData.PayAmount, newData.Remark, newData.PayTime, newData.PayType, newData.DeliveryTime, newData.DeliverySn, newData.DeliveryCorp, newData.ReceiveTime, newData.CancelTime, newData.CancelReasonType, newData.CancelReason, newData.SeckillActivityId, newData.SeckillPrice, newData.ExpireTime, newData.Id)
	return err
}

// FindList 查询列表
func (m *customOrderInfoModel) FindList(ctx context.Context, userId uint64, status, page, pageSize int64) ([]*OrderInfo, int64, error) {
	if page <= 0 || pageSize <= 0 {
		return nil, 0, nil
	}

	var total int64
	var countQuery string
	var queryArgs []interface{}

	// 构建计数查询
	if status == 0 {
		countQuery = fmt.Sprintf("select count(*) from %s where `user_id` = ?", m.table)
		queryArgs = append(queryArgs, userId)
	} else {
		countQuery = fmt.Sprintf("select count(*) from %s where `user_id` = ? and `status` = ?", m.table)
		queryArgs = append(queryArgs, userId, status)
	}

	err := m.conn.QueryRowCtx(ctx, &total, countQuery, queryArgs...)
	if err != nil {
		return []*OrderInfo{}, 0, err
	}

	// 构建列表查询
	offset := (page - 1) * pageSize
	var listQuery string
	var listArgs []interface{}

	if status == 0 {
		listQuery = fmt.Sprintf("select %s from %s where `user_id` = ? order by created_at desc limit ?, ?", orderInfoRows, m.table)
		listArgs = append(listArgs, userId, offset, pageSize)
	} else {
		listQuery = fmt.Sprintf("select %s from %s where `user_id` = ? and `status` = ? order by created_at desc limit ?, ?", orderInfoRows, m.table)
		listArgs = append(listArgs, userId, status, offset, pageSize)
	}

	var resp []*OrderInfo
	err = m.conn.QueryRowsCtx(ctx, &resp, listQuery, listArgs...)
	return resp, total, err
}
