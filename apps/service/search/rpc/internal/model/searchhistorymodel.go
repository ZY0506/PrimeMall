package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SearchHistoryModel = (*customSearchHistoryModel)(nil)

type (
	SearchHistoryModel interface {
		searchHistoryModel
		customSearchHistory
	}

	customSearchHistoryModel struct {
		*defaultSearchHistoryModel
	}

	customSearchHistory interface {
		FindByUserID(ctx context.Context, userID uint64) ([]*SearchHistory, error)
		DeleteByUserID(ctx context.Context, userID uint64) error
		RecordHistory(ctx context.Context, userID uint64, keyword string, resultCount int64) error
	}
)

func NewSearchHistoryModel(conn sqlx.SqlConn) SearchHistoryModel {
	return &customSearchHistoryModel{
		defaultSearchHistoryModel: newSearchHistoryModel(conn),
	}
}

func (m *customSearchHistoryModel) FindByUserID(ctx context.Context, userID uint64) ([]*SearchHistory, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = ? ORDER BY created_at DESC LIMIT 20", searchHistoryRows, m.table)
	var items []*SearchHistory
	if err := m.conn.QueryRowsCtx(ctx, &items, query, userID); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *customSearchHistoryModel) DeleteByUserID(ctx context.Context, userID uint64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE user_id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, userID)
	return err
}

func (m *customSearchHistoryModel) RecordHistory(ctx context.Context, userID uint64, keyword string, resultCount int64) error {
	data := &SearchHistory{
		UserId:      userID,
		Keyword:     keyword,
		ResultCount: resultCount,
		CreatedAt:   time.Now(),
	}
	_, err := m.Insert(ctx, data)
	return err
}
