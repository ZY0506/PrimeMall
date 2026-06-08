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
	hotKeywordFieldNames          = builder.RawFieldNames(&HotKeyword{})
	hotKeywordRows                = strings.Join(hotKeywordFieldNames, ",")
	hotKeywordRowsExpectAutoSet   = strings.Join(stringx.Remove(hotKeywordFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`update_at`", "`update_time`", "`updated_at`"), ",")
	hotKeywordRowsWithPlaceHolder = strings.Join(stringx.Remove(hotKeywordFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`update_at`", "`update_time`", "`updated_at`"), "=?,") + "=?"
)

type (
	hotKeywordModel interface {
		Insert(ctx context.Context, data *HotKeyword) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*HotKeyword, error)
		Update(ctx context.Context, data *HotKeyword) error
		Delete(ctx context.Context, id uint64) error
	}

	defaultHotKeywordModel struct {
		conn  sqlx.SqlConn
		table string
	}

	HotKeyword struct {
		Id          uint64    `db:"id"`
		Keyword     string    `db:"keyword"`
		SearchCount int64     `db:"search_count"`
		Sort        int64     `db:"sort"`
		Status      int64     `db:"status"`
		UpdatedAt   time.Time `db:"updated_at"`
	}
)

func newHotKeywordModel(conn sqlx.SqlConn) *defaultHotKeywordModel {
	return &defaultHotKeywordModel{
		conn:  conn,
		table: "hot_keyword",
	}
}

func (m *defaultHotKeywordModel) Insert(ctx context.Context, data *HotKeyword) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?)", m.table, hotKeywordRowsExpectAutoSet)
	ret, err := m.conn.ExecCtx(ctx, query, data.Keyword, data.SearchCount, data.Sort, data.Status)
	return ret, err
}

func (m *defaultHotKeywordModel) FindOne(ctx context.Context, id uint64) (*HotKeyword, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", hotKeywordRows, m.table)
	var resp HotKeyword
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

func (m *defaultHotKeywordModel) Update(ctx context.Context, data *HotKeyword) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, hotKeywordRowsWithPlaceHolder)
	_, err := m.conn.ExecCtx(ctx, query, data.Keyword, data.SearchCount, data.Sort, data.Status, data.Id)
	return err
}

func (m *defaultHotKeywordModel) Delete(ctx context.Context, id uint64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}
