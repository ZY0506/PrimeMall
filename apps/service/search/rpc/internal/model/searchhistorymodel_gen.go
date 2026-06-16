package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	searchHistoryFieldNames          = builder.RawFieldNames(&SearchHistory{})
	searchHistoryRows                = strings.Join(searchHistoryFieldNames, ",")
	searchHistoryRowsExpectAutoSet   = strings.Join(stringx.Remove(searchHistoryFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`update_at`", "`update_time`", "`updated_at`"), ",")
	searchHistoryRowsWithPlaceHolder = strings.Join(stringx.Remove(searchHistoryFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`update_at`", "`update_time`", "`updated_at`"), "=?,") + "=?"
)

type (
	searchHistoryModel interface {
		Insert(ctx context.Context, data *SearchHistory) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*SearchHistory, error)
		Update(ctx context.Context, data *SearchHistory) error
		Delete(ctx context.Context, id uint64) error
	}

	defaultSearchHistoryModel struct {
		conn  sqlx.SqlConn
		table string
	}

	SearchHistory struct {
		Id          uint64    `db:"id"`
		UserId      uint64    `db:"user_id"`
		Keyword     string    `db:"keyword"`
		ResultCount int64     `db:"result_count"`
		CreatedAt   time.Time `db:"created_at"`
	}
)

func newSearchHistoryModel(conn sqlx.SqlConn) *defaultSearchHistoryModel {
	return &defaultSearchHistoryModel{
		conn:  conn,
		table: "search_history",
	}
}

func (m *defaultSearchHistoryModel) Insert(ctx context.Context, data *SearchHistory) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?)", m.table, searchHistoryRowsExpectAutoSet)
	ret, err := m.conn.ExecCtx(ctx, query, data.UserId, data.Keyword, data.ResultCount)
	return ret, err
}

func (m *defaultSearchHistoryModel) FindOne(ctx context.Context, id uint64) (*SearchHistory, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", searchHistoryRows, m.table)
	var resp SearchHistory
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultSearchHistoryModel) Update(ctx context.Context, data *SearchHistory) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, searchHistoryRowsWithPlaceHolder)
	_, err := m.conn.ExecCtx(ctx, query, data.UserId, data.Keyword, data.ResultCount, data.Id)
	return err
}

func (m *defaultSearchHistoryModel) Delete(ctx context.Context, id uint64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}
