package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HotKeywordModel = (*customHotKeywordModel)(nil)

type (
	HotKeywordModel interface {
		hotKeywordModel
		customHotKeyword
	}

	customHotKeywordModel struct {
		*defaultHotKeywordModel
	}

	customHotKeyword interface {
		FindActiveList(ctx context.Context) ([]*HotKeyword, error)
		IncrSearchCount(ctx context.Context, keyword string) error
	}
)

func NewHotKeywordModel(conn sqlx.SqlConn) HotKeywordModel {
	return &customHotKeywordModel{
		defaultHotKeywordModel: newHotKeywordModel(conn),
	}
}

func (m *customHotKeywordModel) FindActiveList(ctx context.Context) ([]*HotKeyword, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE status = 1 ORDER BY sort ASC, search_count DESC LIMIT 20", hotKeywordRows, m.table)
	var items []*HotKeyword
	if err := m.conn.QueryRowsCtx(ctx, &items, query); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *customHotKeywordModel) IncrSearchCount(ctx context.Context, keyword string) error {
	query := fmt.Sprintf("INSERT INTO %s (keyword, search_count, sort, status) VALUES (?, 1, 0, 1) ON DUPLICATE KEY UPDATE search_count = search_count + 1", m.table)
	_, err := m.conn.ExecCtx(ctx, query, keyword)
	return err
}
