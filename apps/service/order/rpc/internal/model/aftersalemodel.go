package model

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
)

var _ AfterSaleModel = (*customAfterSaleModel)(nil)

type (
	// AfterSaleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAfterSaleModel.
	AfterSaleModel interface {
		afterSaleModel
		InsertTx(ctx context.Context, session sqlx.Session, data *AfterSale) (sql.Result, error)
		UpdateTx(ctx context.Context, session sqlx.Session, newData *AfterSale) error
		FindOneByOrderSnSkuId(ctx context.Context, orderSn string, skuId uint64, unfinishedStatus []int64) (*AfterSale, error)
		FindPageByUserId(ctx context.Context, userId uint64, status, offset, size int64) ([]*AfterSale, int64, error)
		AdminFindPage(ctx context.Context, status, offset, pageSize int64) ([]*AfterSale, int64, error)
		withSession(session sqlx.Session) AfterSaleModel
	}

	customAfterSaleModel struct {
		*defaultAfterSaleModel
	}
)

// NewAfterSaleModel returns a model for the database table.
func NewAfterSaleModel(conn sqlx.SqlConn) AfterSaleModel {
	return &customAfterSaleModel{
		defaultAfterSaleModel: newAfterSaleModel(conn),
	}
}

func (m *customAfterSaleModel) withSession(session sqlx.Session) AfterSaleModel {
	return NewAfterSaleModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customAfterSaleModel) InsertTx(ctx context.Context, session sqlx.Session, data *AfterSale) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, afterSaleRowsExpectAutoSet)
	ret, err := session.ExecCtx(ctx, query, data.AfterSaleSn, data.IdempotencyKey, data.OrderId, data.OrderSn, data.UserId, data.Type, data.Status, data.RefundStatus, data.ApplyAmount, data.ApprovedAmount, data.RealRefundAmount, data.Reason, data.AuditRemark, data.Images, data.ReturnTrackingSn, data.ReturnTrackingCorp, data.ReturnReceivedTime, data.RefundSn, data.RefundTime, data.ThirdRefundSn, data.CancelTime)
	return ret, err
}

func (m *customAfterSaleModel) UpdateTx(ctx context.Context, session sqlx.Session, newData *AfterSale) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, afterSaleRowsWithPlaceHolder)
	_, err := session.ExecCtx(ctx, query, newData.AfterSaleSn, newData.IdempotencyKey, newData.OrderId, newData.OrderSn, newData.UserId, newData.Type, newData.Status, newData.RefundStatus, newData.ApplyAmount, newData.ApprovedAmount, newData.RealRefundAmount, newData.Reason, newData.AuditRemark, newData.Images, newData.ReturnTrackingSn, newData.ReturnTrackingCorp, newData.ReturnReceivedTime, newData.RefundSn, newData.RefundTime, newData.ThirdRefundSn, newData.CancelTime, newData.Id)
	return err
}

func (m *customAfterSaleModel) FindOneByOrderSnSkuId(ctx context.Context, orderSn string, skuId uint64, unfinishedStatus []int64) (*AfterSale, error) {
	// 动态生成 IN 子句的占位符 (?, ?, ...)
	placeholders := make([]string, len(unfinishedStatus))
	for i := range unfinishedStatus {
		placeholders[i] = "?"
	}
	inClause := strings.Join(placeholders, ", ")

	query := fmt.Sprintf("select %s from %s where order_sn = ? and sku_id = ? and status in (%s) limit 1", afterSaleRows, m.table, inClause)

	// 构建参数列表：orderSn, skuId, 然后是 unfinishedStatus 的所有元素
	args := make([]interface{}, 0, 2+len(unfinishedStatus))
	args = append(args, orderSn, skuId)
	for _, s := range unfinishedStatus {
		args = append(args, s)
	}

	var resp AfterSale
	err := m.conn.QueryRowCtx(ctx, &resp, query, args...)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customAfterSaleModel) FindPageByUserId(ctx context.Context, userId uint64, status, offset, size int64) ([]*AfterSale, int64, error) {
	where := "user_id = ?"
	args := []interface{}{userId}
	if status != 0 {
		where += " and status = ?"
		args = append(args, status)
	}

	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*AfterSale{}, 0, nil
	}

	listArgs := append(args, offset, size)
	query := fmt.Sprintf("select %s from %s where %s order by created_at desc limit ?, ?", afterSaleRows, m.table, where)
	var resp []*AfterSale
	if err := m.conn.QueryRowsCtx(ctx, &resp, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return resp, total, nil
}

func (m *customAfterSaleModel) AdminFindPage(ctx context.Context, status, offset, pageSize int64) ([]*AfterSale, int64, error) {
	if pageSize <= 0 {
		return []*AfterSale{}, 0, nil
	}

	var total int64
	var countQuery string
	var queryArgs []interface{}

	if status == 0 {
		countQuery = fmt.Sprintf("select count(*) from %s", m.table)
	} else {
		countQuery = fmt.Sprintf("select count(*) from %s where status = ?", m.table)
		queryArgs = append(queryArgs, status)
	}

	err := m.conn.QueryRowCtx(ctx, &total, countQuery, queryArgs...)
	if err != nil {
		return []*AfterSale{}, 0, err
	}

	if total == 0 {
		return []*AfterSale{}, 0, nil
	}

	offset = (offset - 1) * pageSize
	var listQuery string
	var listArgs []interface{}

	if status == 0 {
		listQuery = fmt.Sprintf("select %s from %s order by created_at desc limit ?, ?", afterSaleRows, m.table)
		listArgs = append(listArgs, offset, pageSize)
	} else {
		listQuery = fmt.Sprintf("select %s from %s where status = ? order by created_at desc limit ?, ?", afterSaleRows, m.table)
		listArgs = append(listArgs, status, offset, pageSize)
	}

	var resp []*AfterSale
	err = m.conn.QueryRowsCtx(ctx, &resp, listQuery, listArgs...)
	return resp, total, err
}
